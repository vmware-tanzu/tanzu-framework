// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate Cluster API provider configuration",
	Long:  "Generate Cluster API provider configuration and cluster plans for creating Tanzu Kubernetes clusters",
}

func init() {
	RootCmd.AddCommand(configCmd)
}
