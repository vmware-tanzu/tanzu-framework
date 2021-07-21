// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var clusterNodePoolCmd = &cobra.Command{
	Use:   "node-pool",
	Short: "Cluster node-pool operations",
}
