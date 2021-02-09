// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "test",
	Description: "Test the CLI",
	Group:       cli.AdminCmdGroup,
}

var local []string

func init() {
	fetchCmd.PersistentFlags().StringSliceVarP(&local, "local", "l", []string{}, "paths to local repository")
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
	c, err := cli.NewCatalog()
	if err != nil {
		log.Fatal(err)
	}
	descs, err := c.List("test")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range descs {
		pluginsCmd.AddCommand(d.TestCmd())
	}

	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch the plugin tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cli.NewCatalog()
		if err != nil {
			log.Fatal(err)
		}
		repos := getRepositories()
		err = c.EnsureTests(repos, "test")
		if err != nil {
			log.Fatal(err)
		}
		return nil
	},
}

var pluginsCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin tests",
}

func getRepositories() *cli.MultiRepo {
	if len(local) != 0 {
		m := cli.NewMultiRepo()
		for _, l := range local {
			n := filepath.Base(l)
			r := cli.NewLocalRepository(n, l)
			m.AddRepository(r)
		}
		return m
	}
	cfg, err := client.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	return cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
}
