package core

import (
	"fmt"

	"github.com/aunum/log"
	"golang.org/x/mod/semver"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

var (
	local string
)

func init() {
	pluginCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	pluginCmd.AddCommand(
		listPluginCmd,
		installPluginCmd,
		upgradePluginCmd,
		deletePluginCmd,
	)
	pluginCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "path to local repository")
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

		repo := getRepository()
		plugins, err := repo.List()
		if err != nil {
			return err
		}

		data := make([][]string, len(plugins))
		for i, plugin := range plugins {
			status := "not installed"
			for _, desc := range descriptors {
				if plugin.Name != desc.Name {
					continue
				}
				compared := semver.Compare(plugin.Version, desc.Version)
				if compared == 1 {
					status = "upgrade available"
					continue
				}
				status = "installed"
			}
			data[i] = []string{plugin.Name, plugin.Version, plugin.Description, status}
		}

		table := cli.NewTableWriter("Name", "Version", "Description", "Status")

		for _, v := range data {
			table.Append(v)
		}
		table.Render()
		return nil

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

		repo := getRepository()

		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}

		if name == cli.AllPlugins {
			return catalog.InstallAll(repo)
		}
		err = catalog.Install(name, cli.VersionLatest, repo)
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

		repo := getRepository()
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

func getRepository() cli.Repository {
	if local != "" {
		return cli.NewLocalRepository(local)
	}
	return cli.NewDefaultRepository()
}
