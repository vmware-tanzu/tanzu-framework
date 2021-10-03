// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	capdiscovery "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "imagepullsecret",
	Description: "Manage image pull secret operations. Image pull secrets enable the package and package repository consumers to authenticate to private registries.",
	Group:       cliv1alpha1.RunCmdGroup,
}

var (
	logLevel     int32
	logFile      string
	outputFormat string
)

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err, "failed to create a new instance of the plugin")
	}

	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path")

	p.AddCommands(
		imagePullSecretAddCmd,
		imagePullSecretDeleteCmd,
		imagePullSecretListCmd,
		imagePullSecretUpdateCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func isSecretGenAPIAvailable() (bool, error) {
	clusterQueryClient, err := capdiscovery.NewClusterQueryClientForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		log.Error(err, "failed to create a new instance of the cluster query builder")
		return false, err
	}

	apiGroup := capdiscovery.Group("secretGenAPIQuery", "secretgen.carvel.dev").WithVersions("v1alpha1").WithResource("secretexports")
	return clusterQueryClient.Query(apiGroup).Execute()
}
