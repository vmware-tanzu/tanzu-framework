// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"time"

	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-cli/pkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
)

var descriptor = plugin.PluginDescriptor{
	Name:        "feature",
	Description: "Operate on features and featuregates",
	Version:     buildinfo.Version,
	Group:       plugin.RunCmdGroup,
	BuildSHA:    buildinfo.SHA,
}

const contextTimeout = 300 * time.Second

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.AddCommands(
		FeatureListCmd,
		FeatureActivateCmd,
		FeatureDeactivateCmd,
	)

	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
