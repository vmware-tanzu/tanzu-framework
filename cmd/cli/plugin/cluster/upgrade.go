// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/constants"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type upgradeClustersOptions struct {
	namespace         string
	kubernetesVersion string
	timeout           time.Duration
	unattended        bool
}

var uc = &upgradeClustersOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:   "upgrade CLUSTER_NAME",
	Short: "Upgrade a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  upgrade,
}

func init() {
	upgradeClusterCmd.Flags().StringVarP(&uc.kubernetesVersion, "kubernetes-version", "k", "", "The kubernetes version to upgrade to")
	upgradeClusterCmd.Flags().StringVarP(&uc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	upgradeClusterCmd.Flags().DurationVarP(&uc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeClusterCmd.Flags().BoolVarP(&uc.unattended, "yes", "y", false, "Upgrade workload cluster without asking for confirmation")
}

func upgrade(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("upgrading cluster with a global server is not implemented yet")
	}
	return upgradeCluster(server, args[0])
}

func upgradeCluster(server *v1alpha1.Server, clusterName string) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	tkgctlClient, err := tkgctl.New(tkgctl.Options{
		ConfigDir:   configDir,
		KubeConfig:  server.ManagementClusterOpts.Path,
		KubeContext: server.ManagementClusterOpts.Context,
	})
	if err != nil {
		return err
	}

	upgradeClusterOptions := tkgctl.UpgradeClusterOptions{
		ClusterName:       clusterName,
		Namespace:         uc.namespace,
		KubernetesVersion: uc.kubernetesVersion,
		SkipPrompt:        uc.unattended,
		Timeout:           uc.timeout,
	}

	return tkgctlClient.UpgradeCluster(upgradeClusterOptions)
}
