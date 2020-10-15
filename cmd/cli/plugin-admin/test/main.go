package main

import (
	"os"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "test",
	Description: "Test CLI",
	Version:     "v0.0.1",
	Group:       cli.AdminCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		testAllCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

var testAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Test all plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := cli.DefaultMultiRepo
		pm, err := m.ListPlugins()
		if err != nil {
			return err
		}
		for _, plugins := range pm {
			for _, plugin := range plugins {

			}
		}

		return nil
	},
}
