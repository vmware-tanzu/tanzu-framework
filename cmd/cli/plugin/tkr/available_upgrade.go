// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
)

const lenMsg = 2

var availableUpgradesCmd = &cobra.Command{
	Use:     "available-upgrades",
	Short:   "Get upgrade information for a Tanzu Kubernetes Release",
	Long:    `Get upgrade information for a Tanzu Kubernetes Release`,
	Aliases: []string{"avup"},
}

var getAvailableUpgradesCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get all available upgrades for a specific Tanzu Kubernetes Release",
	Long:  `Get all available upgrades for a specific Tanzu Kubernetes Release`,
	Args:  cobra.ExactArgs(1),
	RunE:  getAvailableUpgrades,
}

func init() {
	availableUpgradesCmd.AddCommand(getAvailableUpgradesCmd)
}

func getAvailableUpgrades(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClient, err := clusterclient.NewClusterClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	tkrs, err := clusterClient.GetTanzuKubernetesReleases("")
	if err != nil {
		return err
	}

	upgradeMsg := ""

	for i := range tkrs {
		if tkrs[i].Name == args[0] {
			for _, condition := range tkrs[i].Status.Conditions {
				if condition.Type == runv1alpha1.ConditionUpdatesAvailable {
					upgradeMsg = condition.Message
				}
			}
		}
	}

	candidates := make(map[string]bool)
	if strs := strings.Split(upgradeMsg, ": "); len(strs) != lenMsg {
		fmt.Println("There are no available upgrades for this TanzuKubernetesRelease")
	} else {
		names := strings.Split(strs[1], ",")
		for _, name := range names {
			candidates[name] = true
		}
	}

	t := component.NewTableWriter("NAME", "VERSION")
	for i := range tkrs {
		if _, ok := candidates[tkrs[i].Name]; !ok {
			continue
		}

		t.Append([]string{tkrs[i].Name, tkrs[i].Spec.Version})
	}
	t.Render()

	return nil
}
