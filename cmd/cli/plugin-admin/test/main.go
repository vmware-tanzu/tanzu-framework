// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"path/filepath"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "test",
	Description: "Test the CLI",
	Group:       cliv1alpha1.AdminCmdGroup,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
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
	descs, err := cli.ListPlugins("test")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range descs {
		pluginsCmd.AddCommand(cli.TestCmd(d))
	}

	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch the plugin tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		repos := getRepositories()
		err := cli.EnsureTests(repos, "test")
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
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	return cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
}
