// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/builder/command"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "builder",
	Description: "Build Tanzu components",
	Group:       cliv1alpha1.AdminCmdGroup,
	Version:     cli.BuildVersion,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.AddCommands(
		command.CLICmd,
		command.NewInitCmd(),
	)
	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}
