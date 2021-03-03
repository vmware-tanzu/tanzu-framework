// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

var root = &cobra.Command{
	Use:   "tanzu",
	Short: aurora.Bold(`Tanzu CLI`).String(),
}

var noInit bool

// NewRootCmd creates a root command.
func NewRootCmd() (*cobra.Command, error) {
	u := cli.NewMainUsage()
	root.SetUsageFunc(u.Func())

	ni := os.Getenv("TANZU_CLI_NO_INIT")
	if ni != "" {
		noInit = true
	}

	// TODO (pbarker): silencing usage for now as we are getting double usage from plugins on errors
	root.SilenceUsage = true

	root.AddCommand(
		pluginCmd,
		initCmd,
		updateCmd,
		versionCmd,
		completionCmd,
		configCmd,
	)

	catalog, err := cli.NewCatalog()
	if err != nil {
		return nil, err
	}
	plugins, err := catalog.List()
	if err != nil {
		return nil, fmt.Errorf("find available plugins: %w", err)
	}

	// check that all plugins in the core distro are installed or do so.
	if !noInit && !catalog.Distro().IsSatisfied(plugins) {
		s := spin.New("%s   initializing")
		s.Start()
		cfg, err := client.GetConfig()
		if err != nil {
			log.Fatal(err)
		}
		repos := cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
		err = catalog.EnsureDistro(repos)
		if err != nil {
			return nil, err
		}
		plugins, err = catalog.List()
		if err != nil {
			return nil, fmt.Errorf("find available plugins: %w", err)
		}
		s.Stop()
	}
	for _, plugin := range plugins {
		root.AddCommand(plugin.Cmd())
	}

	// Flag parsing must be disabled because the root plugin won't know about all flags.
	root.DisableFlagParsing = true

	return root, nil
}

// Execute executes the CLI.
func Execute() error {
	root, err := NewRootCmd()
	if err != nil {
		return err
	}
	return root.Execute()
}
