// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Tanzu Kubernetes cluster",
	Long:  `Create a Tanzu Kubernetes cluster`,
}

func init() {
	RootCmd.AddCommand(createCmd)
}
