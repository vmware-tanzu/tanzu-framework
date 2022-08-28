// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var scaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "Scale a Tanzu Kubernetes cluster",
	Long:  "Scale a Tanzu Kubernetes cluster",
}

func init() {
	RootCmd.AddCommand(scaleCmd)
}
