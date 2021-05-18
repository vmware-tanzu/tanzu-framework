// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type scaleClusterOptions struct {
	workerCount       int32
	controlPlaneCount int32
	namespace         string
}

var sc = &scaleClusterOptions{}

var scaleClusterCmd = &cobra.Command{
	Use:   "cluster CLUSTER_NAME",
	Short: "Scale a Tanzu Kubernetes Grid cluster",
	Long:  "Scale a Tanzu Kubernetes Grid cluster",
	Example: Examples(`
		# Scales the number of worker nodes of cluster named 'my-cluster' to 5
		tkg scale cluster my-cluster -w 5`),
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runScaleCluster(args[0])
		verifyCommandError(err)
	},
}

func init() {
	scaleClusterCmd.Flags().Int32VarP(&sc.workerCount, "worker-machine-count", "w", 0, "Number of worker nodes to scale to")
	scaleClusterCmd.Flags().Int32VarP(&sc.controlPlaneCount, "controlplane-machine-count", "c", 0, "Number of control plane nodes to scale to")
	scaleClusterCmd.Flags().StringVarP(&sc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")

	scaleCmd.AddCommand(scaleClusterCmd)
}

func runScaleCluster(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.ScaleClusterOptions{
		ClusterName:       clusterName,
		ControlPlaneCount: sc.controlPlaneCount,
		WorkerCount:       sc.workerCount,
		Namespace:         sc.namespace,
	}
	return tkgClient.ScaleCluster(options)
}
