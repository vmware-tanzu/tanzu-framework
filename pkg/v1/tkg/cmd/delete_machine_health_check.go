// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type deleteMachineHealthCheckOptions struct {
	machinehealthCheckName string
	namespace              string
	unattended             bool
}

var deleteMHC = &deleteMachineHealthCheckOptions{}

var deleteMachineHealthCheckCmd = &cobra.Command{
	Use:   "machinehealthcheck CLUSTER_NAME",
	Short: "Delete a MachineHealthCheck object",
	Long:  "Delete a MachineHealthCheck object for the given cluster",
	Example: Examples(`
	# Delete a MachineHealthCheck object of a cluster. By default, CLUSTER_NAME will be used as the name of the MachineHealthCheck
	tkg delete machinehealthcheck my-cluster

	# Delete a MachineHealthCheck object when there are several MachineHealthCheck objects associated with a cluster
	tkg delete machinehealthcheck my-cluster --mhc-name my-mhc`),
	Aliases: []string{"mhc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runDeleteMachineHealthCheck(args[0])
		verifyCommandError(err)
	},
}

func init() {
	deleteMachineHealthCheckCmd.Flags().BoolVarP(&deleteMHC.unattended, "yes", "y", false, "Delete the MachineHealthCheck object without asking for confirmation")
	deleteMachineHealthCheckCmd.Flags().StringVarP(&deleteMHC.machinehealthCheckName, "mhc-name", "", "", "Name of the MachineHealthCheck object")
	deleteMachineHealthCheckCmd.Flags().StringVarP(&deleteMHC.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created, default to the cluster's namespace")
	deleteCmd.AddCommand(deleteMachineHealthCheckCmd)
}

func runDeleteMachineHealthCheck(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.DeleteMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              deleteMHC.namespace,
		MachinehealthCheckName: deleteMHC.machinehealthCheckName,
		SkipPrompt:             deleteMHC.unattended || skipPrompt,
	}
	return tkgClient.DeleteMachineHealthCheck(options)
}
