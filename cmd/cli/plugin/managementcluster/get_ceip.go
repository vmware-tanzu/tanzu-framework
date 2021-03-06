// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/utils"
)

type getCeipOptions struct {
	outputFormat string
}

var gceip = &getCeipOptions{}

var getCeipCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"ceip", "ceip-participations"},
	Short:   "Get the current CEIP opt-in status of the current management cluster",
	Long:    "Get the current CEIP opt-in status of the current management cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetCEIP()
	},
}

func init() {
	getCeipCmd.Flags().StringVarP(&gceip.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")

	ceipCmd.AddCommand(getCeipCmd)
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

	t := component.NewTableWriter("Management-Cluster-Name", "CEIP-Status")
	t.Append([]string{ceipStatus.ClusterName, ceipStatus.CeipStatus})
	t.Render()

	return nil
}
