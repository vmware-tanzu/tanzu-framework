// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/isolated-cluster/imageop"
)

var descriptor = cliapi.PluginDescriptor{
	Name:        "isolated-cluster",
	Description: "Prepopulating images/bundle for internet-restricted environments",
	Group:       cliapi.RunCmdGroup,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

var logLevel int32
var logFile string

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "v", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path")
	p.Cmd.SilenceUsage = true
	p.AddCommands(
		imageop.PublishImagestotarCmd,
		imageop.PublishImagesfromtarCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
