// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"

	"github.com/vmware-tanzu-private/tkg-cli/pkg/constants"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type upgradeClustersOptions struct {
	namespace           string
	tkrName             string
	timeout             time.Duration
	unattended          bool
	osName              string
	osVersion           string
	osArch              string
	vSphereTemplateName string
}

var uc = &upgradeClustersOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:   "upgrade CLUSTER_NAME",
	Short: "Upgrade a cluster",
	Args:  cobra.ExactArgs(1),
	Example: `
  # Upgrade a workload cluster
  tanzu cluster upgrade wc-1

  # Upgrade a workload cluster with specific tkr v1.20.1---vmware.1-tkg.2
  tanzu cluster upgrade wc-1 --tkr v1.20.1---vmware.1-tkg.2

  # Upgrade a workload cluster using specific os name (vsphere)
  tanzu cluster upgrade wc-1 --os-name photon

  # Upgrade a workload cluster using specific os name and version
  tanzu cluster upgrade wc-1 --os-name ubuntu --os-version 20.04

  # Upgrade a workload cluster using specific os name, version and arch
  tanzu cluster upgrade wc-1 --os-name ubuntu --os-version 20.04 --os-arch amd64

  [+] : Options available for: os-name, os-version, os-arch are as follows:
  vSphere:
    --os-name ubuntu --os-version 20.04 --os-arch amd64
    --os-name photon --os-version 3 --os-arch amd64
  aws:
    --os-name ubuntu --os-version 20.04 --os-arch amd64
    --os-name amazon --os-version 2 --os-arch amd64
  azure:
    --os-name ubuntu --os-version 20.04 --os-arch amd64
    --os-name ubuntu --os-version 18.04 --os-arch amd64
`,
	RunE: upgrade,
}

func init() {
	upgradeClusterCmd.Flags().StringVarP(&uc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKR) to upgrade to")
	upgradeClusterCmd.Flags().StringVarP(&uc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	upgradeClusterCmd.Flags().DurationVarP(&uc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeClusterCmd.Flags().BoolVarP(&uc.unattended, "yes", "y", false, "Upgrade workload cluster without asking for confirmation")

	upgradeClusterCmd.Flags().StringVar(&uc.osName, "os-name", "", "OS name to use during cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeClusterCmd.Flags().StringVar(&uc.osVersion, "os-version", "", "OS version to use during cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeClusterCmd.Flags().StringVar(&uc.osArch, "os-arch", "", "OS arch to use during cluster upgrade. Discovered automatically if not provided (See [+])")

	upgradeClusterCmd.Flags().StringVarP(&uc.vSphereTemplateName, "vsphere-vm-template-name", "", "", "The vSphere VM template to be used with upgraded kubernetes version. Discovered automatically if not provided")
	upgradeClusterCmd.Flags().MarkHidden("vsphere-vm-template-name") //nolint
}

func upgrade(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("upgrading cluster with a global server is not implemented yet")
	}
	return upgradeCluster(server, args[0])
}

func upgradeCluster(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	tkrVersion := ""
	if uc.tkrName != "" {
		clusterClient, err := clusterclient.NewClusterClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
		if err != nil {
			return err
		}

		tkrVersion, err = getValidTkrVersionFromTkrForUpgrade(tkgctlClient, clusterClient, clusterName)
		if err != nil {
			return err
		}
	}

	upgradeClusterOptions := tkgctl.UpgradeClusterOptions{
		ClusterName:         clusterName,
		Namespace:           uc.namespace,
		TkrVersion:          tkrVersion,
		SkipPrompt:          uc.unattended,
		Timeout:             uc.timeout,
		OSName:              uc.osName,
		OSVersion:           uc.osVersion,
		OSArch:              uc.osArch,
		VSphereTemplateName: uc.vSphereTemplateName,
	}

	return tkgctlClient.UpgradeCluster(upgradeClusterOptions)
}

func getValidTkrVersionFromTkrForUpgrade(tkgctlClient tkgctl.TKGClient, clusterClient clusterclient.Client, clusterName string) (string, error) {
	result, err := tkgctlClient.DescribeCluster(tkgctl.DescribeTKGClustersOptions{
		ClusterName: clusterName,
		Namespace:   uc.namespace,
	})
	if err != nil {
		return "", err
	}

	tkrs, err := clusterClient.GetTanzuKubernetesReleases("")
	if err != nil {
		return "", err
	}

	tkrForUpgrade, err := getMatchingTkrForTkrName(tkrs, uc.tkrName)
	if err != nil {
		return "", err
	}
	if !isTkrCompatible(&tkrForUpgrade) {
		fmt.Printf("WARNING: TanzuKubernetesRelease %q is not compatible on the management cluster\n", tkrForUpgrade.Name)
	}

	tkrName, ok := result.Cluster.Labels["tanzuKubernetesRelease"]
	if !ok { // old clusters with no TKR label
		return tkrForUpgrade.Spec.Version, nil
	}

	tkr, err := getMatchingTkrForTkrName(tkrs, tkrName)
	if err != nil {
		return "", err
	}

	tkrAvailableUpgrades, err := getAvailableUpgrades(clusterName, &tkr)
	if err != nil {
		return "", err
	}

	for _, availableUpgrade := range tkrAvailableUpgrades {
		if availableUpgrade == uc.tkrName {
			return tkrForUpgrade.Spec.Version, nil
		}
	}

	return "", errors.Errorf("cluster cannot be upgraded to %q, available upgrades %v", uc.tkrName, tkrAvailableUpgrades)
}

func getAvailableUpgrades(clusterName string, tkr *runv1alpha1.TanzuKubernetesRelease) ([]string, error) {
	upgradeMsg := ""
	strLen := 2
	for _, condition := range tkr.Status.Conditions {
		if condition.Type == runv1alpha1.ConditionUpgradeAvailable {
			upgradeMsg = condition.Message
			break
		}
	}

	// Example upgradeMsg - "TKR(s) with later version is available: <tkr-name-1>,<tkr-name-2>"
	strs := strings.Split(upgradeMsg, ": ")
	if len(strs) != strLen {
		return []string{}, errors.Errorf("no available upgrades for cluster %q, namespace %q", clusterName, uc.namespace)
	}
	return strings.Split(strs[1], ","), nil
}

func getMatchingTkrForTkrName(tkrs []runv1alpha1.TanzuKubernetesRelease, tkrName string) (runv1alpha1.TanzuKubernetesRelease, error) {
	for i := range tkrs {
		if tkrs[i].Name == tkrName {
			return tkrs[i], nil
		}
	}

	return runv1alpha1.TanzuKubernetesRelease{}, errors.Errorf("could not find a matching TanzuKubernetesRelease for name %q", tkrName)
}

func isTkrCompatible(tkr *runv1alpha1.TanzuKubernetesRelease) bool {
	for _, condition := range tkr.Status.Conditions {
		if condition.Type == runv1alpha1.ConditionCompatible {
			compatible := string(condition.Status)
			return compatible == "True" || compatible == "true"
		}
	}

	return false
}
