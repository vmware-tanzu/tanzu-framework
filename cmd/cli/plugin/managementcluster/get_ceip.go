// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

var getCeipCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"ceip", "ceip-participations"},
	Short:   "Get the current CEIP opt-in status of the current management cluster",
	Long:    "Get the current CEIP opt-in status of the current management cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetCEIP(cmd)
	},
}

func init() {
	getCeipCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")

	ceipCmd.AddCommand(getCeipCmd)
}

func runGetCEIP(cmd *cobra.Command) error {
	forceUpdateTKGCompatibilityImage := false
	tkgClient, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	if err != nil {
		return err
	}

	ceipStatus, err := tkgClient.GetCEIP()
	if err != nil {
		return err
	}

	t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Management-Cluster-Name", "CEIP-Status")
	t.AddRow(ceipStatus.ClusterName, ceipStatus.CeipStatus)
	t.Render()

	return nil
}
