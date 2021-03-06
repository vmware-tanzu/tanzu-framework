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
	Name:        "management-cluster",
	Description: "Kubernetes management cluster operations",
	Version:     cli.BuildVersion,
	BuildSHA:    "",
	Group:       cli.RunCmdGroup,
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

	p.AddCommands(
		createCmd,
		deleteRegionCmd,
		upgradeRegionCmd,
		credentialsCmd,
		ceipCmd,
		getClusterCmd,
		permissionsCmd,
		importCmd,
		clusterKubeconfigCmd,
		registerCmd,
	)

	if err = p.Execute(); err != nil {
		os.Exit(1)
	}
}
