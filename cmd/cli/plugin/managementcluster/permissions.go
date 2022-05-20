// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/spf13/cobra"

var permissionsCmd = &cobra.Command{
	Use:          "permissions",
	Short:        "Configure permissions on cloud providers",
	SilenceUsage: true,
}

var awsPermissionsCmd = &cobra.Command{
	Use:          "aws",
	Short:        "Configure permissions on AWS",
	SilenceUsage: true,
}

func init() {
	permissionsCmd.AddCommand(awsPermissionsCmd)
}
