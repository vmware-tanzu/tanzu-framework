// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/pluginmanager"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
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
		syncPluginCmd,
	)
	listPluginCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	pluginCmd.PersistentFlags().StringSliceVarP(&local, "local", "l", []string{}, "path to local repository")
	installPluginCmd.Flags().StringVarP(&version, "version", "v", cli.VersionLatest, "version of the plugin")
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage CLI plugins",
	Annotations: map[string]string{
		"group": string(cliv1alpha1.SystemCmdGroup),
	},
}

var listPluginCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			availablePlugins, err := pluginmanager.AvailablePlugins(serverName)
			if err != nil {
				return err
			}

			data := [][]string{}
			for _, p := range availablePlugins {
				data = append(data, []string{p.Name, p.Description, p.Scope,
					p.Source, p.RecommendedVersion, p.Status})
			}

			output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Name", "Description", "Scope", "Discovery", "Version", "Status")
			for _, row := range data {
				vals := make([]interface{}, len(row))
				for i, val := range row {
					vals[i] = val
				}
				output.AddRow(vals...)
			}
			output.Render()

			return nil
		}

		descriptors, err := cli.ListPlugins()
		if err != nil {
			return err
		}

		repos := getRepositories()
		plugins, err := repos.ListPlugins()

		// Failure to query plugin metadata from remote repositories should not
		// prevent display of information about plugins already installed.
		if err != nil {
			log.Warningf("Unable to query remote plugin repositories : %v", err)
		}

		data := [][]string{}
		for repoName, descs := range plugins {
			for _, plugin := range descs {
				if plugin.Name == cli.CoreName {
					continue
				}

				status := "not installed"
				var currentVersion string

				repo, err := repos.GetRepository(repoName)
				if err != nil {
					return err
				}

				versionSelector := repo.VersionSelector()
				latestVersion := plugin.FindVersion(versionSelector)
				for _, desc := range descriptors {
					if plugin.Name != desc.Name {
						continue
					}
					status = "installed"
					currentVersion = desc.Version
					compared := semver.Compare(latestVersion, desc.Version)
					if compared == 1 {
						status = "upgrade available"
					}
				}
				data = append(data, []string{plugin.Name, latestVersion, plugin.Description, repoName, currentVersion, status})
			}
		}

		// show plugins installed locally but not found in repositories
		// failure to query the remote repository will degenerate the plugin list
		// to this view as well
		for _, desc := range descriptors {
			var exists bool
			for _, d := range data {
				if desc.Name == d[0] {
					exists = true
					break
				}
			}
			if !exists {
				data = append(data, []string{desc.Name, "", desc.Description, "", desc.Version, "installed"})
			}
		}

		// sort plugins based on their names
		sort.SliceStable(data, func(i, j int) bool {
			return strings.ToLower(data[i][0]) < strings.ToLower(data[j][0])
		})

		output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Name", "Latest Version", "Description", "Repository", "Version", "Status")
		for _, row := range data {
			vals := make([]interface{}, len(row))
			for i, val := range row {
				vals[i] = val
			}
			output.AddRow(vals...)
		}
		output.Render()
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

		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}
			pd, err := pluginmanager.DescribePlugin(serverName, name)
			if err != nil {
				return err
			}

			b, err := yaml.Marshal(pd)
			if err != nil {
				return errors.Wrap(err, "could not marshal plugin")
			}
			fmt.Println(string(b))

			return nil
		}

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

		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			pluginVersion := version

			if pluginVersion == cli.VersionLatest {
				pluginVersion, err = pluginmanager.GetRecommendedVersionOfPlugin(serverName, name)
				if err != nil {
					return err
				}
			}

			err = pluginmanager.InstallPlugin(serverName, name, pluginVersion)
			if err != nil {
				return err
			}
			log.Successf("successfully installed '%s' plugin", name)
			return nil
		}

		repos := getRepositories()

		if name == cli.AllPlugins {
			return cli.InstallAllMulti(repos)
		}
		repo, err := repos.Find(name)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(name)
		if err != nil {
			return err
		}
		if version == cli.VersionLatest {
			version = plugin.FindVersion(repo.VersionSelector())
		}
		err = cli.InstallPlugin(name, version, repo)
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

		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			pluginVersion, err := pluginmanager.GetRecommendedVersionOfPlugin(serverName, name)
			if err != nil {
				return err
			}

			err = pluginmanager.UpgradePlugin(serverName, name, pluginVersion)
			if err != nil {
				return err
			}
			log.Successf("successfully upgraded plugin '%s' to version '%s'", name, pluginVersion)
			return nil
		}

		repos := getRepositories()
		repo, err := repos.Find(name)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(name)
		if err != nil {
			return err
		}

		versionSelector := repo.VersionSelector()
		err = cli.UpgradePlugin(name, plugin.FindVersion(versionSelector), repo)
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

		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			err = pluginmanager.DeletePlugin(serverName, name)
			if err != nil {
				return err
			}

			log.Successf("successfully deleted plugin '%s'", name)
			return nil
		}

		err = cli.DeletePlugin(name)

		return
	},
}

var cleanPluginCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			return pluginmanager.Clean()
		}
		return cli.Clean()
	},
}

var syncPluginCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}
			err = pluginmanager.SyncPlugins(serverName)
			if err != nil {
				return err
			}
			log.Success("Done")
			return nil
		}
		return errors.Errorf("command is only applicable if `%s` feature is enabled", config.FeatureContextAwareDiscovery)
	},
}

func getRepositories() *cli.MultiRepo {
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(local) != 0 {
		vs := cli.LoadVersionSelector(cfg.ClientOptions.CLI.UnstableVersionSelector)

		opts := []cli.Option{
			cli.WithVersionSelector(vs),
		}

		m := cli.NewMultiRepo()
		for _, l := range local {
			n := filepath.Base(l)
			r := cli.NewLocalRepository(n, l, opts...)
			m.AddRepository(r)
		}
		return m
	}

	return cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
}
