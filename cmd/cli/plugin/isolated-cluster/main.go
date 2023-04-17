// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/isolated-cluster/imagepullop"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/isolated-cluster/imagepushop"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo"
)

var descriptor = plugin.PluginDescriptor{
	Name:        "isolated-cluster",
	Description: "Prepopulating images/bundle for internet-restricted environments",
	Group:       plugin.RunCmdGroup,
	Target:      types.TargetGlobal,
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
		imagepullop.PublishImagestotarCmd,
		imagepushop.PublishImagesfromtarCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
