// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type deleteClustersOptions struct {
	namespace  string
	unattended bool
}

var dc = &deleteClustersOptions{}

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster CLUSTER_NAME",
	Short: "Delete a Tanzu Kubernetes cluster",
	Long:  `Use the management cluster to delete a Tanzu Kubernetes Cluster`,
	Example: Examples(`
		# Delete a workload cluster
		tkg delete cluster my-cluster

		# Delete a workload cluster in particular namespace
		tkg delete cluster my-cluster --namespace=dev-system`),
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runDeleteCluster(args[0])
		verifyCommandError(err)
	},
}

func init() {
	deleteClusterCmd.Flags().StringVarP(&dc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	deleteClusterCmd.Flags().BoolVarP(&dc.unattended, "yes", "y", false, "Delete workload cluster without asking for confirmation")

	deleteCmd.AddCommand(deleteClusterCmd)
}

func runDeleteCluster(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.DeleteClustersOptions{
		ClusterName: clusterName,
		Namespace:   dc.namespace,
		SkipPrompt:  dc.unattended || skipPrompt,
	}
	return tkgClient.DeleteCluster(options)
}
