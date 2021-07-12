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

type setMachineHealthCheckNodeOptions struct {
	machineHealthCheckName string
	namespace              string
	matchLabels            string
	unhealthyConditions    string
	nodeStartupTimeout     string
}

var setMHCNode = &setMachineHealthCheckNodeOptions{}

var setMachineHealthCheckNodeCmd = &cobra.Command{
	Use:   "set CLUSTER_NAME",
	Short: "Create or update a MachineHealthCheck for a cluster",
	Long:  "Create or update a MachineHealthCheck for a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  setMachineHealthCheckNode,
}

func init() {
	setMachineHealthCheckNodeCmd.Flags().StringVarP(&setMHCNode.machineHealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	setMachineHealthCheckNodeCmd.Flags().StringVarP(&setMHCNode.namespace, "namespace", "n", "", "Namespace of the cluster")
	setMachineHealthCheckNodeCmd.Flags().StringVar(&setMHCNode.nodeStartupTimeout, "node-startup-timeout", "", "Any machine being created that takes longer than this duration to join the cluster is considered to have failed and will be remediated")
	setMachineHealthCheckNodeCmd.Flags().StringVar(&setMHCNode.matchLabels, "match-labels", "", "Label selector to match machines whose health will be exercised")
	setMachineHealthCheckNodeCmd.Flags().StringVar(&setMHCNode.unhealthyConditions, "unhealthy-conditions", "", "A list of the conditions that determine whether a node is considered unhealthy. Available condition types: [Ready, MemoryPressure,DiskPressure,PIDPressure, NetworkUnavailable], Available condition status: [True, False, Unknown]")
	machineHealthCheckNodeCmd.AddCommand(setMachineHealthCheckNodeCmd)
}

func setMachineHealthCheckNode(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("setting machine healthcheck with a global server is not implemented yet")
	}
	return runCreateMachineHealthCheckNode(server, args[0])
}

func runCreateMachineHealthCheckNode(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	options := tkgctl.SetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		MachineHealthCheckName: setMHCNode.machineHealthCheckName,
		Namespace:              setMHCNode.namespace,
		MatchLabels:            setMHCNode.matchLabels,
		UnhealthyConditions:    setMHCNode.unhealthyConditions,
		NodeStartupTimeout:     setMHCNode.nodeStartupTimeout,
	}
	return tkgctlClient.SetMachineHealthCheck(options)
}
