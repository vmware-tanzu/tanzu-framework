// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import "github.com/spf13/cobra"

var packageBundleCmd = &cobra.Command{
	Use:   "package-bundle",
	Short: "Package bundle operations",
}

func init() {
	rootCmd.AddCommand(packageBundleCmd)
}
