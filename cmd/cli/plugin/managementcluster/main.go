// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var (
	// BuildEdition is the edition the CLI was built for.
	BuildEdition string
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "management-cluster",
	Description: "Kubernetes management cluster operations",
	Version:     cli.BuildVersion,
	BuildSHA:    "",
	Group:       cliv1alpha1.RunCmdGroup,
	Aliases:     []string{"mc", "management-clusters"},
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
