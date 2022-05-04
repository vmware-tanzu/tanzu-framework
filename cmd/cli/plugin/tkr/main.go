// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/spf13/cobra"

	tkrv1alpha1 "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/tkr/v1alpha1"
	tkrv1alpha3 "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/tkr/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "kubernetes-release",
	Description: "Kubernetes release operations",
	Group:       cliv1alpha1.RunCmdGroup,
	Aliases:     []string{"kr", "kubernetes-releases"},
}

var (
	logLevel int32
	logFile  string

	v1alpha1CmdsList = []*cobra.Command{
		tkrv1alpha1.GetTanzuKubernetesRleasesCmd,
		tkrv1alpha1.OsCmd,
		tkrv1alpha1.AvailableUpgradesCmd,
		tkrv1alpha1.ActivateCmd,
		tkrv1alpha1.DeactivateCmd,
	}

	v1alpha3CmdsList = []*cobra.Command{
		tkrv1alpha3.GetTanzuKubernetesRleasesCmd,
		tkrv1alpha3.OsCmd,
		tkrv1alpha3.ActivateCmd,
		tkrv1alpha3.DeactivateCmd,
	}
)

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "v", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path")
	// Add the right set of commands based on the feature flag
	if config.IsFeatureActivated(config.FeatureFlagTKRVersionV1Alpha3) {
		tkrv1alpha3.LogFile = logFile
		tkrv1alpha3.LogLevel = logLevel
		p.AddCommands(v1alpha3CmdsList...)
	} else {
		tkrv1alpha1.LogFile = logFile
		tkrv1alpha1.LogLevel = logLevel
		p.AddCommands(v1alpha1CmdsList...)
	}
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
