// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

func init() {
	versionCmd.SetUsageFunc(cli.SubCmdUsageFunc)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Annotations: map[string]string{
		"group": string(cliapi.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("version: %s\nbuildDate: %s\nsha: %s\n", buildinfo.Version, buildinfo.Date, buildinfo.SHA)
		return nil
	},
}
