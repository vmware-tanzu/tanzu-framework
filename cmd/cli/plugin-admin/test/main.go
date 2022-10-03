// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	pluginrtbuildinfo "github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

var descriptor = cliapi.PluginDescriptor{
	Name:        "test",
	Description: "Test the CLI",
	Group:       cliapi.AdminCmdGroup,
	// TODO: When plugins have their own buildInfo, we need to update "Version" and "BuildSHA"
	// 		to plugin own versions instead of plugin runtime version
	Version:              pluginrtbuildinfo.Version,
	BuildSHA:             pluginrtbuildinfo.SHA,
	PluginRuntimeVersion: pluginrtbuildinfo.Version,
}

var local string

func init() {
	fetchCmd.Flags().StringVarP(&local, "local", "l", "", "path to local repository")
	_ = fetchCmd.MarkFlagRequired("local")
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.AddCommands(
		fetchCmd,
		pluginsCmd,
	)

	_, standalonePlugins, err := pluginmanager.InstalledPlugins("")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range standalonePlugins {
		// Check if test plugin binary installed. If available add a plugin command
		_, err := os.Stat(cli.TestPluginPathFromPluginPath(d.InstallationPath))
		if err != nil {
			continue
		}
		pluginsCmd.AddCommand(cli.TestCmd(d.DeepCopy()))
	}

	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}

var pluginsCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin tests",
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch the plugin tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pluginmanager.InstallPluginsFromLocalSource("all", "", local, true)
	},
}
