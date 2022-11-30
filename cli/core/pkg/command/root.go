// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package command provides commands
package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

// RootCmd is the core root Tanzu command
var RootCmd = &cobra.Command{
	Use: "tanzu",
	// Don't have Cobra print the error message, the CLI will
	// print it itself in a nicer format.
	SilenceErrors: true,
}

// NewRootCmd creates a root command.
func NewRootCmd() (*cobra.Command, error) {
	uFunc := cli.NewMainUsage().Func()
	RootCmd.SetUsageFunc(uFunc)
	k8sCmd.SetUsageFunc(uFunc)
	tmcCmd.SetUsageFunc(uFunc)

	// configure defined environment variables under tanzu config file
	cliconfig.ConfigureEnvVariables()

	RootCmd.Short = component.Bold(`Tanzu CLI`)

	// TODO (pbarker): silencing usage for now as we are getting double usage from plugins on errors
	RootCmd.SilenceUsage = true

	RootCmd.AddCommand(
		pluginCmd,
		initCmd,
		updateCmd,
		versionCmd,
		completionCmd,
		configCmd,
		genAllDocsCmd,
	)

	// If the context and target feature is enabled, add the corresponding commands under root.
	if config.IsFeatureActivated(cliconfig.FeatureContextCommand) {
		RootCmd.AddCommand(
			contextCmd,
			k8sCmd,
			tmcCmd,
		)
		mapCtxTypeToCmd := map[cliv1alpha1.Target]*cobra.Command{
			cliv1alpha1.TargetK8s: k8sCmd,
			cliv1alpha1.TargetTMC: tmcCmd,
		}
		if err := addPluginsToTarget(mapCtxTypeToCmd); err != nil {
			return nil, err
		}
	}

	serverPlugins, standalonePlugins, err := pluginmanager.InstalledPlugins()
	if err != nil {
		return nil, err
	}

	plugins := serverPlugins
	plugins = append(plugins, standalonePlugins...)

	if err = config.CopyLegacyConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to copy legacy configuration directory to new location: %w", err)
	}

	var maskedPlugins []string

	for i := range plugins {
		// Only add plugins that should be available as root level command
		if isPluginRootCmdTargeted(&plugins[i]) {
			cmd := cli.GetCmd(&plugins[i])
			// check and find if same command already exists as part of root command
			matchedCmd := findSubCommand(RootCmd, cmd)
			if matchedCmd == nil { // If the subcommand for the plugin doesn't exist add the command
				RootCmd.AddCommand(cmd)
			} else if plugins[i].Scope == common.PluginScopeContext && isStandalonePluginCommand(matchedCmd) {
				// If the subcommand already exists but plugin is `Context-Scoped` plugin
				// then context-scoped plugin gets higher precedence, so, replace the existing command
				// to point to new command by removing and adding new command.
				maskedPlugins = append(maskedPlugins, matchedCmd.Name())
				RootCmd.RemoveCommand(matchedCmd)
				RootCmd.AddCommand(cmd)
			} else {
				maskedPlugins = append(maskedPlugins, plugins[i].Name)
			}
		}
	}

	if len(maskedPlugins) > 0 {
		fmt.Fprintf(os.Stderr, "Warning, Masking commands for plugins %q because a core command or other plugin with that name already exists. \n", strings.Join(maskedPlugins, ", "))
	}

	duplicateAliasWarning()

	// Flag parsing must be deactivated because the root plugin won't know about all flags.
	RootCmd.DisableFlagParsing = true

	return RootCmd, nil
}

var k8sCmd = &cobra.Command{
	Use:     "kubernetes",
	Short:   "Tanzu CLI plugins that target a Kubernetes cluster",
	Aliases: []string{"k8s"},
	Annotations: map[string]string{
		"group": string(cliapi.TargetCmdGroup),
	},
}

var tmcCmd = &cobra.Command{
	Use:     "mission-control",
	Short:   "Tanzu CLI plugins that target a Tanzu Mission Control endpoint",
	Aliases: []string{"tmc"},
	Annotations: map[string]string{
		"group": string(cliapi.TargetCmdGroup),
	},
}

func addPluginsToTarget(mapCtxTypeToCmd map[cliv1alpha1.Target]*cobra.Command) error {
	installedPlugins, standalonePlugins, err := pluginmanager.InstalledPlugins()
	if err != nil {
		return fmt.Errorf("unable to find installed plugins: %w", err)
	}

	installedPlugins = append(installedPlugins, standalonePlugins...)

	for i := range installedPlugins {
		if cmd, exists := mapCtxTypeToCmd[installedPlugins[i].Target]; exists {
			cmd.AddCommand(cli.GetCmd(&installedPlugins[i]))
		}
	}
	return nil
}

func duplicateAliasWarning() {
	var aliasMap = make(map[string][]string)
	for _, command := range RootCmd.Commands() {
		for _, alias := range command.Aliases {
			aliases, ok := aliasMap[alias]
			if !ok {
				aliasMap[alias] = []string{command.Name()}
			} else {
				aliasMap[alias] = append(aliases, command.Name())
			}
		}
	}

	for alias, plugins := range aliasMap {
		if len(plugins) > 1 {
			fmt.Fprintf(os.Stderr, "Warning, the alias %s is duplicated across plugins: %s\n\n", alias, strings.Join(plugins, ", "))
		}
	}
}

// Execute executes the CLI.
func Execute() error {
	root, err := NewRootCmd()
	if err != nil {
		return err
	}
	return root.Execute()
}

func findSubCommand(rootCmd, subCmd *cobra.Command) *cobra.Command {
	arrSubCmd := rootCmd.Commands()
	for i := range arrSubCmd {
		if arrSubCmd[i].Name() == subCmd.Name() {
			return arrSubCmd[i]
		}
	}
	return nil
}

func isPluginRootCmdTargeted(plugin *cliapi.PluginDescriptor) bool {
	// Only '<none>' targeted and `k8s` targeted plugins are considered root cmd targeted plugins
	return plugin != nil && (plugin.Target == cliv1alpha1.TargetNone || plugin.Target == cliv1alpha1.TargetK8s)
}

func isStandalonePluginCommand(cmd *cobra.Command) bool {
	scope, exists := cmd.Annotations["scope"]
	return exists && scope == common.PluginScopeStandalone
}
