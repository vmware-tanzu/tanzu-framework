// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	tkrv1alpha1 "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/tkr/v1alpha1"
	tkrv1alpha3 "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/tkr/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"

	"github.com/aunum/log"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

var descriptor = cliapi.PluginDescriptor{
	Name:                "kubernetes-release",
	Description:         "Kubernetes release operations",
	Group:               cliapi.RunCmdGroup,
	Aliases:             []string{"kr", "kubernetes-releases"},
	Version:             buildinfo.Version,
	BuildSHA:            buildinfo.SHA,
	DefaultFeatureFlags: DefaultFeatureFlagsForTKRPlugin,
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

	// Add the right set of commands based on the TKR API version
	if isTKRAPIVersionV1Alpha3() {
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

// isTKRAPIVersionV1Alpha3 determines the TKR API version based on the management-cluster feature-gate
func isTKRAPIVersionV1Alpha3() bool {
	// If feature-flag is activated return true
	if config.IsFeatureActivated(FeatureFlagTKRVersionV1Alpha3) {
		return true
	}

	// else check the feature-gate on the management-cluster to determine the TKR API version
	// if clusterclass feature-gate is enabled on the cluster, we can assume that TKR API v1alpha3 is available
	server, err := config.GetCurrentServer()
	if err != nil || server.IsGlobal() {
		return false
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	fgHelper := tkgctl.NewFeatureGateHelper(&clusterClientOptions, server.ManagementClusterOpts.Context, server.ManagementClusterOpts.Path)
	activated, _ := fgHelper.FeatureActivatedInNamespace(context.Background(), constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
	return activated
}
