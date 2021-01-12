package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
)

var availbleUpgradesCmd = &cobra.Command{
	Use:   "available-upgrades",
	Short: "Get upgrading information for a Tanzu Kubernetes Release",
	Long:  `Get upgrading information for a Tanzu Kubernetes Release`,
}

var getAvailableUpgradesCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get all available upgrades for a specific Tanzu Kubernetes Release",
	Long:  `Get all available upgrades for a specific Tanzu Kubernetes Release`,
	Args:  cobra.ExactArgs(1),
	RunE:  getAvailableUpgrades,
}

func init() {
	availbleUpgradesCmd.AddCommand(getAvailableUpgradesCmd)
}

func getAvailableUpgrades(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
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

	for _, tkr := range tkrs {
		if tkr.Name == args[0] {
			for _, condition := range tkr.Status.Conditions {
				if condition.Type == runv1alpha1.ConditionUpgradeAvailable {
					upgradeMsg = condition.Message
				}
			}
		}
	}

	candidates := make(map[string]bool)
	if strs := strings.Split(upgradeMsg, ": "); len(strs) != 2 {
		fmt.Println("There are no availble upgrades for this TanzuKubernetesRelease")
	} else {
		names := strings.Split(strs[1], ",")
		for _, name := range names {
			candidates[name] = true
		}
	}

	t := component.NewTableWriter("NAME", "VERSION")
	for _, tkr := range tkrs {
		if _, ok := candidates[tkr.Name]; !ok {
			continue
		}

		t.Append([]string{tkr.Name, tkr.Spec.Version})
	}
	t.Render()

	return nil
}
