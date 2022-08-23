// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	cli "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo"
)

func init() {
	versionCmd.SetUsageFunc(cli.SubCmdUsageFunc)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Annotations: map[string]string{
		"group": string(cliv1alpha1.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("version: %s\nbuildDate: %s\nsha: %s\n", buildinfo.Version, buildinfo.Date, buildinfo.SHA)
		return nil
	},
}
