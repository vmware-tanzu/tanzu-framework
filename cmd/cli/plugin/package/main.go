// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"
	"github.com/cppforlife/go-cli-ui/ui"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	kctrlcmd "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd"
	kctrlcmdcore "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd/core"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	capdiscovery "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "package",
	Description: "Tanzu package management",
	Group:       cliv1alpha1.RunCmdGroup,
}

var logLevel int32
var outputFormat string
var kubeConfig string

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	err = nonExitingMain(p)
	if err != nil {
		os.Exit(1)
	}
}

func nonExitingMain(p *plugin.Plugin) error {
	confUI := ui.NewConfUI(ui.NewNoopLogger())

	defer confUI.Flush()

	kctrlcmd.AttachKctrlPackageCommandTree(p.Cmd, confUI, kctrlcmdcore.PackageCommandTreeOpts{BinaryName: "tanzu", PositionalArgs: true,
		Color: false, JSON: false})

	if err := p.Execute(); err != nil {
		return err
	}

	return nil
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

	apiGroup1 := capdiscovery.Group("packageMetadateAPIQuery", tkgpackagedatamodel.DataPackagingAPIName).WithVersions(tkgpackagedatamodel.PackagingAPIVersion).WithResource("packagemetadatas")
	apiGroup2 := capdiscovery.Group("packageAPIQuery", tkgpackagedatamodel.DataPackagingAPIName).WithVersions(tkgpackagedatamodel.PackagingAPIVersion).WithResource("packages")
	apiGroup3 := capdiscovery.Group("packageRepositoryAPIQuery", tkgpackagedatamodel.PackagingAPIName).WithVersions(tkgpackagedatamodel.PackagingAPIVersion).WithResource("packagerepositories")
	apiGroup4 := capdiscovery.Group("packageInstallAPIQuery", tkgpackagedatamodel.PackagingAPIName).WithVersions(tkgpackagedatamodel.PackagingAPIVersion).WithResource("packageinstalls")

	return clusterQueryClient.Query(apiGroup1, apiGroup2, apiGroup3, apiGroup4).Execute()
}
