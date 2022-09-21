// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	capdiscovery "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
)

var descriptor = cliapi.PluginDescriptor{
	Name:        "secret",
	Description: "Tanzu secret management",
	Group:       cliapi.RunCmdGroup,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

var (
	logLevel     int32
	outputFormat string
	kubeConfig   string
)

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err, "failed to create a new instance of the plugin")
	}

	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")

	p.AddCommands(
		registrySecretCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func isSecretGenAPIAvailable(kubeCfgPath string) (bool, error) {
	cfg, err := kappclient.GetKubeConfig(kubeCfgPath)
	if err != nil {
		return false, err
	}
	clusterQueryClient, err := capdiscovery.NewClusterQueryClientForConfig(cfg)
	if err != nil {
		log.Error(err, "failed to create a new instance of the cluster query builder")
		return false, err
	}

	apiGroup := capdiscovery.Group("secretGenAPIQuery", packagedatamodel.SecretGenAPIName).WithVersions(packagedatamodel.SecretGenAPIVersion).WithResource("secretexports")
	return clusterQueryClient.Query(apiGroup).Execute()
}
