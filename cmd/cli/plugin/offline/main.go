// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "offline",
	Description: "Prepares your installation for an air-gapped platform",
	Group:       cliv1alpha1.RunCmdGroup,
}

var logLevel int32

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "v", 0, "Number for the log level verbosity (0-9)")
	p.AddCommands(prepareCmd)
	if err = p.Execute(); err != nil {
		os.Exit(1)
	}
}
