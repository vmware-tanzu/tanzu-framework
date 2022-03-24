// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	tkrutils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/utils"
)

var availableUpgradesCmd = &cobra.Command{
	Use:          "available-upgrades",
	Short:        "Get upgrade information for a cluster",
	Long:         `Get upgrade information for a cluster`,
	Aliases:      []string{"avup"},
	SilenceUsage: true,
}

var getAvailableUpgradesCmd = &cobra.Command{
	Use:          "get CLUSTER_NAME",
	Short:        "Get all available upgrades for a cluster",
	Long:         `Get all available upgrades for a cluster`,
	Args:         cobra.ExactArgs(1),
	RunE:         availableUpgrades,
	SilenceUsage: true,
}

type availableUpgradeOptions struct {
	namespace string
}

var au = &availableUpgradeOptions{}

func init() {
	getAvailableUpgradesCmd.Flags().StringVarP(&au.namespace, "namespace", "n", "default", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	availableUpgradesCmd.AddCommand(getAvailableUpgradesCmd)
}

func availableUpgrades(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting available upgrades with a global server is not implemented yet")
	}
	clusterName := args[0]
	clusterTKRName, err := getClusterTKRName(server, clusterName, au.namespace)
	if err != nil {
		return err
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return err
	}

	tkrs, err := clusterClient.GetTanzuKubernetesReleases("")
	if err != nil {
		return err
	}

	var upgradeTKrs []string
	for i := range tkrs {
		if tkrs[i].Name == clusterTKRName {
			upgradeTKrs, err = tkrutils.GetAvailableUpgrades(clusterName, au.namespace, &tkrs[i])
			if err != nil {
				if _, ok := err.(tkrutils.UpgradesNotAvailableError); ok {
					fmt.Printf("%v\n", err)
					return nil
				}
				return err
			}
		}
	}

	candidates := make(map[string]bool)
	for _, name := range upgradeTKrs {
		candidates[name] = true
	}

	t := component.NewOutputWriter(cmd.OutOrStdout(), "table", "NAME", "VERSION", "COMPATIBLE")
	for i := range tkrs {
		if _, ok := candidates[tkrs[i].Name]; !ok {
			continue
		}
		if !tkrutils.IsTkrActive(&tkrs[i]) {
			continue
		}
		compatible := "False"
		if tkrutils.IsTkrCompatible(&tkrs[i]) {
			compatible = "True"
		}
		t.AddRow(tkrs[i].Name, tkrs[i].Spec.Version, compatible)
	}
	t.Render()

	return nil
}

func getClusterTKRNameFromClusterInfo(clusterName string, clustersInfo []client.ClusterInfo) (string, error) {
	for idx := range clustersInfo {
		if clustersInfo[idx].Name == clusterName && clustersInfo[idx].Labels != nil {
			clusterTKr, err := getClusterTKRNameFromClusterLabels(clustersInfo[idx].Labels)
			if err != nil {
				return "", errors.Wrap(err, "failed to get cluster's current TKr name")
			}
			return clusterTKr, nil
		}
	}
	return "", errors.New("failed to get cluster's current TKr name")
}

func getClusterTKRName(server *configv1alpha1.Server, clusterName, namespace string) (string, error) {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return "", errors.Wrap(err, "failed to create TKG client")
	}

	ccOptions := tkgctl.ListTKGClustersOptions{
		ClusterName: "",
		Namespace:   namespace,
		IncludeMC:   false,
	}
	clusters, err := tkgctlClient.GetClusters(ccOptions)
	if err != nil {
		return "", errors.Wrap(err, "failed to get clusters Information")
	}
	clusterTkrVersion, err := getClusterTKRNameFromClusterInfo(clusterName, clusters)
	if err != nil {
		return "", errors.Wrap(err, "failed to get cluster TKr name from clusters Information")
	}
	return clusterTkrVersion, nil
}
