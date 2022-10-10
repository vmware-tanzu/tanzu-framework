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
		// Hide the default completion command of the plugin.
		// Shell completion is enabled using the Tanzu CLI's `completion` command so a plugin
		// does not need its own `completion` command.  Having such a command is just
		// confusing for users. However, we don't disable it completely for two reasons:
		//   1. backwards-compatibility, as the command used to be available for some plugins
		//   2. to allow shell completion when using the plugin as a native program (mostly for testing)
		// Note that a plugin can completely disable this command itself using:
		//  plugin.Cmd.CompletionOptions.DisableDefaultCmd = true
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
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
