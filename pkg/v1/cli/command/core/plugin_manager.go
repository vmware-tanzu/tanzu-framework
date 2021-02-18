// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"path/filepath"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

var (
	local   []string
	version string
)

func init() {
	pluginCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	pluginCmd.AddCommand(
		listPluginCmd,
		installPluginCmd,
		upgradePluginCmd,
		describePluginCmd,
		deletePluginCmd,
		repoCmd,
		cleanPluginCmd,
	)
	pluginCmd.PersistentFlags().StringSliceVarP(&local, "local", "l", []string{}, "path to local repository")
	installPluginCmd.Flags().StringVarP(&version, "version", "v", cli.VersionLatest, "version of the plugin")
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage CLI plugins",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
}

var listPluginCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}
		descriptors, err := catalog.List()
		if err != nil {
			return err
		}

		repo := getRepositories()
		plugins, err := repo.ListPlugins()
		if err != nil {
			return err
		}

		data := [][]string{}
		for repo, descs := range plugins {
			for _, plugin := range descs {

				status := "not installed"
				var currentVersion string
				for _, desc := range descriptors {
					if plugin.Name != desc.Name {
						continue
					}
					compared := semver.Compare(plugin.VersionLatest(), desc.Version)
					if compared == 1 {
						status = "upgrade available"
						continue
					}
					status = "installed"
					currentVersion = desc.Version
				}
				data = append(data, []string{plugin.Name, plugin.VersionLatest(), plugin.Description, repo, currentVersion, status})
			}
		}

		table := component.NewTableWriter("Name", "Latest Version", "Description", "Repository", "Version", "Status")

		for _, v := range data {
			table.Append(v)
		}
		table.Render()
		return nil
	},
}

var describePluginCmd = &cobra.Command{
	Use:   "describe [name]",
	Short: "Describe a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		name := args[0]

		repos := getRepositories()

		repo, err := repos.Find(name)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(name)
		if err != nil {
			return err
		}

		b, err := yaml.Marshal(plugin)
		if err != nil {
			return errors.Wrap(err, "could not marshal plugin")
		}
		fmt.Println(string(b))
		return
	},
}

var installPluginCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		name := args[0]

		repos := getRepositories()

		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}

		if name == cli.AllPlugins {
			return catalog.InstallAllMulti(repos)
		}
		repo, err := repos.Find(name)
		if err != nil {
			return err
		}
		err = catalog.Install(name, version, repo)
		if err != nil {
			return
		}
		log.Successf("successfully installed %s", name)
		return
	},
}

var upgradePluginCmd = &cobra.Command{
	Use:   "upgrade [name]",
	Short: "Upgrade a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		name := args[0]

		repos := getRepositories()
		repo, err := repos.Find(name)
		if err != nil {
			return err
		}
		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}

		err = catalog.Install(name, cli.VersionLatest, repo)
		return
	},
}

var deletePluginCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		name := args[0]

		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}

		err = catalog.Delete(name)

		return
	},
}

var cleanPluginCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}
		return catalog.Clean()
	},
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
