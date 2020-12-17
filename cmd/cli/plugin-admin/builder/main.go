package main

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/builder/command"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "builder",
	Description: "Build Tanzu components",
	Version:     "v0.0.1",
	Group:       cli.AdminCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		command.CLICmd,
		command.InitCmd,
	)
	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}
