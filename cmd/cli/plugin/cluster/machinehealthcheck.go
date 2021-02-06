// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var machineHealthCheckCmd = &cobra.Command{
	Use:   "machinehealthcheck",
	Short: "MachineHealthCheck operations for a cluster",
	Long:  `Get, set, or delete a MachineHealthCheck object for a Tanzu Kubernetes cluster`,
}
