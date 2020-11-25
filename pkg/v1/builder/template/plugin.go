package template

// PluginReadMe target
var PluginReadMe = Target{
	Filepath: "cmd/plugin/{{ .PluginName }}/README.md",
	Template: `# {{ .PluginName}} `,
}

// PluginMain target
// TODO (pbarker): proper logging
var PluginMain = Target{
	Filepath: "cmd/plugin/{{ .PluginName | ToLower }}/main.go",
	Template: `package main

import (
	"os"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "{{ .PluginName | ToLower }}",
	Description: "",  // provide a description
	Version:     "v0.0.1",
	Group:       cli.ManageCmdGroup, // set group
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		// Add commands
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
	`,
}

// PluginTest target
var PluginTest = Target{
	Filepath: "cmd/plugin/{{ .PluginName }}/test/main.go",
	Template: `package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/test/cmp"
)

var pluginName = "{{ .PluginName | ToLower }}"

var descriptor = cli.NewTestFor(pluginName)

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
	m := clitest.NewMain(pluginName, c, Cleanup)
	defer m.Finish()

	// example test

	// testName := clitest.GenerateName()
	// defCmp := &cmp.DefinedComparer{}
	//
	// err := m.RunTest(
	// 	"create a {{ .PluginName | ToLower }}",
	// 	fmt.Sprintf("{{ .PluginName | ToLower }} create -n %s", testName),
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
	`,
}
