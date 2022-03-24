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

type setMachineHealthCheckCPOptions struct {
	machineHealthCheckName string
	namespace              string
	matchLabels            string
	unhealthyConditions    string
	nodeStartupTimeout     string
}

var setMHCCP = &setMachineHealthCheckCPOptions{}

var setMachineHealthCheckCPCmd = &cobra.Command{
	Use:          "set CLUSTER_NAME",
	Short:        "Create or update a MachineHealthCheck for a cluster",
	Long:         "Create or update a MachineHealthCheck for a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         setMachineHealthCheckCP,
	SilenceUsage: true,
}

func init() {
	setMachineHealthCheckCPCmd.Flags().StringVarP(&setMHCCP.machineHealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	setMachineHealthCheckCPCmd.Flags().StringVarP(&setMHCCP.namespace, "namespace", "n", "", "Namespace of the cluster")
	setMachineHealthCheckCPCmd.Flags().StringVar(&setMHCCP.nodeStartupTimeout, "node-startup-timeout", "", "Any machine being created that takes longer than this duration to join the cluster is considered to have failed and will be remediated")
	setMachineHealthCheckCPCmd.Flags().StringVar(&setMHCCP.matchLabels, "match-labels", "", "Label selector to match machines whose health will be exercised")
	setMachineHealthCheckCPCmd.Flags().StringVar(&setMHCCP.unhealthyConditions, "unhealthy-conditions", "", "A list of the conditions that determine whether a node is considered unhealthy. Available condition types: [Ready, MemoryPressure,DiskPressure,PIDPressure, NetworkUnavailable], Available condition status: [True, False, Unknown]")
	machineHealthCheckControlPlaneCmd.AddCommand(setMachineHealthCheckCPCmd)
}

func setMachineHealthCheckCP(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("setting machine healthcheck with a global server is not implemented yet")
	}
	return runCreateMachineHealthCheckCP(server, args[0])
}

func runCreateMachineHealthCheckCP(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	if setMHCCP.matchLabels == "" {
		setMHCCP.matchLabels = controlPlaneLabel + ": "
	}

	options := tkgctl.SetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		MachineHealthCheckName: setMHCCP.machineHealthCheckName,
		Namespace:              setMHCCP.namespace,
		MatchLabels:            setMHCCP.matchLabels,
		UnhealthyConditions:    setMHCCP.unhealthyConditions,
		NodeStartupTimeout:     setMHCCP.nodeStartupTimeout,
	}
	return tkgctlClient.SetMachineHealthCheck(options)
}
