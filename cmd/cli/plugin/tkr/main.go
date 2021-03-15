// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "kubernetes-release",
	Description: "Kubernetes release operations",
	Group:       cli.RunCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		getTanzuKubernetesRleasesCmd,
		osCmd,
		availableUpgradesCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
