// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
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
		versionString := fmt.Sprintf("version: %s", buildinfo.Version)
		if len(buildinfo.CommunityEditionVersion) > 0 {
			versionString = fmt.Sprintf("%s %s", versionString, fmt.Sprintf("(Community Edition %s)", buildinfo.CommunityEditionVersion))
		}
		fmt.Printf("%s\n", versionString)
		fmt.Printf("buildDate: %s\nsha: %s\n", buildinfo.Date, buildinfo.SHA)
		return nil
	},
}
