// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/clusters"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	tkrutils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
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

const (
	LegacyClusterTKRLabel = "tanzuKubernetesRelease"
)

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
	RunE:         upgrade,
	SilenceUsage: true,
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
	upgradeClusterCmd.Flags().MarkHidden("vsphere-vm-template-name") // nolint
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

func upgradeCluster(server *configapi.Server, clusterName string) error {
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

	edition, err := config.GetEdition()
	if err != nil {
		return err
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
		Edition:             edition,
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

	// TODO: update this condition after CLI fully support the package based cc.
	// Since CLI should support the pre package-based-cc where the updatesAvailable condition was part of
	// TKRs, code checking the TKRs for available upgrade should remain.
	cluster, err := getClusterResource(clusterClient, clusterName, uc.namespace)
	if err == nil && capiconditions.Has(cluster, runv1.ConditionUpdatesAvailable) {
		return getValidTKRVersionFromClusterForUpgrade(cluster, uc.tkrName)
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

	// If existing TKR on the cluster and the one requested are same
	// continue the upgrade using the given TKR
	if tkrName == tkrForUpgrade.GetName() {
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
	tkrName, tkrLabelExists = clusterLabels[runv1.LabelTKR]
	if !tkrLabelExists {
		tkrName, tkrLabelExists = clusterLabels[LegacyClusterTKRLabel] // legacy TKGm
	}
	if !tkrLabelExists {
		return "", errors.New("failed to get cluster TKr name from cluster labels")
	}
	return tkrName, nil
}

func getValidTKRVersionFromClusterForUpgrade(cluster *capiv1.Cluster, tkrName string) (string, error) {
	tkrVersion := strings.ReplaceAll(tkrName, "---", "+")
	// If the existing TKR on the cluster and the one requested are same
	// continue the upgrade using the given TKR
	if clusterTKR, err := getClusterTKRNameFromClusterLabels(cluster.Labels); err == nil && clusterTKR == tkrName {
		return tkrVersion, nil
	}

	updates := clusters.AvailableUpgrades(cluster)
	if len(updates) == 0 {
		return "", errors.Errorf("no available upgrades for cluster '%s', namespace '%s'", cluster.Name, cluster.Namespace)
	}

	matchingUpdates := updates.Filter(func(tv string) bool {
		return version.Prefixes(tv).Has(tkrVersion)
	})
	if len(matchingUpdates) == 0 {
		updateNames := updates.Map(func(tv string) string {
			return strings.ReplaceAll(tv, "+", "---")
		})
		return "", errors.Errorf("cluster cannot be upgraded, no compatible upgrades found matching the TKR name/prefix '%s', available compatible upgrades %s", tkrName, updateNames.Slice())
	}
	return latestTkrVersion(matchingUpdates.Slice()), nil
}

func getClusterResource(clusterClient clusterclient.Client, clusterName, clusterNamespace string) (*capiv1.Cluster, error) {
	if clusterNamespace == "" {
		clusterNamespace = constants.DefaultNamespace
	}
	cluster := &capiv1.Cluster{}
	err := clusterClient.GetResource(cluster, clusterName, clusterNamespace, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get cluster %q from namespace %q", clusterName, clusterNamespace)
	}
	return cluster, nil
}

// latestTkrVersion returns the latest version
// pre-req versions should not be empty and versions are valid semantic versions
func latestTkrVersion(versions []string) string {
	sort.Slice(versions, func(i, j int) bool {
		vi, _ := version.ParseSemantic(versions[i])
		vj, _ := version.ParseSemantic(versions[j])
		return vi.LessThan(vj)
	})

	log.V(4).Infof("Using the TKr version '%s' ", versions[len(versions)-1])
	return versions[len(versions)-1]
}
