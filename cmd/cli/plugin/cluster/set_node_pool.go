// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
)

type clusterSetNodePoolCmdOptions struct {
	FilePath  string
	Namespace string
}

var setNodePoolOptions clusterSetNodePoolCmdOptions

var clusterSetNodePoolCmd = &cobra.Command{
	Use:   "set CLUSTER_NAME",
	Short: "Set node pool for cluster",
	RunE:  runSetNodePool,
}

func init() {
	clusterSetNodePoolCmd.Flags().StringVarP(&setNodePoolOptions.FilePath, "file", "f", "", "The file describing the node pool (required)")
	clusterSetNodePoolCmd.Flags().StringVar(&setNodePoolOptions.Namespace, "namespace", "default", "The namespace the cluster is found in.")
	_ = clusterSetNodePoolCmd.MarkFlagRequired("file")
	clusterNodePoolCmd.AddCommand(clusterSetNodePoolCmd)
}

func runSetNodePool(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("setting machine healthcheck with a global server is not implemented yet")
	}
	return SetNodePool(server, args[0])
}

// SetNodePool creates or updates a node pool
func SetNodePool(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	var nodePool client.NodePool
	var fileContent []byte
	if fileContent, err = os.ReadFile(setNodePoolOptions.FilePath); err != nil {
		return errors.New(fmt.Sprintf("Unable to read file %s", setNodePoolOptions.FilePath))
	}

	if err = yaml.Unmarshal(fileContent, &nodePool); err != nil {
		return errors.Wrap(err, "Could not parse file contents")
	}

	options := client.SetMachineDeploymentOptions{
		ClusterName: clusterName,
		Namespace:   setNodePoolOptions.Namespace,
		NodePool:    nodePool,
	}

	return tkgctlClient.SetMachineDeployment(&options)
}
