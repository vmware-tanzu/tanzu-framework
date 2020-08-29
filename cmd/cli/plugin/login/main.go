package main

import (
	"os"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "login",
	Description: "Login to the platform",
	Version:     "v0.0.1",
	Group:       cli.SystemCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		globalLoginCmd,
		regionLoginCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

var globalLoginCmd = &cobra.Command{
	Use:   "global",
	Short: "login to the global control plane",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var regionLoginCmd = &cobra.Command{
	Use:   "region",
	Short: "login to a regional control plane",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
