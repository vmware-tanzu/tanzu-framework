// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

type getCeipOptions struct {
	outputFormat string
}

var gceip = &getCeipOptions{}

var getCeipCmd = &cobra.Command{
	Use:     "ceip-participation",
	Aliases: []string{"ceip", "ceip-participations"},
	Short:   "Get the current CEIP opt-in status of the current management cluster",
	Long:    "Get the current CEIP opt-in status of the current management cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetCEIP()
	},
}

func init() {
	getCeipCmd.Flags().StringVarP(&gceip.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")

	getCmd.AddCommand(getCeipCmd)
}

func runGetCEIP() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	ceipStatus, err := tkgClient.GetCEIP()
	if err != nil {
		return err
	}

	// if output format is specified use that output format to render output
	// if not table format will be used
	if gceip.outputFormat != "" {
		return utils.RenderOutput(ceipStatus, gceip.outputFormat)
	}

	t := utils.CreateTableWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Management-Cluster-Name", "CEIP-Status"})
	t.AppendRow(table.Row{ceipStatus.ClusterName, ceipStatus.CeipStatus})
	t.Render()

	return nil
}
