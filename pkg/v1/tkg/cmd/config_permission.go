// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var configPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Configure permissions on the cloud providers",
	Long:  "Configure permissions on the cloud providers",
}

func init() {
	configCmd.AddCommand(configPermissionsCmd)
}
