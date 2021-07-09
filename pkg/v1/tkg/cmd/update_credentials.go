// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var updateCredsCommand = &cobra.Command{
	Use:   "update-credentials",
	Short: "Update a Tanzu Kubernetes cluster credentials",
	Long:  "Update a Tanzu Kubernetes cluster credentials",
}

func init() {
	RootCmd.AddCommand(updateCredsCommand)
}
