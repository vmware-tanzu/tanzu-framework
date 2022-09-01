// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type deleteClustersOptions struct {
	namespace  string
	unattended bool
}

var dc = &deleteClustersOptions{}

var deleteClusterCmd = &cobra.Command{
	Use:          "delete CLUSTER_NAME",
	Short:        "Delete a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         deleteCmd,
	SilenceUsage: true,
}

func init() {
	deleteClusterCmd.Flags().StringVarP(&dc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	deleteClusterCmd.Flags().BoolVarP(&dc.unattended, "yes", "y", false, "Delete workload cluster without asking for confirmation")
}

func deleteCmd(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("deleting cluster with a global server is not implemented yet")
	}
	return deleteCluster(server, args[0])
}

func deleteCluster(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	deleteClusterOptions := tkgctl.DeleteClustersOptions{
		ClusterName: clusterName,
		Namespace:   dc.namespace,
		SkipPrompt:  dc.unattended,
	}

	return tkgctlClient.DeleteCluster(deleteClusterOptions)
}
