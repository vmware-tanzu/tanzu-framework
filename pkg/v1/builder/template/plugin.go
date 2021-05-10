// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

// PluginReadMe target
var PluginReadMe = Target{
	Filepath: "cmd/plugin/{{ .PluginName }}/README.md",
	Template: `# {{ .PluginName}}
## Summary
###### [provide a high-level summary of the feature functionality supported by the plugin]

## Overview
###### [outline how the plugin works - boxes and lines diagrams are helpful here]

## Installation
###### [describe any unique or nuanced installation requirements for the plugin]

## Usage
###### [at minimum, provide the output from --help - providing example command output and/or more detailed explanations of commands may also be valuable]

## Documentation
###### [include, or provide links to, additional resources that users or contributors may find useful here]

## Versioning
###### [describe how this plugin follows, or the degree to which it follows or doesn’t follow semver]

## Contribution
###### [describe whether/how/where issues/PR’s should be submitted]

## Development
###### [describe steps to clone, test, build/install locally, etc..]

## License
###### [name and link to the project this plugin is licensed under]`,
}

// PluginMain target
// TODO (pbarker): proper logging
var PluginMain = Target{
	Filepath: "cmd/plugin/{{ .PluginName | ToLower }}/main.go",
	Template: `package main

import (
	"os"

	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "{{ .PluginName | ToLower }}",
	Description: "{{ .Description | ToLower }}",
	Version:     "v0.0.1",
	Group:       cliv1alpha1.ManageCmdGroup, // set group
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
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
)

var pluginName = "{{ .PluginName | ToLower }}"

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
	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
	`,
}
