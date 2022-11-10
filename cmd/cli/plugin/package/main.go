// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	capdiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/package/kctrl"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

var descriptor = cliapi.PluginDescriptor{
	Name:                "package",
	Description:         "Tanzu package management",
	Group:               cliapi.RunCmdGroup,
	Version:             buildinfo.Version,
	BuildSHA:            buildinfo.SHA,
	DefaultFeatureFlags: DefaultFeatureFlagsForPackagePlugin,
}

var logLevel int32
var outputFormat string
var kubeConfig string

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	if config.IsFeatureActivated(FeatureFlagPackagePluginKctrlCommandTree) {
		kctrl.Invoke(p)
		if err := p.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")

	p.AddCommands(
		repositoryCmd,
		packageInstallCmd,
		packageAvailableCmd,
		packageInstalledCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

// getOutputFormat gets the desired output format for package commands that need the ListTable format
// for its output.
func getOutputFormat() string {
	format := outputFormat
	if format != string(component.JSONOutputType) && format != string(component.YAMLOutputType) {
		// For table output, we want to force the list table format for this part
		format = string(component.ListTableOutputType)
	}
	return format
}

func isPackagingAPIAvailable(kubeCfgPath string) (bool, error) {
	cfg, err := kappclient.GetKubeConfig(kubeCfgPath)
	if err != nil {
		return false, err
	}
	clusterQueryClient, err := capdiscovery.NewClusterQueryClientForConfig(cfg)
	if err != nil {
		log.Error(err, "failed to create a new instance of the cluster query builder")
		return false, err
	}

	apiGroup1 := capdiscovery.Group("packageMetadateAPIQuery", packagedatamodel.DataPackagingAPIName).WithVersions(packagedatamodel.PackagingAPIVersion).WithResource("packagemetadatas")
	apiGroup2 := capdiscovery.Group("packageAPIQuery", packagedatamodel.DataPackagingAPIName).WithVersions(packagedatamodel.PackagingAPIVersion).WithResource("packages")
	apiGroup3 := capdiscovery.Group("packageRepositoryAPIQuery", packagedatamodel.PackagingAPIName).WithVersions(packagedatamodel.PackagingAPIVersion).WithResource("packagerepositories")
	apiGroup4 := capdiscovery.Group("packageInstallAPIQuery", packagedatamodel.PackagingAPIName).WithVersions(packagedatamodel.PackagingAPIVersion).WithResource("packageinstalls")

	return clusterQueryClient.Query(apiGroup1, apiGroup2, apiGroup3, apiGroup4).Execute()
}
