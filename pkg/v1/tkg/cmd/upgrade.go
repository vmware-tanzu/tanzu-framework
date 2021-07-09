// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade a Tanzu Kubernetes cluster",
	Long:  "Upgrade a Tanzu Kubernetes cluster",
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
}
