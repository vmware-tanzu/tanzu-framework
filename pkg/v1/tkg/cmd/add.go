// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an existing resource to the current configuration",
	Long:  "Add an existing resource to the current configuration",
}

func init() {
	RootCmd.AddCommand(addCmd)
}
