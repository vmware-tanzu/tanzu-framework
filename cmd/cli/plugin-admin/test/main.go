package main

import (
	"os"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "test",
	Description: "Test the CLI",
	Version:     "v0.0.1",
	Group:       cli.AdminCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	c, err := cli.NewCatalog()
	if err != nil {
		log.Fatal(err)
	}
	err = c.EnsureTests(cli.DefaultMultiRepo)
	if err != nil {
		log.Fatal(err)
	}
	descs, err := c.List()
	if err != nil {
		log.Fatal(err)
	}
	for _, desc := range descs {
		p.AddCommands(desc.TestCmd())
	}
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
