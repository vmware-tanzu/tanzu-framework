// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a management cluster or Tanzu Kubernetes cluster",
	Long:  "Delete a management cluster or Tanzu Kubernetes cluster",
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
