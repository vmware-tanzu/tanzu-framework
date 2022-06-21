// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"time"

	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/signals"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "feature",
	Description: "Operate on features and featuregates",
	Version:     buildinfo.Version,
	Group:       cliv1alpha1.RunCmdGroup,
}

const contextTimeout = 30 * time.Second

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

	ctx := signals.SetupSignalHandler()
	ctx, cancel := context.WithTimeout(ctx, contextTimeout)
	p.AddContext(ctx)

	if err := p.Execute(); err != nil {
		cancel()
		os.Exit(1)
	}
	cancel()
}
