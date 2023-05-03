// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/aunum/log"

	plugintypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo"
)

var descriptor = plugin.PluginDescriptor{
	Name:        "codegen",
	Description: "Tanzu code generation tool",
	Target:      plugintypes.TargetGlobal,
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
