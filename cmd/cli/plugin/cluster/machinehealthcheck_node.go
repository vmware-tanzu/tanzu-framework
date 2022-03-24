// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var machineHealthCheckNodeCmd = &cobra.Command{
	Use:          "node",
	Short:        "MachineHealthCheck operations for the nodes of a cluster",
	Long:         `Get, set, or delete a MachineHealthCheck object for the nodes of a Tanzu Kubernetes cluster`,
	SilenceUsage: true,
}
