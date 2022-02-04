// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main // nolint:dupl

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type deleteMachineHealthCheckNodeOptions struct {
	machinehealthCheckName string
	namespace              string
	unattended             bool
	matchLabel             string
}

var deleteMHCNode = &deleteMachineHealthCheckNodeOptions{}

var deleteMachineHealthCheckNodeCmd = &cobra.Command{
	Use:   "delete CLUSTER_NAME",
	Short: "Delete a MachineHealthCheck object",
	Long:  "Delete a MachineHealthCheck object of the nodes of a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteMachineHealthCheckNode,
}

func init() {
	deleteMachineHealthCheckNodeCmd.Flags().BoolVarP(&deleteMHCNode.unattended, "yes", "y", false, "Delete the MachineHealthCheck object without asking for confirmation")
	deleteMachineHealthCheckNodeCmd.Flags().StringVarP(&deleteMHCNode.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	deleteMachineHealthCheckNodeCmd.Flags().StringVarP(&deleteMHCNode.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created, default to the cluster's namespace")
	machineHealthCheckNodeCmd.AddCommand(deleteMachineHealthCheckNodeCmd)
}

func deleteMachineHealthCheckNode(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("deleting machine healthcheck with a global server is not implemented yet")
	}
	return runDeleteMachineHealthCheckNode(server, args[0])
}

func runDeleteMachineHealthCheckNode(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	if deleteMHCNode.matchLabel == "" {
		deleteMHCNode.matchLabel = nodePoolLabel
	}

	options := tkgctl.DeleteMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              deleteMHCNode.namespace,
		MachinehealthCheckName: deleteMHCNode.machinehealthCheckName,
		SkipPrompt:             deleteMHCNode.unattended,
		MatchLabel:             deleteMHCNode.matchLabel,
	}
	return tkgctlClient.DeleteMachineHealthCheck(options)
}
