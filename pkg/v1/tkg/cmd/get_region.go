// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

type getRegionsOptions struct {
	clusterName  string
	outputFormat string
}

var gr = &getRegionsOptions{}

var getRegionsCmd = &cobra.Command{
	Use:     "management-cluster",
	Aliases: []string{"mc", "management-clusters"},
	Short:   "Get the currently defined management cluster contexts",
	Long:    "Get the currently defined management cluster contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetRegions()
	},
}

func init() {
	getRegionsCmd.Flags().StringVarP(&gr.clusterName, "name", "n", "", "The name of the management cluster")
	getRegionsCmd.Flags().StringVarP(&gr.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")

	getCmd.AddCommand(getRegionsCmd)
}

func runGetRegions() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	regions, err := tkgClient.GetRegions(gr.clusterName)
	if err != nil {
		return err
	}

	// if output format is specified use that output format to render output
	// if not table format will be used
	if gr.outputFormat != "" {
		return utils.RenderOutput(regions, gr.outputFormat)
	}

	t := utils.CreateTableWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Management-Cluster-Name", "Context-Name", "Status"})
	for _, r := range regions {
		if r.IsCurrentContext {
			t.AppendRow(table.Row{r.ClusterName + " *", r.ContextName, r.Status})
		} else {
			t.AppendRow(table.Row{r.ClusterName, r.ContextName, r.Status})
		}
	}
	t.Render()

	return nil
}
