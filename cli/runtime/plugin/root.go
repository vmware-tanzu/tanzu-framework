// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/spf13/cobra"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

func newRootCmd(descriptor *cliapi.PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:     descriptor.Name,
		Short:   descriptor.Description,
		Aliases: descriptor.Aliases,
	}
	cobra.AddTemplateFuncs(TemplateFuncs)
	cmd.SetUsageTemplate(CmdTemplate)

	cmd.AddCommand(
		newDescribeCmd(descriptor.Description),
		newVersionCmd(descriptor.Version),
		newInfoCmd(descriptor),
	)

	return cmd
}
