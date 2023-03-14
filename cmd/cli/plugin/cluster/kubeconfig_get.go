// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	tkgauth "github.com/vmware-tanzu/tanzu-framework/tkg/auth"

	tkgclient "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

// this is a collection of the options that are actually passed as flags to the command itself
// but will be captured in this struct.  See the init() function
type getClusterKubeconfigOptions struct {
	namespace       string
	exportFile      string
	adminKubeconfig bool
}

// instantiate a struct var
var getKCOptions = &getClusterKubeconfigOptions{}

// TODO (BEN): ensure workload cluster code path works
// work.1
// create the Cobra command, and delegate to the function that
// will be responsible for generating the kubeconfig itself.
var getClusterKubeconfigCmd = &cobra.Command{
	Use:   "get CLUSTER_NAME",
	Short: "Get kubeconfig of a cluster",
	Long:  `Get kubeconfig of a cluster and merge the context into the default kubeconfig file`,
	Example: `
    # Get workload cluster kubeconfig
    tanzu cluster kubeconfig get CLUSTER_NAME

    # Get workload cluster admin kubeconfig
    tanzu cluster kubeconfig get CLUSTER_NAME --admin`,
	Args: cobra.ExactArgs(1),
	// the main function to run when teh command is invoked.
	RunE:         getKubeconfig,
	SilenceUsage: true,
}

// simply initialize the variables and capture whatever the user passes to them.
func init() {
	getClusterKubeconfigCmd.Flags().BoolVarP(&getKCOptions.adminKubeconfig, "admin", "", false, "Get admin kubeconfig of the workload cluster")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for workload cluster")

	clusterKubeconfigCmd.AddCommand(getClusterKubeconfigCmd)
}

// work.2
func getKubeconfig(cmd *cobra.Command, args []string) error {
	// not a flag, this is the first (and perhaps only?) arg to the command.
	workloadClusterName := args[0]
	// not sure what this does exactly yet.
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("get cluster kubeconfig with a global server is not implemented yet")
	}
	// delegates out
	return getClusterKubeconfig(server, workloadClusterName)
}

// work.4
// get the actual kubeconfig
// two options, admin or pinniped version.
func getClusterKubeconfig(server *configapi.Server, workloadClusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	if getKCOptions.adminKubeconfig {
		return getAdminKubeconfig(tkgctlClient, workloadClusterName)
	}
	// work.5->
	return getPinnipedKubeconfig(tkgctlClient, workloadClusterName)
}

func getAdminKubeconfig(tkgctlClient tkgctl.TKGClient, workloadClusterName string) error {
	getClusterCredentialsOptions := tkgctl.GetWorkloadClusterCredentialsOptions{
		ClusterName: workloadClusterName,
		Namespace:   getKCOptions.namespace,
		ExportFile:  getKCOptions.exportFile,
	}
	return tkgctlClient.GetCredentials(getClusterCredentialsOptions)
}

// work.5
func getPinnipedKubeconfig(tkgctlClient tkgctl.TKGClient, workloadClusterName string) error {
	getClusterPinnipedInfoOptions := tkgctl.GetClusterPinnipedInfoOptions{
		ClusterName:         workloadClusterName,
		Namespace:           getKCOptions.namespace,
		IsManagementCluster: false,
	}

	clusterPinnipedInfo, err := tkgctlClient.GetClusterPinnipedInfo(getClusterPinnipedInfoOptions)
	if err != nil {
		return err
	}

	// For (legacy) workload clusters, the audience is just <cluster-name>
	audience := clusterPinnipedInfo.ClusterName
	if clusterPinnipedInfo.ClusterAudience != nil {
		audience = *clusterPinnipedInfo.ClusterAudience
	}

	pinnipedSupervisorDiscoveryOpts := tkgctl.GetClusterPinnipedSupervisorDiscoveryOptions{
		Endpoint: fmt.Sprintf("%s/.well-known/openid-configuration", clusterPinnipedInfo.PinnipedInfo.Data.Issuer),
		CABundle: clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle,
	}
	// work.6->
	supervisorDiscoveryInfo, err := tkgctlClient.GetPinnipedSupervisorDiscovery(pinnipedSupervisorDiscoveryOpts)
	if err != nil {
		return err

	}

	// TODO(BEN): remove this, we don't need it once done
	//   alternatively do log.Debug() instead
	fmt.Printf("🦄 (workload cluster) this is the response from the well-known endpoint: \n%+v\n", supervisorDiscoveryInfo)
	log.Infof("🦄 (workload cluster) this is the response from the well-known endpoint: \n%+v\n", supervisorDiscoveryInfo)

	// work.7 finally generate the kubeconfig file
	// this seems pretty reasonable entrypoint again.
	// this is the same entrypoint as is called by the mgmt cluster plugin flow
	kubeconfig, err := tkgauth.GetPinnipedKubeconfig(
		clusterPinnipedInfo.ClusterInfo,
		clusterPinnipedInfo.PinnipedInfo,
		clusterPinnipedInfo.ClusterName,
		audience,
		supervisorDiscoveryInfo)

	if err != nil {
		return errors.Wrap(err, "unable to get kubeconfig")
	}

	kubeconfigbytes, err := json.Marshal(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "unable to marshall the kubeconfig")
	}
	err = tkgclient.MergeKubeConfigWithoutSwitchContext(kubeconfigbytes, getKCOptions.exportFile)
	if err != nil {
		return errors.Wrap(err, "unable to merge cluster kubeconfig into the current kubeconfig path")
	}

	if getKCOptions.exportFile != "" {
		log.Infof("You can now access the cluster by specifying '--kubeconfig %s' flag when using `kubectl` command \n", getKCOptions.exportFile)
	} else {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s'\n", kubeconfig.CurrentContext)
	}
	return nil
}
