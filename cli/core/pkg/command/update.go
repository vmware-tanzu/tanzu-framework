// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var yesUpdate bool

func init() {
	updateCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	updateCmd.Flags().BoolVarP(&yesUpdate, "yes", "y", false, "force update; skip prompt")
	updateCmd.Flags().StringVarP(&local, "local", "l", "", "path to local repository")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI",
	Annotations: map[string]string{
		"group": string(cliapi.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			return errors.New("CLI self-update is currently not supported. For updating plugins please use the `tanzu plugin sync` command")
		}

		// clean the catalog cache when updating the cli
		if err := cli.CleanCatalogCache(); err != nil {
			log.Debugf("Failed to clean the Plugin descriptors cache %v", err)
		}
		// TODO: cli.ListPlugins is deprecated: Use pluginmanager.AvailablePluginsFromLocalSource or pluginmanager.AvailablePlugins instead
		plugins, err := cli.ListPlugins()
		if err != nil {
			return err
		}

		repos := getRepositories()
		coreRepo, err := repos.Find(cli.CoreName)
		if err != nil {
			return err
		}

		type updateInfo struct {
			version string
			repo    cli.Repository
		}

		updateMap := map[*cliapi.PluginDescriptor]updateInfo{}
		for _, plugin := range plugins {
			if plugin.Name == cli.CoreName {
				continue
			}
			update, repo, version, err := cli.HasPluginUpdateIn(repos, plugin)
			if err != nil {
				log.Warningf("could not find local plugin %q in any remote repositories", plugin.Name)
				continue
			}
			if update {
				updateMap[plugin] = updateInfo{version, repo}
			}
		}

		coreUpdate, coreVersion, err := cli.HasUpdate(coreRepo)
		if err != nil {
			return err
		}

		if len(updateMap) == 0 && !coreUpdate {
			log.Info("everything up to date")
			return nil
		}

		log.Info("the following updates will take place:")
		if coreUpdate {
			fmt.Printf("     %s %s → %s\n", cli.CoreName, buildinfo.Version, coreVersion)
		}
		for plugin, version := range updateMap {
			fmt.Printf("     %s %s → %s\n", plugin.Name, plugin.Version, version)
		}
		// formatting
		fmt.Println()

		if !yesUpdate {
			input := &survey.Input{Message: "would you like to continue? [y/n]"}
			var resp string
			err := survey.AskOne(input, &resp)
			if err != nil {
				return err
			}
			update := strings.ToLower(resp)
			if update != "y" && update != "yes" {
				log.Info("aborting update")
				return nil
			}
		}
		for plugin, info := range updateMap {
			err := cli.InstallPlugin(plugin.Name, info.version, info.repo)
			if err != nil {
				return err
			}
		}

		// update core
		err = cli.Update(coreRepo)
		if err != nil {
			return err
		}

		// formatting
		fmt.Println()
		log.Success("successfully updated CLI")
		return nil
	},
}
