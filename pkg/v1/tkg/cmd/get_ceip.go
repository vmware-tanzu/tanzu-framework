// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

var outputFormat string

var getCeipCmd = &cobra.Command{
	Use:     "ceip-participation",
	Aliases: []string{"ceip", "ceip-participations"},
	Short:   "Get the current CEIP opt-in status of the current management cluster",
	Long:    "Get the current CEIP opt-in status of the current management cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetCEIP(cmd)
	},
}

func init() {
	getCeipCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml|table")

	getCmd.AddCommand(getCeipCmd)
}

func runGetCEIP(cmd *cobra.Command) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	ceipStatus, err := tkgClient.GetCEIP()
	if err != nil {
		return err
	}
	var t component.OutputWriter
	if outputFormat == string(component.JSONOutputType) || outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), outputFormat, ceipStatus)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Management-Cluster-Name", "CEIP-Status")
		t.AddRow(ceipStatus.ClusterName, ceipStatus.CeipStatus)
	}
	t.Render()

	return nil
}
