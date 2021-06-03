// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"

	"github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"

	corev1 "k8s.io/api/core/v1"
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

  # Upgrade a workload cluster with tkr prefix v1.20.1
  tanzu cluster upgrade wc-1 --tkr v1.20.1

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
	upgradeClusterCmd.Flags().StringVarP(&uc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKR) to upgrade to. If TKR name prefix is provided, the latest compatible TKR matching the TKR name prefix would be used")
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
	// If the complete TKR name is provided, use it
	if err == nil {
		return getValidTKRVersionForUpgradeGivenFullTKRName(clusterName, uc.namespace, result.ClusterInfo.Labels, &tkrForUpgrade, tkrs)
	}
	return getValidTKRVersionForUpgradeGivenTKRNamePrefix(clusterName, uc.tkrName, result.ClusterInfo.K8sVersion, result.ClusterInfo.Labels, tkrs)
}

func getValidTKRVersionForUpgradeGivenFullTKRName(clusterName, namespace string, clusterLabels map[string]string,
	tkrForUpgrade *runv1alpha1.TanzuKubernetesRelease, tkrs []runv1alpha1.TanzuKubernetesRelease) (string, error) {
	var userWarningMsg string
	if !isTkrCompatible(tkrForUpgrade) {
		userWarningMsg = fmt.Sprintf("WARNING: TanzuKubernetesRelease %q is not compatible on the management cluster", tkrForUpgrade.Name)
	}

	tkrName, err := getClusterTKRVersion(clusterLabels)
	if err != nil { // old clusters with no TKR label
		if userWarningMsg != "" {
			fmt.Println(userWarningMsg)
		}
		return tkrForUpgrade.Spec.Version, nil
	}

	// get the TKR associated with the cluster
	tkr, err := getMatchingTkrForTkrName(tkrs, tkrName)
	if err != nil {
		return "", err
	}
	// get available upgrades from the TKR associated with the cluster currently
	tkrAvailableUpgrades, err := getAvailableUpgrades(clusterName, &tkr)
	if err != nil {
		return "", err
	}

	for _, availableUpgrade := range tkrAvailableUpgrades {
		if availableUpgrade == tkrForUpgrade.Name {
			if userWarningMsg != "" {
				fmt.Println(userWarningMsg)
			}
			return tkrForUpgrade.Spec.Version, nil
		}
	}
	// custom TKRs may not follow the Tanzu naming conventions, so they may not be appear in available upgrade list
	if userWarningMsg == "" {
		userWarningMsg = fmt.Sprintf("WARNING: TanzuKubernetesRelease %q is not a recommended upgrade version for cluster %q, namespace %q", tkrForUpgrade.Name, clusterName, namespace)
	} else {
		userWarningMsg = fmt.Sprintf("%s, and not a recommended upgrade version for cluster %q, namespace %q", userWarningMsg, clusterName, namespace)
	}
	fmt.Println(userWarningMsg)
	return tkrForUpgrade.Spec.Version, nil
}

func getValidTKRVersionForUpgradeGivenTKRNamePrefix(clusterName string, tkrNamePrefix, clusterK8sVersion string,
	clusterLabels map[string]string, tkrs []runv1alpha1.TanzuKubernetesRelease) (string, error) {
	var err error
	var upgradeEligibleTKRs []runv1alpha1.TanzuKubernetesRelease

	tkrName, err := getClusterTKRVersion(clusterLabels)
	if err != nil { // old clusters with no TKR label
		upgradeEligibleTKRs, err = getUpgradableTKRsCompatibleWithClusterK8sVersion(clusterK8sVersion, tkrs)
		if err != nil {
			return "", errors.Wrap(err, "failed to get upgrade eligible TKRs")
		}
	} else {
		upgradeEligibleTKRs, err = getUpgradableTKRsCompatibleWithClusterTKR(clusterName, tkrName, tkrs)
		if err != nil {
			return "", errors.Wrap(err, "failed to get upgrade eligible TKRs")
		}
	}
	if len(upgradeEligibleTKRs) == 0 {
		return "", errors.New("cluster cannot be upgraded as there are no available upgrades")
	}

	upgradeEligiblePrefixMatchedTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	// filter the eligible tkrs matching prefix
	for idx := range upgradeEligibleTKRs {
		if strings.HasPrefix(upgradeEligibleTKRs[idx].Name, tkrNamePrefix) {
			upgradeEligiblePrefixMatchedTKRs = append(upgradeEligiblePrefixMatchedTKRs, upgradeEligibleTKRs[idx])
		}
	}

	if len(upgradeEligiblePrefixMatchedTKRs) == 0 {
		return "", errors.Errorf("cluster cannot be upgraded, no compatible upgrades found matching the TKR prefix '%s', available compatible upgrades %v", tkrNamePrefix, getTKRNamesFromTKRs(upgradeEligibleTKRs))
	}

	return getLatestTkrVersion(upgradeEligiblePrefixMatchedTKRs)
}

func getUpgradableTKRsCompatibleWithClusterK8sVersion(clusterk8sVersion string, tkrs []runv1alpha1.TanzuKubernetesRelease) ([]runv1alpha1.TanzuKubernetesRelease, error) {
	upgradeEligibleTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	for idx := range tkrs {
		if !isTkrCompatible(&tkrs[idx]) {
			continue
		}
		compareResult, err := utils.CompareVMwareVersionStrings(clusterk8sVersion, tkrs[idx].Spec.KubernetesVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "error while comparing kubernetes versions %s,%s", clusterk8sVersion, tkrs[idx].Spec.KubernetesVersion)
		}

		if compareResult > 0 {
			continue
		}

		if !utils.CheckKubernetesUpgradeCompatibility(clusterk8sVersion, tkrs[idx].Spec.KubernetesVersion) {
			continue
		}
		upgradeEligibleTKRs = append(upgradeEligibleTKRs, tkrs[idx])
	}
	return upgradeEligibleTKRs, nil
}
func getUpgradableTKRsCompatibleWithClusterTKR(clusterName, clusterTKRName string,
	tkrs []runv1alpha1.TanzuKubernetesRelease) ([]runv1alpha1.TanzuKubernetesRelease, error) {
	tkr, err := getMatchingTkrForTkrName(tkrs, clusterTKRName)
	if err != nil {
		return nil, err
	}
	tkrAvailableUpgrades, err := getAvailableUpgrades(clusterName, &tkr)
	if err != nil {
		return nil, err
	}

	upgradeEligibleTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	for _, tkrName := range tkrAvailableUpgrades {
		for idx := range tkrs {
			if !isTkrCompatible(&tkrs[idx]) {
				continue
			}
			if tkrName == tkrs[idx].Name {
				upgradeEligibleTKRs = append(upgradeEligibleTKRs, tkrs[idx])
			}
		}
	}
	return upgradeEligibleTKRs, nil
}

func getAvailableUpgrades(clusterName string, tkr *runv1alpha1.TanzuKubernetesRelease) ([]string, error) {
	upgradeMsg := ""

	for _, condition := range tkr.Status.Conditions {
		if condition.Type == runv1alpha1.ConditionUpdatesAvailable && condition.Status == corev1.ConditionTrue {
			upgradeMsg = condition.Message
			break
		}
		// If the TKR's have deprecated UpgradeAvailable condition use it
		if condition.Type == runv1alpha1.ConditionUpgradeAvailable && condition.Status == corev1.ConditionTrue {
			upgradeMsg = condition.Message
			break
		}
	}

	if upgradeMsg == "" {
		return []string{}, errors.Errorf("no available upgrades for cluster %q, namespace %q", clusterName, uc.namespace)
	}

	var availableUpgradeList []string
	//TODO: Message format was changed to follow TKGs, keeping this old format check for backward compatibility.Can be cleaned up after couple minor version releases.
	if strings.Contains(upgradeMsg, "TKR(s)") {
		// Example for TKGm :upgradeMsg - "Deprecated, TKR(s) with later version is available: <tkr-name-1>,<tkr-name-2>"
		strs := strings.Split(upgradeMsg, ": ")
		if len(strs) != 2 { //nolint
			return []string{}, errors.Errorf("no available upgrades for cluster %q, namespace %q", clusterName, uc.namespace)
		}
		availableUpgradeList = strings.Split(strs[1], ",")
	} else {
		// Example for TKGs :upgradeMsg - [<tkr-version-1> <tkr-version-2>]"
		strs := strings.Split(strings.TrimRight(strings.TrimLeft(upgradeMsg, "["), "]"), " ")
		if len(strs) == 0 {
			return []string{}, errors.Errorf("no available upgrades for cluster %q, namespace %q", clusterName, uc.namespace)
		}
		availableUpgradeList = strs
	}

	// convert them to tkrName if the available upgrade list contains TKR versions
	for idx := range availableUpgradeList {
		if !strings.HasPrefix(availableUpgradeList[idx], "v") {
			availableUpgradeList[idx] = "v" + availableUpgradeList[idx]
		}
		availableUpgradeList[idx] = utils.GetTkrNameFromTkrVersion(availableUpgradeList[idx])
	}

	return availableUpgradeList, nil
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

func getTKRNamesFromTKRs(tkrs []runv1alpha1.TanzuKubernetesRelease) []string {
	tkrNames := []string{}
	for idx := range tkrs {
		tkrNames = append(tkrNames, tkrs[idx].Name)
	}
	return tkrNames
}

func getClusterTKRVersion(clusterLabels map[string]string) (string, error) {
	tkrLabelExists := false
	tkrName := ""
	// TODO: Once the TKGs and TKGm label for TKR are same, update this
	tkrName, tkrLabelExists = clusterLabels["run.tanzu.vmware.com/tkr"] // TKGs
	if !tkrLabelExists {
		tkrName, tkrLabelExists = clusterLabels["tanzuKubernetesRelease"] // TKGm
	}
	if !tkrLabelExists {
		return "", errors.New("failed to get cluster TKR version from cluster labels")
	}
	return tkrName, nil
}
