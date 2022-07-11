// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

type deleteNodePoolOptions struct {
	nodePoolName string
	namespace    string
}

var deleteNP = &deleteNodePoolOptions{}

var deleteNodePoolCmd = &cobra.Command{
	Use:          "delete CLUSTER_NAME",
	Short:        "Delete a NodePool object of a cluster",
	Long:         "Delete a NodePool object of a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         deleteNodePool,
	SilenceUsage: true,
}

func init() {
	deleteNodePoolCmd.Flags().StringVarP(&deleteNP.nodePoolName, "name", "n", "", "Name of the NodePool object")
	_ = deleteNodePoolCmd.MarkFlagRequired("name")
	deleteNodePoolCmd.Flags().StringVar(&deleteNP.namespace, "namespace", "", "The namespace where the NodePool object was created, default to the cluster's namespace")
	clusterNodePoolCmd.AddCommand(deleteNodePoolCmd)
}

func deleteNodePool(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("deleting node pools with a global server is not implemented yet")
	}
	return runDeleteNodePool(server, args[0])
}

func runDeleteNodePool(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	options := client.DeleteMachineDeploymentOptions{
		ClusterName: clusterName,
		Namespace:   deleteNP.namespace,
		Name:        deleteNP.nodePoolName,
	}

	err = tkgctlClient.DeleteMachineDeployment(options)
	if err == nil {
		log.Infof("Node pool '%s' is being deleted", options.Name)
	}

	return err
}
