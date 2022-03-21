// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import "github.com/spf13/cobra"

var repositoryBundleCmd = &cobra.Command{
	Use:   "repo-bundle",
	Short: "Package repo bundle operations",
}

func init() {
	rootCmd.AddCommand(repositoryBundleCmd)
}
