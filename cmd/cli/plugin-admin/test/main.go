package main

import (
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "test",
	Description: "Test the CLI",
	Version:     "v0.0.1",
	Group:       cli.AdminCmdGroup,
}

var local string

func init() {
	fetchCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "path to local repository")
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
	if local != "" {
		return cli.NewMultiRepo(cli.NewLocalRepository("local", local))
	}
	cfg, err := client.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	return cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
}
