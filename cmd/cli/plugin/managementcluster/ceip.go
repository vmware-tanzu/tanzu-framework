// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/spf13/cobra"

var ceipCmd = &cobra.Command{
	Use:          "ceip-participation",
	Short:        "Get or set ceip participation",
	Long:         `Get or set ceip participation`,
	SilenceUsage: true,
}
