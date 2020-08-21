package main

import (
	"os"

	"github.com/aunum/log"

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
	p.AddCommands()
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
