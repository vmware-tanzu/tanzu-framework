// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	tkgauth "github.com/vmware-tanzu/tanzu-framework/tkg/auth"

	tkgclient "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type getClusterKubeconfigOptions struct {
	namespace       string
	exportFile      string
	adminKubeconfig bool
}

var getKCOptions = &getClusterKubeconfigOptions{}

var getClusterKubeconfigCmd = &cobra.Command{
	Use:   "get CLUSTER_NAME",
	Short: "Get kubeconfig of a cluster",
	Long:  `Get kubeconfig of a cluster and merge the context into the default kubeconfig file`,
	Example: `
    # Get workload cluster kubeconfig
    tanzu cluster kubeconfig get CLUSTER_NAME

    # Get workload cluster admin kubeconfig
    tanzu cluster kubeconfig get CLUSTER_NAME --admin`,
	Args:         cobra.ExactArgs(1),
	RunE:         getKubeconfig,
	SilenceUsage: true,
}

func init() {
	getClusterKubeconfigCmd.Flags().BoolVarP(&getKCOptions.adminKubeconfig, "admin", "", false, "Get admin kubeconfig of the workload cluster")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for workload cluster")

	clusterKubeconfigCmd.AddCommand(getClusterKubeconfigCmd)
}

func getKubeconfig(cmd *cobra.Command, args []string) error {
	workloadClusterName := args[0]

	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("get cluster kubeconfig with a global server is not implemented yet")
	}
	return getClusterKubeconfig(server, workloadClusterName)
}

func getClusterKubeconfig(server *configapi.Server, workloadClusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	if getKCOptions.adminKubeconfig {
		return getAdminKubeconfig(tkgctlClient, workloadClusterName)
	}
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

	kubeconfig, err := tkgauth.GetPinnipedKubeconfig(clusterPinnipedInfo.ClusterInfo, clusterPinnipedInfo.PinnipedInfo,
		clusterPinnipedInfo.ClusterName, audience)

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
