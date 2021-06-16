// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
)

// RootCmd is the core root Tanzu command
var RootCmd = &cobra.Command{
	Use: "tanzu",
}

var (
	noInit bool
	color  = true
)

// NewRootCmd creates a root command.
func NewRootCmd() (*cobra.Command, error) {
	u := cli.NewMainUsage()
	RootCmd.SetUsageFunc(u.Func())

	ni := os.Getenv("TANZU_CLI_NO_INIT")
	if ni != "" {
		noInit = true
	}
	if os.Getenv("TANZU_CLI_NO_COLOR") != "" {
		color = false
	}

	au := aurora.NewAurora(color)
	RootCmd.Short = au.Bold(`Tanzu CLI`).String()

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

	plugins, err := cli.ListPlugins()
	if err != nil {
		return nil, fmt.Errorf("find available plugins: %w", err)
	}

	if err = config.CopyLegacyConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to copy legacy configuration directory to new location: %w", err)
	}

	// check that all plugins in the core distro are installed or do so.
	if !noInit && !cli.IsDistributionSatisfied(plugins) {
		s := spin.New("%s   initializing")
		s.Start()
		cfg, err := config.GetClientConfig()
		if err != nil {
			log.Fatal(err)
		}
		repos := cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
		err = cli.EnsureDistro(repos)
		if err != nil {
			return nil, err
		}
		plugins, err = cli.ListPlugins()
		if err != nil {
			return nil, fmt.Errorf("find available plugins: %w", err)
		}
		s.Stop()
	}
	for _, plugin := range plugins {
		RootCmd.AddCommand(cli.GetCmd(plugin))
	}

	duplicateAliasWarning()

	// Flag parsing must be deactivated because the root plugin won't know about all flags.
	RootCmd.DisableFlagParsing = true

	return RootCmd, nil
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
