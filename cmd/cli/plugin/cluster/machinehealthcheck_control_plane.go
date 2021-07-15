// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var machineHealthCheckControlPlaneCmd = &cobra.Command{
	Use:   "control-plane",
	Short: "MachineHealthCheck operations for the control plane a cluster",
	Long:  `Get, set, or delete a MachineHealthCheck object for the control plane of a Tanzu Kubernetes cluster`,
}
