// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aunum/log"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

var descriptor = cli.PluginDescriptor{
	Name:        "cluster",
	Description: "Kubernetes cluster operations",
	Version:     "v0.0.1",
	Group:       cli.RunCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		createClusterCmd,
		listClustersCmd,
		deleteClusterCmd,
		upgradeClusterCmd,
		scaleClusterCmd,
		machineHealthCheckCmd,
		credentialsCmd,
		kubeconfigClusterCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

var updateClusterCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in progress...")
		return nil
	},
}

func getConfigDir() (string, error) {
	tanzuConfigDir, err := client.LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tanzuConfigDir, "tkg"), nil
}
