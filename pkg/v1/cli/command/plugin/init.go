// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
)

func newInitCmd(desc *cliv1alpha1.PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Initialize a plugin",
		Long:         "Initialize a plugin",
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// invoke postInstall for the plugin
			return desc.PostInstallHook()
		},
	}

	return cmd
}
