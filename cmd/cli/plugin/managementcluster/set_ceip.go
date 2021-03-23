// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var isProd string
var labels string

var setCeipCmd = &cobra.Command{
	Use:   "set OPT_IN_BOOL",
	Short: "Set the opt-in preference for CEIP of the current management cluster",
	Long:  "Set the opt-in preference for CEIP of the current management cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetCeip,
}

func init() {
	setCeipCmd.Flags().StringVarP(&isProd, "isProd", "", "", "use --isProd false to write telemetry data to the staging datastore")
	setCeipCmd.Flags().MarkHidden("isProd") //nolint
	setCeipCmd.Flags().StringVarP(&labels, "labels", "", "", "use --labels=entitlement-account-number=\"num1\",env-type=\"env\" to self-identify the customer's account number and environmment")
	ceipCmd.AddCommand(setCeipCmd)
}

func runSetCeip(cmd *cobra.Command, args []string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	return tkgClient.SetCeip(args[0], isProd, labels)
}
