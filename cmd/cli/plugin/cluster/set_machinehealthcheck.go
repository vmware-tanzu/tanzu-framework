// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"

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
	Use:   "set CLUSTER_NAME",
	Short: "Create or update a MachineHealthCheck for a cluster",
	Long:  "Create or update a MachineHealthCheck for a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  setMachineHealthCheck,
}

func init() {
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.machineHealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.namespace, "namespace", "n", "", "Namespace of the cluster")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.nodeStartupTimeout, "node-startup-timeout", "", "", "Any machine being created that takes longer than this duration to join the cluster is considered to have failed and will be remediated")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.matchLabels, "match-labels", "", "", "Label selector to match machines whose health will be exercised")
	setMachineHealthCheckCmd.Flags().StringVarP(&setMHC.unhealthyConditions, "unhealthy-conditions", "", "", "A list of the conditions that determine whether a node is considered unhealthy. Available condition types: [Ready, MemoryPressure,DiskPressure,PIDPressure, NetworkUnavailable], Available condition status: [True, False, Unknown]")
	machineHealthCheckCmd.AddCommand(setMachineHealthCheckCmd)
}

func setMachineHealthCheck(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("setting machine healthcheck with a global server is not implemented yet")
	}
	return runCreateMachineHealthCheck(server, args[0])
}

func runCreateMachineHealthCheck(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
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
	return tkgctlClient.SetMachineHealthCheck(options)
}
