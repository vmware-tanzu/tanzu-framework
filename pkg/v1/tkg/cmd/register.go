// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a cluster to Tanzu Mission Control",
	Long:  "Register a cluster to Tanzu Mission Control",
}

func init() {
	// hide the register command until TMC supports registering TKG clusters
	registerCmd.Hidden = true
	RootCmd.AddCommand(registerCmd)
}
