// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	utilclusters "github.com/vmware-tanzu/tanzu-framework/apis/run/util/clusters"
	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	tkrutils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
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

func availableUpgrades(cmd *cobra.Command, args []string) error { //nolint:gocyclo
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}
	if server.IsGlobal() {
		return errors.New("getting available upgrades with a global server is not implemented yet")
	}
	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return err
	}

	clusterName := args[0]
	tkrs, err := clusterClient.GetTanzuKubernetesReleases("")
	if err != nil {
		return err
	}

	// TODO: update this condition after CLI fully support the package based cc.
	// Since CLI should support the pre package-based-cc where the updatesAvailable condition was part of
	// TKRs, code checking the TKRs for available upgrade should remain.
	cluster, err := getClusterResource(clusterClient, clusterName, au.namespace)
	if err == nil && capiconditions.Has(cluster, runv1.ConditionUpdatesAvailable) {
		return availableUpgradesFromCluster(cluster, tkrs, cmd.OutOrStdout())
	}

	clusterTKRName, err := getClusterTKRName(server, clusterName, au.namespace)
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

func getClusterTKRName(server *configapi.Server, clusterName, namespace string) (string, error) {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return "", errors.Wrap(err, "failed to create TKG client")
	}

	ccOptions := tkgctl.ListTKGClustersOptions{
		ClusterName: "",
		Namespace:   namespace,
		IncludeMC:   false,
	}
	clustersInfo, err := tkgctlClient.GetClusters(ccOptions)
	if err != nil {
		return "", errors.Wrap(err, "failed to get clusters Information")
	}
	clusterTkrVersion, err := getClusterTKRNameFromClusterInfo(clusterName, clustersInfo)
	if err != nil {
		return "", errors.Wrap(err, "failed to get cluster TKr name from clusters Information")
	}
	return clusterTkrVersion, nil
}

func availableUpgradesFromCluster(cluster *capiv1.Cluster, tkrs []runv1alpha1.TanzuKubernetesRelease, cmdOut io.Writer) error {
	updates := utilclusters.AvailableUpgrades(cluster)
	if len(updates) == 0 {
		fmt.Printf("no available upgrades for cluster '%s', namespace '%s'", cluster.Name, cluster.Namespace)
		return nil
	}

	t := component.NewOutputWriter(cmdOut, "table", "NAME", "VERSION", "COMPATIBLE")
	for i := range tkrs {
		if _, ok := updates[tkrs[i].Spec.Version]; !ok {
			continue
		}
		t.AddRow(tkrs[i].Name, tkrs[i].Spec.Version, "True")
	}
	t.Render()
	return nil
}
