// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/cmd"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/kubernetes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"

	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

var descriptor = cliapi.PluginDescriptor{
	Name:        "telemetry",
	Description: "configure cluster-wide settings for vmware tanzu telemetry",
	Group:       cliapi.RunCmdGroup,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	out := component.NewOutputWriter(os.Stdout, string(component.YAMLOutputType))

	sc := cmd.NewStatusCmd(kubernetes.GetDynamicClient, out)
	uc := cmd.NewUpdateCmd(kubernetes.GetDynamicClient, out)

	p.AddCommands(
		sc.Cmd,
		uc.Cmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
