// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/fatih/color"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/plugin"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/command"
	component "github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	local       string
	version     string
	forceDelete bool
	target      string
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
		discoverySourceCmd,
	)
	listPluginCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	listPluginCmd.Flags().StringVarP(&local, "local", "l", "", "path to local discovery/distribution source")
	installPluginCmd.Flags().StringVarP(&local, "local", "l", "", "path to local discovery/distribution source")
	installPluginCmd.Flags().StringVarP(&version, "version", "v", cli.VersionLatest, "version of the plugin")
	deletePluginCmd.Flags().BoolVarP(&forceDelete, "yes", "y", false, "delete the plugin without asking for confirmation")

	if config.IsFeatureActivated(cliconfig.FeatureContextCommand) {
		installPluginCmd.Flags().StringVarP(&target, "target", "", "", "target of the plugin")
		upgradePluginCmd.Flags().StringVarP(&target, "target", "", "", "target of the plugin")
		deletePluginCmd.Flags().StringVarP(&target, "target", "", "", "target of the plugin")
		describePluginCmd.Flags().StringVarP(&target, "target", "", "", "target of the plugin")
	}

	command.DeprecateCommand(repoCmd, "")
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage CLI plugins",
	Annotations: map[string]string{
		"group": string(cliapi.SystemCmdGroup),
	},
}

var listPluginCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var availablePlugins []plugin.Discovered
		if local != "" {
			// get absolute local path
			local, err = filepath.Abs(local)
			if err != nil {
				return err
			}
			availablePlugins, err = pluginmanager.AvailablePluginsFromLocalSource(local)
		} else {
			availablePlugins, err = pluginmanager.AvailablePlugins()
		}
		sort.Sort(plugin.DiscoveredSorter(availablePlugins))

		if err != nil {
			return err
		}

		if config.IsFeatureActivated(cliconfig.FeatureContextCommand) && (outputFormat == "" || outputFormat == string(component.TableOutputType)) {
			displayPluginListOutputSplitViewContext(availablePlugins, cmd.OutOrStdout())
		} else {
			displayPluginListOutputListView(availablePlugins, cmd.OutOrStdout())
		}

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
		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			pd, err := pluginmanager.DescribePlugin(pluginName, cliv1alpha1.Target(target))
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

		repo, err := repos.Find(pluginName)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(pluginName)
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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {

			// Invoke install plugin from local source if local files are provided
			if local != "" {
				// get absolute local path
				local, err = filepath.Abs(local)
				if err != nil {
					return err
				}
				err = pluginmanager.InstallPluginsFromLocalSource(pluginName, version, cliv1alpha1.Target(target), local, false)
				if err != nil {
					return err
				}
				if pluginName == cli.AllPlugins {
					log.Successf("successfully installed all plugins")
				} else {
					log.Successf("successfully installed '%s' plugin", pluginName)
				}
				return nil
			}

			// Invoke plugin sync if install all plugins is mentioned
			if pluginName == cli.AllPlugins {
				err = pluginmanager.SyncPlugins()
				if err != nil {
					return err
				}
				log.Successf("successfully installed all plugins")
				return nil
			}

			pluginVersion := version
			if pluginVersion == cli.VersionLatest {
				pluginVersion, err = pluginmanager.GetRecommendedVersionOfPlugin(pluginName, cliv1alpha1.Target(target))
				if err != nil {
					return err
				}
			}

			err = pluginmanager.InstallPlugin(pluginName, pluginVersion, cliv1alpha1.Target(target))
			if err != nil {
				return err
			}
			log.Successf("successfully installed '%s' plugin", pluginName)
			return nil
		}

		repos := getRepositories()

		if pluginName == cli.AllPlugins {
			return cli.InstallAllMulti(repos)
		}
		repo, err := repos.Find(pluginName)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(pluginName)
		if err != nil {
			return err
		}
		if version == cli.VersionLatest {
			version = plugin.FindVersion(repo.VersionSelector())
		}
		err = cli.InstallPlugin(pluginName, version, repo)
		if err != nil {
			return err
		}
		log.Successf("successfully installed %s", pluginName)
		return nil
	},
}

var upgradePluginCmd = &cobra.Command{
	Use:   "upgrade [name]",
	Short: "Upgrade a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			pluginVersion, err := pluginmanager.GetRecommendedVersionOfPlugin(pluginName, cliv1alpha1.Target(target))
			if err != nil {
				return err
			}

			err = pluginmanager.UpgradePlugin(pluginName, pluginVersion, cliv1alpha1.Target(target))
			if err != nil {
				return err
			}
			log.Successf("successfully upgraded plugin '%s' to version '%s'", pluginName, pluginVersion)
			return nil
		}

		repos := getRepositories()
		repo, err := repos.Find(pluginName)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(pluginName)
		if err != nil {
			return err
		}

		versionSelector := repo.VersionSelector()
		err = cli.UpgradePlugin(pluginName, plugin.FindVersion(versionSelector), repo)
		if err != nil {
			return err
		}
		log.Successf("successfully upgraded plugin %s", pluginName)
		return nil
	},
}

var deletePluginCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			deletePluginOptions := pluginmanager.DeletePluginOptions{
				PluginName:  pluginName,
				Target:      cliv1alpha1.Target(target),
				ForceDelete: forceDelete,
			}

			err = pluginmanager.DeletePlugin(deletePluginOptions)
			if err != nil {
				return err
			}

			log.Successf("successfully deleted plugin '%s'", pluginName)
			return nil
		}

		err = cli.DeletePlugin(pluginName)
		if err != nil {
			return err
		}
		log.Successf("successfully deleted plugin %s", pluginName)
		return nil
	},
}

var cleanPluginCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			err = pluginmanager.Clean()
			if err != nil {
				return err
			}
			log.Success("successfully cleaned up all plugins")
			return nil
		}

		err = cli.Clean()
		if err != nil {
			return err
		}
		log.Success("successfully cleaned up all plugins")
		return nil
	},
}

var syncPluginCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			err = pluginmanager.SyncPlugins()
			if err != nil {
				return err
			}
			log.Success("Done")
			return nil
		}
		return errors.Errorf("command is only applicable if `%s` feature is enabled", cliconfig.FeatureContextAwareCLIForPlugins)
	},
}

func getRepositories() *cli.MultiRepo {
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	if local != "" {
		vs := cli.LoadVersionSelector(cfg.ClientOptions.CLI.UnstableVersionSelector)

		opts := []cli.Option{
			cli.WithVersionSelector(vs),
		}

		m := cli.NewMultiRepo()
		n := filepath.Base(local)
		r := cli.NewLocalRepository(n, local, opts...)
		m.AddRepository(r)
		return m
	}

	return cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
}

// getInstalledElseAvailablePluginVersion return installed plugin version if plugin is installed
// if not installed it returns available recommanded plugin version
func getInstalledElseAvailablePluginVersion(p *plugin.Discovered) string {
	installedOrAvailableVersion := p.InstalledVersion
	if installedOrAvailableVersion == "" {
		installedOrAvailableVersion = p.RecommendedVersion
	}
	return installedOrAvailableVersion
}

func displayPluginListOutputListView(availablePlugins []plugin.Discovered, writer io.Writer) {
	var data [][]string
	var output component.OutputWriter

	for index := range availablePlugins {
		data = append(data, []string{availablePlugins[index].Name, availablePlugins[index].Description, availablePlugins[index].Scope,
			availablePlugins[index].Source, getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status})
	}
	output = component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Scope", "Discovery", "Version", "Status")

	for _, row := range data {
		vals := make([]interface{}, len(row))
		for i, val := range row {
			vals[i] = val
		}
		output.AddRow(vals...)
	}
	output.Render()
}

func displayPluginListOutputSplitViewContext(availablePlugins []plugin.Discovered, writer io.Writer) {
	var dataStandalone [][]string
	var outputStandalone component.OutputWriter
	dataContext := make(map[string][][]string)
	outputContext := make(map[string]component.OutputWriter)

	outputStandalone = component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Discovery", "Version", "Status")

	for index := range availablePlugins {
		if availablePlugins[index].Scope == common.PluginScopeStandalone {
			newRow := []string{availablePlugins[index].Name, availablePlugins[index].Description, string(availablePlugins[index].Target),
				availablePlugins[index].Source, getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status}
			dataStandalone = append(dataStandalone, newRow)
		} else {
			newRow := []string{availablePlugins[index].Name, availablePlugins[index].Description, string(availablePlugins[index].Target),
				getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status}
			outputContext[availablePlugins[index].ContextName] = component.NewOutputWriter(writer, outputFormat, "Name", "Description", "Target", "Version", "Status")
			data := dataContext[availablePlugins[index].ContextName]
			data = append(data, newRow)
			dataContext[availablePlugins[index].ContextName] = data
		}
	}

	addDataToOutputWriter := func(output component.OutputWriter, data [][]string) {
		for _, row := range data {
			vals := make([]interface{}, len(row))
			for i, val := range row {
				vals[i] = val
			}
			output.AddRow(vals...)
		}
	}

	cyanBold := color.New(color.FgCyan).Add(color.Bold)
	cyanBoldItalic := color.New(color.FgCyan).Add(color.Bold, color.Italic)

	_, _ = cyanBold.Println("Standalone Plugins")
	addDataToOutputWriter(outputStandalone, dataStandalone)
	outputStandalone.Render()

	for context, writer := range outputContext {
		fmt.Println("")
		_, _ = cyanBold.Println("Context Plugins: ", cyanBoldItalic.Sprintf(context))
		data := dataContext[context]
		addDataToOutputWriter(writer, data)
		writer.Render()
	}
}
