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

type deleteMachineHealthCheckCPOptions struct {
	machinehealthCheckName string
	namespace              string
	unattended             bool
	matchLabel             string
}

var deleteMHCCP = &deleteMachineHealthCheckCPOptions{}

var deleteMachineHealthCheckCPCmd = &cobra.Command{
	Use:   "delete CLUSTER_NAME",
	Short: "Delete a MachineHealthCheck object of the control plane of a cluster",
	Long:  "Delete a MachineHealthCheck object of the control plane of a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteMachineHealthCheckCP,
}

func init() {
	deleteMachineHealthCheckCPCmd.Flags().BoolVarP(&deleteMHCCP.unattended, "yes", "y", false, "Delete the MachineHealthCheck object without asking for confirmation")
	deleteMachineHealthCheckCPCmd.Flags().StringVarP(&deleteMHCCP.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	deleteMachineHealthCheckCPCmd.Flags().StringVarP(&deleteMHCCP.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created, default to the cluster's namespace")
	machineHealthCheckControlPlaneCmd.AddCommand(deleteMachineHealthCheckCPCmd)
}

func deleteMachineHealthCheckCP(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting machine healthcheck with a global server is not implemented yet")
	}
	return runDeleteMachineHealthCheckCP(server, args[0])
}

func runDeleteMachineHealthCheckCP(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	if deleteMHCCP.matchLabel == "" {
		deleteMHCCP.matchLabel = "cluster.x-k8s.io/control-plane"
	}

	options := tkgctl.DeleteMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              deleteMHCCP.namespace,
		MachinehealthCheckName: deleteMHCCP.machinehealthCheckName,
		SkipPrompt:             deleteMHCCP.unattended,
		MatchLabel:             deleteMHCCP.matchLabel,
	}
	return tkgctlClient.DeleteMachineHealthCheck(options)
}
