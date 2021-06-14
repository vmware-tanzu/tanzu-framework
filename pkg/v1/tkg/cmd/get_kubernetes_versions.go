// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

var getkvCmd = &cobra.Command{
	Use:     "kubernetesversions",
	Aliases: []string{"kv", "kubernetesversion"},
	Short:   "Get the list of supported kubernetes versions for workload clusters",
	Long:    "Get the list of supported kubernetes versions for workload clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetTKR(cmd)
	},
}

func init() {
	getkvCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml|table")
	getCmd.AddCommand(getkvCmd)
}

func runGetTKR(cmd *cobra.Command) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	tkrInfo, err := tkgClient.GetKubernetesVersions()
	if err != nil {
		return err
	}
	sort.Strings(tkrInfo.Versions)

	var t component.OutputWriter
	if outputFormat == string(component.JSONOutputType) || outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), outputFormat, tkrInfo)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Versions")
		for _, k8sVersion := range tkrInfo.Versions {
			t.AddRow(k8sVersion)
		}
	}
	t.Render()

	return nil
}
