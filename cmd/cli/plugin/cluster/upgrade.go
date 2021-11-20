// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	tkrutils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/utils"
)

type upgradeClustersOptions struct {
	namespace             string
	tkrName               string
	timeout               time.Duration
	unattended            bool
	osName                string
	osVersion             string
	osArch                string
	vSphereTemplateName   string
	// mdVSphereTemplateName is used to indicate the VM templated for
	// worker nodes, now it is used only for Windows nodes because
	// Linux cluster doesn't support master nodes and worker nodes
	// using different OS.  Not sure whether this will be supported
	// in the future for Linxu cluster, so let's use this name to
	// leave scalability.
	mdVSphereTemplateName string
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
	upgradeClusterCmd.Flags().StringVarP(&uc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKr) to upgrade to. If TKr name prefix is provided, the latest compatible TKr matching the TKr name prefix would be used")
	upgradeClusterCmd.Flags().StringVarP(&uc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	upgradeClusterCmd.Flags().DurationVarP(&uc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeClusterCmd.Flags().BoolVarP(&uc.unattended, "yes", "y", false, "Upgrade workload cluster without asking for confirmation")

	upgradeClusterCmd.Flags().StringVar(&uc.osName, "os-name", "", "OS name to use during cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeClusterCmd.Flags().StringVar(&uc.osVersion, "os-version", "", "OS version to use during cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeClusterCmd.Flags().StringVar(&uc.osArch, "os-arch", "", "OS arch to use during cluster upgrade. Discovered automatically if not provided (See [+])")

	upgradeClusterCmd.Flags().StringVarP(&uc.vSphereTemplateName, "vsphere-vm-template-name", "", "", "The vSphere VM template to be used with upgraded kubernetes version. Discovered automatically if not provided")
	upgradeClusterCmd.Flags().MarkHidden("vsphere-vm-template-name") //nolint

	upgradeClusterCmd.Flags().StringVarP(&uc.mdVSphereTemplateName, "vsphere-windows-vm-template-name", "", "", "The vSpherew Windows VM template to be used with upgraded kubernetes version. Discovered automatically if not provided")
        //upgradeClusterCmd.Flags().MarkHidden("vsphere-windows-vm-template-name") //nolint
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
		clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
		clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
		if err != nil {
			return err
		}

		tkrVersion, err = getValidTkrVersionFromTkrForUpgrade(tkgctlClient, clusterClient, clusterName)
		if err != nil {
			return err
		}
	}

	upgradeClusterOptions := tkgctl.UpgradeClusterOptions{
		ClusterName:           clusterName,
		Namespace:             uc.namespace,
		TkrVersion:            tkrVersion,
		SkipPrompt:            uc.unattended,
		Timeout:               uc.timeout,
		OSName:                uc.osName,
		OSVersion:             uc.osVersion,
		OSArch:                uc.osArch,
		VSphereTemplateName:   uc.vSphereTemplateName,
		MDVSphereTemplateName: uc.mdVSphereTemplateName,
		Edition:               BuildEdition,
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
	return getValidTKRVersionForUpgradeGivenTKRNamePrefix(clusterName, uc.namespace, uc.tkrName, result.ClusterInfo.K8sVersion, result.ClusterInfo.Labels, tkrs)
}

func getValidTKRVersionForUpgradeGivenFullTKRName(clusterName, namespace string, clusterLabels map[string]string,
	tkrForUpgrade *runv1alpha1.TanzuKubernetesRelease, tkrs []runv1alpha1.TanzuKubernetesRelease) (string, error) {

	var userWarningMsg string

	if !tkrutils.IsTkrActive(tkrForUpgrade) {
		return "", errors.Errorf("Tanzu Kubernetes release %q is deactivated and cannot be used", tkrForUpgrade.Name)
	}
	if !tkrutils.IsTkrCompatible(tkrForUpgrade) {
		userWarningMsg = fmt.Sprintf("WARNING: Tanzu Kubernetes release %q is not compatible on the management cluster", tkrForUpgrade.Name)
	}

	tkrName, err := getClusterTKRNameFromClusterLabels(clusterLabels)
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
	tkrAvailableUpgrades, err := tkrutils.GetAvailableUpgrades(clusterName, namespace, &tkr)
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
		userWarningMsg = fmt.Sprintf("WARNING: Tanzu Kubernetes release %q is not a recommended upgrade version for cluster %q, namespace %q", tkrForUpgrade.Name, clusterName, namespace)
	} else {
		userWarningMsg = fmt.Sprintf("%s, and not a recommended upgrade version for cluster %q, namespace %q", userWarningMsg, clusterName, namespace)
	}
	fmt.Println(userWarningMsg)
	return tkrForUpgrade.Spec.Version, nil
}

func getValidTKRVersionForUpgradeGivenTKRNamePrefix(clusterName, namespace, tkrNamePrefix, clusterK8sVersion string,
	clusterLabels map[string]string, tkrs []runv1alpha1.TanzuKubernetesRelease) (string, error) {

	var err error
	var upgradeEligibleTKRs []runv1alpha1.TanzuKubernetesRelease

	tkrName, err := getClusterTKRNameFromClusterLabels(clusterLabels)
	if err != nil { // old clusters with no TKR label
		upgradeEligibleTKRs, err = getUpgradableTKRsCompatibleWithClusterK8sVersion(clusterK8sVersion, tkrs)
		if err != nil {
			return "", errors.Wrap(err, "failed to get upgrade eligible TKrs")
		}
	} else {
		upgradeEligibleTKRs, err = getUpgradableTKRsCompatibleWithClusterTKR(clusterName, namespace, tkrName, tkrs)
		if err != nil {
			return "", errors.Wrap(err, "failed to get upgrade eligible TKrs")
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
		return "", errors.Errorf("cluster cannot be upgraded, no compatible upgrades found matching the TKr prefix '%s', available compatible upgrades %v", tkrNamePrefix, getTKRNamesFromTKRs(upgradeEligibleTKRs))
	}

	return getLatestTkrVersion(upgradeEligiblePrefixMatchedTKRs)
}

func getUpgradableTKRsCompatibleWithClusterK8sVersion(clusterk8sVersion string, tkrs []runv1alpha1.TanzuKubernetesRelease) ([]runv1alpha1.TanzuKubernetesRelease, error) {
	upgradeEligibleTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	for idx := range tkrs {
		if !tkrutils.IsTkrCompatible(&tkrs[idx]) {
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
func getUpgradableTKRsCompatibleWithClusterTKR(clusterName, namespace, clusterTKRName string,
	tkrs []runv1alpha1.TanzuKubernetesRelease) ([]runv1alpha1.TanzuKubernetesRelease, error) {

	tkr, err := getMatchingTkrForTkrName(tkrs, clusterTKRName)
	if err != nil {
		return nil, err
	}
	tkrAvailableUpgrades, err := tkrutils.GetAvailableUpgrades(clusterName, namespace, &tkr)
	if err != nil {
		return nil, err
	}

	upgradeEligibleTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	for _, tkrName := range tkrAvailableUpgrades {
		for idx := range tkrs {
			if !tkrutils.IsTkrCompatible(&tkrs[idx]) || !tkrutils.IsTkrActive(&tkrs[idx]) {
				continue
			}
			if tkrName == tkrs[idx].Name {
				upgradeEligibleTKRs = append(upgradeEligibleTKRs, tkrs[idx])
			}
		}
	}
	return upgradeEligibleTKRs, nil
}

func getMatchingTkrForTkrName(tkrs []runv1alpha1.TanzuKubernetesRelease, tkrName string) (runv1alpha1.TanzuKubernetesRelease, error) {
	for i := range tkrs {
		if tkrs[i].Name == tkrName {
			return tkrs[i], nil
		}
	}

	return runv1alpha1.TanzuKubernetesRelease{}, errors.Errorf("could not find a matching Tanzu Kubernetes release for name %q", tkrName)
}

func getTKRNamesFromTKRs(tkrs []runv1alpha1.TanzuKubernetesRelease) []string {
	tkrNames := []string{}
	for idx := range tkrs {
		tkrNames = append(tkrNames, tkrs[idx].Name)
	}
	return tkrNames
}

func getClusterTKRNameFromClusterLabels(clusterLabels map[string]string) (string, error) {
	tkrLabelExists := false
	tkrName := ""
	// TODO: Once the TKGs and TKGm label for TKR are same, update this
	tkrName, tkrLabelExists = clusterLabels["run.tanzu.vmware.com/tkr"] // TKGs
	if !tkrLabelExists {
		tkrName, tkrLabelExists = clusterLabels["tanzuKubernetesRelease"] // TKGm
	}
	if !tkrLabelExists {
		return "", errors.New("failed to get cluster TKr name from cluster labels")
	}
	return tkrName, nil
}
