// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type scaleClustersOptions struct {
	namespace         string
	nodePoolName      string
	workerCount       int32
	controlPlaneCount int32
}

var sc = &scaleClustersOptions{}

var scaleClusterCmd = &cobra.Command{
	Use:          "scale CLUSTER_NAME",
	Short:        "Scale a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         scale,
	SilenceUsage: true,
}

func init() {
	scaleClusterCmd.Flags().Int32VarP(&sc.workerCount, "worker-machine-count", "w", 0, "The number of worker nodes to scale to. Assumes unchanged if not specified")
	scaleClusterCmd.Flags().Int32VarP(&sc.controlPlaneCount, "controlplane-machine-count", "c", 0, "The number of control plane nodes to scale to. Assumes unchanged if not specified")
	scaleClusterCmd.Flags().StringVarP(&sc.nodePoolName, "node-pool-name", "p", "", "The name of the node-pool to scale")
	scaleClusterCmd.Flags().StringVarP(&sc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
}

func scale(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("scaling cluster with a global server is not implemented yet")
	}
	return scaleCluster(server, args[0])
}

func scaleCluster(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	scaleClusterOptions := tkgctl.ScaleClusterOptions{
		ClusterName:       clusterName,
		ControlPlaneCount: sc.controlPlaneCount,
		WorkerCount:       sc.workerCount,
		Namespace:         sc.namespace,
		NodePoolName:      sc.nodePoolName,
	}

	return tkgctlClient.ScaleCluster(scaleClusterOptions)
}
