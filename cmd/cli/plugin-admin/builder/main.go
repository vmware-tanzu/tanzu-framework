// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
	Group:       cli.AdminCmdGroup,
	Version:     cli.BuildVersion,
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
