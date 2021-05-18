// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

type getKubernetesVersionsOptions struct {
	outputFormat string
}

var gkv = &getKubernetesVersionsOptions{}

var getkvCmd = &cobra.Command{
	Use:     "kubernetesversions",
	Aliases: []string{"kv", "kubernetesversion"},
	Short:   "Get the list of supported kubernetes versions for workload clusters",
	Long:    "Get the list of supported kubernetes versions for workload clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetTKR()
	},
}

func init() {
	getkvCmd.Flags().StringVarP(&gkv.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")

	getCmd.AddCommand(getkvCmd)
}

func runGetTKR() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	tkrInfo, err := tkgClient.GetKubernetesVersions()
	if err != nil {
		return err
	}
	sort.Strings(tkrInfo.Versions)

	// if output format is specified use that output format to render output
	// if not table format will be used
	if gkv.outputFormat != "" {
		return utils.RenderOutput(tkrInfo, gkv.outputFormat)
	}

	t := utils.CreateTableWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Versions"})
	for _, k8sVersion := range tkrInfo.Versions {
		t.AppendRow(table.Row{k8sVersion})
	}
	t.Render()

	return nil
}
