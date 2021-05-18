// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import "github.com/spf13/cobra"

var deRegisterCmd = &cobra.Command{
	Use:   "deregister",
	Short: "Deregister a cluster from Tanzu Mission Control",
	Long:  "Deregister a cluster from Tanzu Mission Control",
}

func init() {
	// hide the deregister command until TMC supports registering TKG clusters
	deRegisterCmd.Hidden = true
	RootCmd.AddCommand(deRegisterCmd)
}
