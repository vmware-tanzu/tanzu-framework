// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"time"

	"github.com/aunum/log"

	plugintypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo"
)

var descriptor = plugin.PluginDescriptor{
	Name:        "feature",
	Description: "Operate on features and featuregates",
	Target:      plugintypes.TargetK8s,
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
