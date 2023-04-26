// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-cli/pkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
)

var descriptor = plugin.PluginDescriptor{
	Name:        "codegen",
	Description: "Tanzu code generation tool",
	Group:       plugin.AdminCmdGroup,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.AddCommands(
		GenerateCmd,
	)

	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}
