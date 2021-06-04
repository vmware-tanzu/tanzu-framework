// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/utils"
)

type getTKROptions struct {
	listAll bool
}

var gtkr = &getTKROptions{}

var getTanzuKubernetesRleasesCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get available Tanzu Kubernetes Releases",
	Long:  "Get available Tanzu Kubernetes Releases",
	RunE:  getKubernetesReleases,
}

func init() {
	getTanzuKubernetesRleasesCmd.Flags().BoolVarP(&gtkr.listAll, "all", "a", false, "List all the available Tanzu Kubernetes releases including Incompatible and deactivated")
}
func getKubernetesReleases(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return err
	}
	tkrName := ""
	if len(args) != 0 {
		tkrName = args[0]
	}
	return runGetKubernetesReleases(clusterClient, tkrName)
}

func runGetKubernetesReleases(clusterClient clusterclient.Client, tkrName string) error {
	tkrs, err := clusterClient.GetTanzuKubernetesReleases(tkrName)
	if err != nil {
		return err
	}

	t := component.NewTableWriter("NAME", "VERSION", "COMPATIBLE", "ACTIVE", "UPDATES AVAILABLE")
	for i := range tkrs {
		compatible := ""
		upgradeAvailable := ""
		for _, condition := range tkrs[i].Status.Conditions {
			if condition.Type == runv1alpha1.ConditionCompatible {
				compatible = string(condition.Status)
			}
			// ConditionUpgradeAvailable is deprecated, but check here for backward compatibility
			if condition.Type == runv1alpha1.ConditionUpdatesAvailable || condition.Type == runv1alpha1.ConditionUpgradeAvailable {
				upgradeAvailable = string(condition.Status)
			}
		}
		labels := tkrs[i].Labels
		activeStatus := "True"
		if labels != nil {
			if _, exists := labels[constants.TanzuKubernetesReleaseInactiveLabel]; exists {
				activeStatus = "False"
			}
		}

		if !gtkr.listAll && (!utils.IsTkrCompatible(&tkrs[i]) || !utils.IsTkrActive(&tkrs[i])) {
			continue
		}
		t.Append([]string{tkrs[i].Name, tkrs[i].Spec.Version, compatible, activeStatus, upgradeAvailable})
	}
	t.Render()
	return nil
}
