// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var (
	isProd string
	labels string
)

var setCeipCmd = &cobra.Command{
	Use:     "ceip-participation OPT_IN_BOOL",
	Short:   "Set the opt-in preference for CEIP",
	Long:    "Set the opt-in preference for CEIP of the current management cluster",
	Aliases: []string{"ceip", "ceip-participations"},
	Args:    cobra.ExactArgs(1),
	Example: Examples(`
		# Change current management cluster to opt-in to VMware CEIP
		tkg set ceip-participation true

		# Change current management cluster to opt-out of VMware CEIP
		tkg set ceip-participation false

		[*] : VMware's Customer Experience Improvement Program ("CEIP") provides VMware with information that enables
		VMware to improve its products and services and fix problems. By choosing to participate in CEIP, you agree that
		VMware may collect technical information about your use of VMware products and services on a regular basis. This
		information does not personally identify you.
	`),
	Run: func(cmd *cobra.Command, args []string) {
		err := runSetCeip(args[0])
		verifyCommandError(err)
	},
}

func init() {
	setCeipCmd.Flags().StringVarP(&isProd, "isProd", "", "", "use --isProd false to write telemetry data to the staging datastore")
	setCeipCmd.Flags().MarkHidden("isProd") //nolint
	setCeipCmd.Flags().StringVarP(&labels, "labels", "", "", "use --labels=entitlement-account-number=\"num1\",env-type=\"env\" to self-identify the customer's account number and environmment")
	setCmd.AddCommand(setCeipCmd)
}

func runSetCeip(ceipOptIn string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	return tkgClient.SetCeip(ceipOptIn, isProd, labels)
}
