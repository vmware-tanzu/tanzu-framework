// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/codegen"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "codegen",
	Description: "Tanzu code generation tool",
	Group:       cliv1alpha1.AdminCmdGroup,
	Version:     buildinfo.Version,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.AddCommands(
		codegen.GenerateCmd,
	)

	if err := p.Execute(); err != nil {
		log.Fatal(err)
	}
}
