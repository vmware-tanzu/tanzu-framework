// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type setMachineHealthCheckOptions struct {
	machineHealthCheckName string
	namespace              string
	matchLabels            string
	unhealthyConditions    string
	nodeStartupTimeout     string
}

var setMHC = &setMachineHealthCheckOptions{}

var setMachineHealthCheckCmd = &cobra.Command{
	Use:   "machinehealthcheck CLUSTERNAME",
	Short: "Create or update a MachineHealthCheck for a cluster",
	Long:  "Create or update a MachineHealthCheck for a cluster",
	Example: Examples(`
		# Create a MachineHealthCheck with default configuration
		tkg set machinehealthcheck my-cluster

		# Create or update a MachineHealthCheck with customized NodeStartupTimeout
		tkg set machinehealthcheck my-custer --node-startup-timeout 10m

		# Create or update a MachineHealthCheck with customized name
		tkg set machinehealthcheck my-custer --mhc-name my-mhc

		# Create or update a MachineHealthCheck with customized UnhealthyConditions
		tkg set machinehealthcheck my-custer --unhealthy-conditions "Ready:False:5m,Ready:Unknown:5m"

		# Create or update a MachineHealthCheck with customized node labels
		tkg set machinehealthcheck my-custer  --match-labels "key1:value1,key2:value2"
	`),
	Aliases: []string{"mhc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runCreateMachineHealthCheck(args[0])
		verifyCommandError(err)
	},
}

func init() {
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.machineHealthCheckName, "mhc-name", "", "", "Name of the MachineHealthCheck object")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.namespace, "namespace", "", "", "Namespace of the cluster")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.nodeStartupTimeout, "node-startup-timeout", "", "", "Any machine being created that takes longer than this duration to join the cluster is considered to have failed and will be remediated")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.matchLabels, "match-labels", "", "", "Label selector to match machines whose health will be exercised")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.unhealthyConditions, "unhealthy-conditions", "", "", "A list of the conditions that determine whether a node is considered unhealthy. Available condition types: [Ready, MemoryPressure,DiskPressure,PIDPressure, NetworkUnavailable], Available condition status: [True, False, Unknown]")
	setCmd.AddCommand(setMachineHealthCheckCmd)
}

func runCreateMachineHealthCheck(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.SetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		MachineHealthCheckName: setMHC.machineHealthCheckName,
		Namespace:              setMHC.namespace,
		MatchLabels:            setMHC.matchLabels,
		UnhealthyConditions:    setMHC.unhealthyConditions,
		NodeStartupTimeout:     setMHC.nodeStartupTimeout,
	}
	return tkgClient.SetMachineHealthCheck(options)
}
