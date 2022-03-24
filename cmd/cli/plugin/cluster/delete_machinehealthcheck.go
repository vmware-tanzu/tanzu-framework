// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type deleteMachineHealthCheckOptions struct {
	machinehealthCheckName string
	namespace              string
	unattended             bool
}

var deleteMHC = &deleteMachineHealthCheckOptions{}

var deleteMachineHealthCheckCmd = &cobra.Command{
	Use:          "delete CLUSTER_NAME",
	Short:        "Delete a MachineHealthCheck object of a cluster",
	Long:         "Delete a MachineHealthCheck object of a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         deleteMachineHealthCheck,
	SilenceUsage: true,
}

func init() {
	deleteMachineHealthCheckCmd.Flags().BoolVarP(&deleteMHC.unattended, "yes", "y", false, "Delete the MachineHealthCheck object without asking for confirmation")
	deleteMachineHealthCheckCmd.Flags().StringVarP(&deleteMHC.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	deleteMachineHealthCheckCmd.Flags().StringVarP(&deleteMHC.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created, default to the cluster's namespace")
	machineHealthCheckCmd.AddCommand(deleteMachineHealthCheckCmd)
}

func deleteMachineHealthCheck(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("deleting machine healthcheck with a global server is not implemented yet")
	}
	return runDeleteMachineHealthCheck(server, args[0])
}

func runDeleteMachineHealthCheck(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	options := tkgctl.DeleteMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              deleteMHC.namespace,
		MachinehealthCheckName: deleteMHC.machinehealthCheckName,
		SkipPrompt:             deleteMHC.unattended,
	}
	return tkgctlClient.DeleteMachineHealthCheck(options)
}
