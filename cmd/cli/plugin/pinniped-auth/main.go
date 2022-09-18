// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
)

var descriptor = cliapi.PluginDescriptor{
	Name:        "pinniped-auth",
	Description: "Pinniped authentication operations (usually not directly invoked)",
	Group:       cliapi.RunCmdGroup,
	Hidden:      true,
	Aliases:     []string{"pa", "pinniped-auths"},
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		loginOIDCCommand(getPinnipedCLICmd),
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
