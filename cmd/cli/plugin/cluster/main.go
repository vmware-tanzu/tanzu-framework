// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"

	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

var descriptor = cli.PluginDescriptor{
	Name:        "cluster",
	Description: "Kubernetes cluster operations",
	Version:     cli.BuildVersion,
	Group:       cli.RunCmdGroup,
}

var logLevel int32
var logFile string

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "v", "v", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVar(&logFile, "log_file", "", "Log file path")

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
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func getConfigDir() (string, error) {
	tanzuConfigDir, err := client.LocalDir()
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
