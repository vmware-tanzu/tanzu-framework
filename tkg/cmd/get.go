// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Tanzu Kubernetes Grid resource(s)",
	Long:  "Get Tanzu Kubernetes Grid resource(s)",
}

func init() {
	RootCmd.AddCommand(getCmd)
}
