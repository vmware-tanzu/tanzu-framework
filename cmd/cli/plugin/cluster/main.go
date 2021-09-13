// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"

	"github.com/aunum/log"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

var (
	// BuildEdition is the edition the CLI was built for. This variable will be set during the build.
	BuildEdition string
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "cluster",
	Description: "Kubernetes cluster operations",
	Group:       cliv1alpha1.RunCmdGroup,
	Aliases:     []string{"cl", "clusters"},
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
		createClusterCmd,
		listClustersCmd,
		deleteClusterCmd,
		upgradeClusterCmd,
		scaleClusterCmd,
		machineHealthCheckCmd,
		credentialsCmd,
		clusterKubeconfigCmd,
		getClustersCmd,
		availableUpgradesCmd,
		clusterNodePoolCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func getConfigDir() (string, error) {
	tanzuConfigDir, err := config.LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tanzuConfigDir, "tkg"), nil
}

func createTKGClient(kubeconfig, kubecontext string) (tkgctl.TKGClient, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}
	return tkgctl.New(tkgctl.Options{
		ConfigDir:   configDir,
		KubeConfig:  kubeconfig,
		KubeContext: kubecontext,
		LogOptions:  tkgctl.LoggingOptions{Verbosity: logLevel, File: logFile},
	})
}
