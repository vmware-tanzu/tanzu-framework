// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
)

const (
	controlPlaneLabel      = "cluster.x-k8s.io/control-plane"
	nodePoolLabel          = "node-pool"
	machineDeploymentLabel = "topology.cluster.x-k8s.io/deployment-name"
)

var machineHealthCheckCmd = &cobra.Command{
	Use:          "machinehealthcheck",
	Short:        "MachineHealthCheck operations for a cluster",
	Long:         `Get, set, or delete a MachineHealthCheck object for a Tanzu Kubernetes cluster`,
	Aliases:      []string{"mhc"},
	SilenceUsage: true,
}

func init() {
	machineHealthCheckCmd.AddCommand(machineHealthCheckControlPlaneCmd)
	machineHealthCheckCmd.AddCommand(machineHealthCheckNodeCmd)

	cli.DeprecateCommandWithAlternative(deleteMachineHealthCheckCmd, "1.5.0", "tanzu cluster machinehealthcheck node delete or tanzu cluster machinehealthcheck control-plane delete")
	cli.DeprecateCommandWithAlternative(getMachineHealthCheckCmd, "1.5.0", "tanzu cluster machinehealthcheck node get or tanzu cluster machinehealthcheck control-plane get")
	cli.DeprecateCommandWithAlternative(setMachineHealthCheckCmd, "1.5.0", "tanzu cluster machinehealthcheck node set or tanzu cluster machinehealthcheck control-plane set")
}
