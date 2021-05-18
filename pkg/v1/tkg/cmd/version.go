// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/buildinfo"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

type versionOptions struct {
	outputFormat string
}

var vo = &versionOptions{}

type versionInfo struct {
	Version   string
	GitCommit string
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of the Tanzu Kubernetes Grid CLI",
	Long:  "Display the version of the Tanzu Kubernetes Grid CLI",

	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion()
	},
}

func init() {
	versionCmd.Flags().StringVarP(&vo.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")

	RootCmd.AddCommand(versionCmd)
}

func runVersion() error {
	verInfo := versionInfo{Version: buildinfo.Version, GitCommit: buildinfo.Commit}

	if vo.outputFormat != "" {
		return utils.RenderOutput(verInfo, vo.outputFormat)
	}

	fmt.Println("Client:")
	fmt.Printf("\tVersion: %s\n", verInfo.Version)
	fmt.Printf("\tGit commit: %s\n", verInfo.GitCommit)

	return nil
}
