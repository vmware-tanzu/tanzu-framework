// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	tkgclient "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
	tkgutils "github.com/vmware-tanzu/tanzu-framework/tkg/utils"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	tkgauth "github.com/vmware-tanzu/tanzu-framework/tkg/auth"
)

type getClusterKubeconfigOptions struct {
	adminKubeconfig bool
	exportFile      string
}

var getKCOptions = &getClusterKubeconfigOptions{}

// mgmt.1
// the command to add to generate the kubeconfig for the mgmt cluster
// this will delegate to getKubecofig, which will call the function
// chain below.
var getClusterKubeconfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Kubeconfig of a management cluster",
	Long:  `Get Kubeconfig of a management cluster and merge the context into the default kubeconfig file`,
	Example: `
	# Get management cluster kubeconfig
	tanzu management-cluster kubeconfig get
	
	# Get management cluster admin kubeconfig
	tanzu management-cluster kubeconfig get --admin`,
	RunE:         getKubeconfig,
	SilenceUsage: true,
}

func init() {
	getClusterKubeconfigCmd.Flags().BoolVarP(&getKCOptions.adminKubeconfig, "admin", "", false, "Get admin kubeconfig of the management cluster")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for management cluster")

	clusterKubeconfigCmd.AddCommand(getClusterKubeconfigCmd)
}

// mgmt.2
// the entrypoint defined by getClusterKubeconfigCmd (a Cobra cmd)
func getKubeconfig(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("get management cluster kubeconfig with a global server is not implemented yet")
	}
	return getClusterKubeconfig(server)
}

// mgmt.3
func getClusterKubeconfig(server *configapi.Server) error {
	forceUpdateTKGCompatibilityImage := false
	tkgctlClient, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	if err != nil {
		return err
	}

	mcClustername, err := tkgutils.GetClusterNameFromKubeconfigAndContext(server.ManagementClusterOpts.Path,
		server.ManagementClusterOpts.Context)
	if err != nil {
		return errors.Wrap(err, "failed to get management cluster name from kubeconfig")
	}

	if getKCOptions.adminKubeconfig {
		// this is the non-pinniped path
		return getAdminKubeconfig(tkgctlClient, mcClustername)
	}
	// generate the cluster specific kubeconfig file.
	// mgmt.4
	return getPinnipedKubeconfig(tkgctlClient, mcClustername)
}

func getAdminKubeconfig(tkgctlClient tkgctl.TKGClient, mcClustername string) error {
	getClusterCredentialsOptions := tkgctl.GetWorkloadClusterCredentialsOptions{
		ClusterName: mcClustername,
		Namespace:   TKGSystemNamespace,
		ExportFile:  getKCOptions.exportFile,
	}
	return tkgctlClient.GetCredentials(getClusterCredentialsOptions)
}

// mgmt.5
// this function delegates to the tkgauth.GetPinnipedKubeconfig() to generate the kubeconfig file itself.
func getPinnipedKubeconfig(tkgctlClient tkgctl.TKGClient, mcClustername string) error {
	getClusterPinnipedInfoOptions := tkgctl.GetClusterPinnipedInfoOptions{
		ClusterName:         mcClustername,
		Namespace:           TKGSystemNamespace,
		IsManagementCluster: true,
	}

	clusterPinnipedInfo, err := tkgctlClient.GetClusterPinnipedInfo(getClusterPinnipedInfoOptions)
	if err != nil {
		return err
	}

	// for management cluster the audience would be set to IssuerURL
	audience := clusterPinnipedInfo.PinnipedInfo.Data.Issuer

	pinnipedSupervisorDiscoveryOpts := tkgctl.GetClusterPinnipedSupervisorDiscoveryOptions{
		Endpoint: fmt.Sprintf("%s/.well-known/openid-configuration", clusterPinnipedInfo.PinnipedInfo.Data.Issuer),
		CABundle: clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle,
	}
	supervisorDiscoveryInfo, err := tkgctlClient.GetPinnipedSupervisorDiscovery(pinnipedSupervisorDiscoveryOpts)
	if err != nil {
		return err
	}

	// TODO(BEN): remove this, we don't need it once done
	fmt.Printf("🦄 this is the response from the well-known endpoint: \n%+v\n", supervisorDiscoveryInfo)
	log.Infof("🦄 this is the response from the well-known endpoint: \n%+v\n", supervisorDiscoveryInfo)

	// mgmt.6
	// this seems a pretty reasonable entrypoint to pursue, however is it
	// "getting" an existing kubeconfig, or is it generating a new kubeconfig?
	kubeconfig, _ := tkgauth.GetPinnipedKubeconfig(
		clusterPinnipedInfo.ClusterInfo,
		clusterPinnipedInfo.PinnipedInfo,
		clusterPinnipedInfo.ClusterName,
		audience,
		supervisorDiscoveryInfo)

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
