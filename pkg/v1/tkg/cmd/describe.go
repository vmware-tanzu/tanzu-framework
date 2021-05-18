// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe Tanzu Kubernetes Grid resource(s)",
	Long:  "Describe Tanzu Kubernetes Grid resource(s)",
}

func init() {
	RootCmd.AddCommand(describeCmd)
}
