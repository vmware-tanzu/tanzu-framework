package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
)

var descriptor = cli.NewTestFor("cluster")

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.Cmd.RunE = test
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func test(c *cobra.Command, _ []string) error {
	m := clitest.NewMain("cluster", c, Cleanup)
	defer m.Finish()

	err := m.RunTest(
		"get a cluster",
		"cluster get",
		func(t *clitest.Test) error {
			err := t.ExecContainsString("in progress...")
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
