// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	pluginrtbuildinfo "github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

var descriptor = cliapi.PluginDescriptor{
	Name:        "pinniped-auth",
	Description: "Pinniped authentication operations (usually not directly invoked)",
	Group:       cliapi.RunCmdGroup,
	Hidden:      true,
	Aliases:     []string{"pa", "pinniped-auths"},
	// TODO: When plugins have their own buildInfo, we need to update "Version" and "BuildSHA"
	// 		to plugin own versions instead of plugin runtime version
	Version:              pluginrtbuildinfo.Version,
	BuildSHA:             pluginrtbuildinfo.SHA,
	PluginRuntimeVersion: pluginrtbuildinfo.Version,
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
