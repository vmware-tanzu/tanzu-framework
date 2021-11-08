// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
)

var (
	// BuildEdition is the edition the CLI was built for.
	BuildEdition string
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:            "management-cluster",
	Description:     "Kubernetes management cluster operations",
	Version:         buildinfo.Version,
	BuildSHA:        "",
	Group:           cliv1alpha1.RunCmdGroup,
	Aliases:         []string{"mc", "management-clusters"},
	PostInstallHook: postInstallHook,
}

var (
	logLevel     int32
	logFile      string
	outputFormat string
)

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
	)

	if err = p.Execute(); err != nil {
		os.Exit(1)
	}
}
