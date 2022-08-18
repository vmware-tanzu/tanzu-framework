// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/util/conditions"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

const lenMsg = 2

var AvailableUpgradesCmd = &cobra.Command{
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
	AvailableUpgradesCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	AvailableUpgradesCmd.AddCommand(getAvailableUpgradesCmd)
}

func getAvailableUpgrades(cmd *cobra.Command, args []string) error {
	clusterClient, err := getClusterClient()
	if err != nil {
		return err
	}

	tkrs, err := clusterClient.GetTanzuKubernetesReleases("")
	if err != nil {
		return err
	}

	results := availableUpdatesInTKRs(tkrs, args[0])
	if len(results) == 0 {
		fmt.Println("There are no available upgrades for this TanzuKubernetesRelease")
		return nil
	}

	t := component.NewOutputWriter(AvailableUpgradesCmd.OutOrStdout(), outputFormat, "NAME", "VERSION")
	for i := range results {
		t.AddRow(results[i].Name, results[i].Spec.Version)
	}
	t.Render()

	return nil
}

func getClusterClient() (clusterclient.Client, error) {
	server, err := config.GetCurrentServer()
	if err != nil {
		return nil, err
	}

	if server.IsGlobal() {
		return nil, errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return nil, err
	}
	return clusterClient, nil
}

func availableUpdatesInTKRs(tkrs []runv1alpha1.TanzuKubernetesRelease, tkrName string) []runv1alpha1.TanzuKubernetesRelease {
	tkr := tkrByName(tkrs, tkrName)
	if tkr == nil {
		return nil
	}

	candidates := namesFromUpdatesAvailable(tkr)
	if len(candidates) == 0 {
		candidates = namesFromUpgradeAvailable(tkr) // fall back to UpgradeAvailable condition
	}

	results := filterTKRs(tkrs, func(tkr *runv1alpha1.TanzuKubernetesRelease) bool {
		if !conditions.IsTrue(tkr, runv1alpha1.ConditionCompatible) {
			return false
		}
		_, exists := candidates[tkr.Name]
		return exists
	})
	return results
}

type stringSet map[string]struct{}

func (set stringSet) Add(ss ...string) stringSet {
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

func namesFromUpdatesAvailable(tkr *runv1alpha1.TanzuKubernetesRelease) stringSet {
	conditionUpdatesAvailable := conditions.Get(tkr, runv1alpha1.ConditionUpdatesAvailable)
	if conditionUpdatesAvailable == nil || conditionUpdatesAvailable.Status != corev1.ConditionTrue {
		return stringSet{}
	}

	message := strings.Trim(conditionUpdatesAvailable.Message, "[]")
	message = strings.ReplaceAll(message, ",", " ")

	versions := strings.Fields(message)

	names := mapStrings(versions, tkrNameFromVersion)
	candidates := stringSet{}.Add(names...)
	return candidates
}

func namesFromUpgradeAvailable(tkr *runv1alpha1.TanzuKubernetesRelease) stringSet {
	conditionUpgradeAvailable := conditions.Get(tkr, runv1alpha1.ConditionUpgradeAvailable)
	if conditionUpgradeAvailable == nil || conditionUpgradeAvailable.Status != corev1.ConditionTrue {
		return stringSet{}
	}

	strs := strings.Split(conditionUpgradeAvailable.Message, ": ")
	if len(strs) != lenMsg {
		return stringSet{}
	}

	versions := strings.Split(strs[1], ",")

	names := mapStrings(versions, tkrNameFromVersion)
	candidates := stringSet{}.Add(names...)
	return candidates
}

func tkrNameFromVersion(version string) string {
	return strings.ReplaceAll(version, "+", "---")
}

type stringMapper func(s string) string

func mapStrings(ss []string, m stringMapper) []string {
	result := make([]string, len(ss))
	for i, s := range ss {
		result[i] = m(s)
	}
	return result
}

type tkrPredicate func(release *runv1alpha1.TanzuKubernetesRelease) bool

func filterTKRs(tkrs []runv1alpha1.TanzuKubernetesRelease, p tkrPredicate) []runv1alpha1.TanzuKubernetesRelease {
	result := make([]runv1alpha1.TanzuKubernetesRelease, 0, len(tkrs))
	for i := range tkrs {
		if p(&tkrs[i]) {
			result = append(result, tkrs[i])
		}
	}
	return result
}

func tkrByName(tkrs []runv1alpha1.TanzuKubernetesRelease, name string) *runv1alpha1.TanzuKubernetesRelease {
	for i := range tkrs {
		if tkrs[i].Name == name {
			return &tkrs[i]
		}
	}
	return nil
}
