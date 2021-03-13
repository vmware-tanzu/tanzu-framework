// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"strings"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/AlecAivazis/survey/v2"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

var yesUpdate bool

func init() {
	updateCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	updateCmd.Flags().BoolVarP(&yesUpdate, "yes", "y", false, "force update; skip prompt")
	updateCmd.Flags().StringSliceVarP(&local, "local", "l", []string{}, "path to local repository")
	updateCmd.Flags().BoolVarP(&includeUnstable, "include-unstable", "u", false, "include unstable versions of the plugins")

	updateCmd.Flags().MarkHidden("include-unstable") //nolint
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}
		plugins, err := catalog.List()
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

		versionSelector := cli.DefaultVersionSelector
		if includeUnstable {
			versionSelector = cli.SelectVersionAny
		}

		updateMap := map[*cli.PluginDescriptor]updateInfo{}
		for _, plugin := range plugins {
			if plugin.Name == cli.CoreName {
				continue
			}
			update, repo, version, err := plugin.HasUpdateIn(repos, versionSelector)
			if err != nil {
				log.Warningf("could not find local plugin %q in any remote repositories", plugin.Name)
				continue
			}
			if update {
				updateMap[plugin] = updateInfo{version, repo}
			}
		}

		coreUpdate, coreVersion, err := cli.HasUpdate(coreRepo, versionSelector)
		if err != nil {
			return err
		}

		if len(updateMap) == 0 && !coreUpdate {
			log.Info("everything up to date")
			return nil
		}

		log.Info("the following updates will take place:")
		if coreUpdate {
			fmt.Printf("     %s %s → %s\n", cli.CoreName, cli.BuildVersion, coreVersion)
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
			err := catalog.Install(plugin.Name, info.version, info.repo)
			if err != nil {
				return err
			}
		}

		// update core
		err = cli.Update(coreRepo, versionSelector)
		if err != nil {
			return err
		}

		// formatting
		fmt.Println()
		log.Success("successfully updated CLI")
		return nil
	},
}
