// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
)

type versionInfo struct {
	Version   string
	GitCommit string
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of the Tanzu Kubernetes Grid CLI",
	Long:  "Display the version of the Tanzu Kubernetes Grid CLI",

	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion(cmd)
	},
}

func init() {
	versionCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml|table")

	RootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command) error {
	verInfo := versionInfo{Version: buildinfo.Version, GitCommit: buildinfo.SHA}

	if outputFormat == string(component.JSONOutputType) || outputFormat == string(component.YAMLOutputType) {
		output := component.NewObjectWriter(cmd.OutOrStdout(), outputFormat, verInfo)
		output.Render()
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Client:")
	fmt.Fprintf(cmd.OutOrStdout(), "\tVersion: %s\n", verInfo.Version)
	fmt.Fprintf(cmd.OutOrStdout(), "\tGit commit: %s\n", verInfo.GitCommit)

	return nil
}
