package core

import (
	"fmt"
	"path/filepath"

	"github.com/aunum/log"
	"golang.org/x/mod/semver"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

var (
	local []string
)

func init() {
	pluginCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	pluginCmd.AddCommand(
		listPluginCmd,
		installPluginCmd,
		upgradePluginCmd,
		deletePluginCmd,
		repoCmd,
		cleanPluginCmd,
	)
	pluginCmd.PersistentFlags().StringSliceVarP(&local, "local", "l", []string{}, "path to local repository")
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
				data = append(data, []string{plugin.Name, plugin.Version, plugin.Description, repo, status})
			}
		}

		table := component.NewTableWriter("Name", "Version", "Description", "Repository", "Status")

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
