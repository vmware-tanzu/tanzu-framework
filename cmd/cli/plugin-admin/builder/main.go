package main

import (
	"github.com/spf13/cobra"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "builder",
	Description: "Build Tanzu services",
	Version:     "v0.0.1",
	Group:       cli.AdminCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		initCmd,
	)
	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a repository",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}
