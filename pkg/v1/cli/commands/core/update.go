package core

import (
	"fmt"
	"strings"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/AlecAivazis/survey/v2"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
)

var yesUpdate bool

func init() {
	updateCmd.SetUsageFunc(plugin.UsageFunc)
	updateCmd.Flags().BoolVarP(&yesUpdate, "yes", "y", false, "force update; skip prompt")
	updateCmd.Flags().StringVarP(&local, "local", "l", "", "path to local repository")
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

		repo := getRepository()

		updateMap := map[cli.PluginDescriptor]string{}
		for _, plugin := range plugins {
			update, version, err := plugin.HasUpdate(repo)
			if err != nil {
				return err
			}
			if update {
				updateMap[plugin] = version
			}
		}
		coreUpdate, coreVersion, err := cli.HasUpdate(repo)
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
		for plugin, version := range updateMap {
			err := catalog.Install(plugin.Name, version, repo)
			if err != nil {
				return err
			}
		}

		// update core
		err = cli.Update(repo)
		if err != nil {
			return err
		}

		// formatting
		fmt.Println()
		log.Success("successfully updated CLI")
		return nil
	},
}
