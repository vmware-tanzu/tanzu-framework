package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	clitest "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cli"
)

var pluginName = "feature"

var descriptor = cli.NewTestFor(pluginName)

func main() {
	retcode := 0

	defer func() { os.Exit(retcode) }()
	defer func() { _ = Cleanup() }()

	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Println(err)
		retcode = 1
		return
	}
	p.Cmd.RunE = test
	if err := p.Execute(); err != nil {
		retcode = 1
		return
	}
}

//nolint:gocritic
func test(c *cobra.Command, _ []string) error {
	m := clitest.NewMain(pluginName, c, Cleanup)
	defer m.Finish()

	// example test

	// testName := clitest.GenerateName()
	//
	// err := m.RunTest(
	// 	"create a feature",
	// 	fmt.Sprintf("feature create -n %s", testName),
	// 	func(t *clitest.Test) error {
	// 		err := t.ExecContainsString("created")
	// 		if err != nil {
	// 			return err
	// 		}
	// 		return nil
	// 	},
	// )
	// if err != nil {
	// 	return err
	// }
	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
