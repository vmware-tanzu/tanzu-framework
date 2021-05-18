// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type getCredentialsOptions struct {
	namespace  string
	exportFile string
}

var gc = &getCredentialsOptions{}

var getCredentialsCmd = &cobra.Command{
	Use:   "credentials CLUSTER_NAME",
	Short: "Get kubeconfig of a Tanzu Kubernetes cluster",
	Long:  `Get kubeconfig of a Tanzu Kubernetes cluster by name and merge the context into the default kubeconfig file`,
	Example: Examples(`
		# Get workload cluster kubeconfig
		tkg get credentials my-cluster`),
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.UnsetStdoutStderr()
		err := runGetCredentials(args[0])
		verifyCommandError(err)
	},
}

func init() {
	getCredentialsCmd.Flags().StringVarP(&gc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	getCredentialsCmd.Flags().StringVarP(&gc.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for workload cluster")

	getCmd.AddCommand(getCredentialsCmd)
}

func runGetCredentials(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.GetWorkloadClusterCredentialsOptions{
		ClusterName: clusterName,
		Namespace:   gc.namespace,
		ExportFile:  gc.exportFile,
	}
	return tkgClient.GetCredentials(options)
}
