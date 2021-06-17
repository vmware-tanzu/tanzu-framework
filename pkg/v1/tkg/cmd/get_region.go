// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
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
		return runGetRegions(cmd)
	},
}

func init() {
	getRegionsCmd.Flags().StringVarP(&gr.clusterName, "name", "n", "", "The name of the management cluster")
	getRegionsCmd.Flags().StringVarP(&gr.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml|table")

	getCmd.AddCommand(getRegionsCmd)
}

func runGetRegions(cmd *cobra.Command) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	regions, err := tkgClient.GetRegions(gr.clusterName)
	if err != nil {
		return err
	}

	var t component.OutputWriter
	if outputFormat == string(component.JSONOutputType) || outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), outputFormat, regions)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStderr(), gr.outputFormat, "Management-Cluster-Name", "Context-Name", "Status")
		for _, r := range regions {
			if r.IsCurrentContext {
				t.AddRow(r.ClusterName+" *", r.ContextName, r.Status)
			} else {
				t.AddRow(r.ClusterName, r.ContextName, r.Status)
			}
		}
	}
	t.Render()

	return nil
}
