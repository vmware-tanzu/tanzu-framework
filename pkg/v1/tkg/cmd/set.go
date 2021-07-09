// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Configure some aspect of the Tanzu Kubernetes Grid CLI",
	Long:  "Configure some aspect of the Tanzu Kubernetes Grid CLI",
}

func init() {
	RootCmd.AddCommand(setCmd)
}
