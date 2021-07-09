// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package utils provides some utility functionalities for TKR.
package utils

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
)

// UpgradesNotAvailableError is an error type to return when upgrades are not available
type UpgradesNotAvailableError struct {
	message string
}

// NewUpgradesNotAvailableError returns a struct of type UpgradesNotAvailableError
func NewUpgradesNotAvailableError(message string) UpgradesNotAvailableError {
	return UpgradesNotAvailableError{
		message: message,
	}
}

func (e UpgradesNotAvailableError) Error() string {
	return e.message
}

// IsTkrCompatible checks if TKR is compatible
func IsTkrCompatible(tkr *runv1alpha1.TanzuKubernetesRelease) bool {
	for _, condition := range tkr.Status.Conditions {
		if condition.Type == runv1alpha1.ConditionCompatible {
			compatible := string(condition.Status)
			return compatible == "True" || compatible == "true"
		}
	}

	return false
}

// IsTkrActive checks if the TKR is active
func IsTkrActive(tkr *runv1alpha1.TanzuKubernetesRelease) bool {
	labels := tkr.Labels
	if labels != nil {
		if _, exists := labels[constants.TanzuKubernetesReleaseInactiveLabel]; exists {
			return false
		}
	}
	return true
}

// GetAvailableUpgrades returns the available upgrades of a TKR
func GetAvailableUpgrades(clusterName, namespace string, tkr *runv1alpha1.TanzuKubernetesRelease) ([]string, error) {
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
		return []string{}, NewUpgradesNotAvailableError(fmt.Sprintf("no available upgrades for cluster %q, namespace %q", clusterName, namespace))
	}

	var availableUpgradeList []string
	//TODO: Message format was changed to follow TKGs, keeping this old format check for backward compatibility.Can be cleaned up after couple minor version releases.
	if strings.Contains(upgradeMsg, "TKR(s)") {
		// Example for TKGm :upgradeMsg - "Deprecated, TKR(s) with later version is available: <tkr-name-1>,<tkr-name-2>"
		strs := strings.Split(upgradeMsg, ": ")
		if len(strs) != 2 {
			return []string{}, NewUpgradesNotAvailableError(fmt.Sprintf("no available upgrades for cluster %q, namespace %q", clusterName, namespace))
		}
		availableUpgradeList = strings.Split(strs[1], ",")
	} else {
		// Example for TKGs :upgradeMsg - [<tkr-version-1> <tkr-version-2>]"
		strs := strings.Split(strings.TrimRight(strings.TrimLeft(upgradeMsg, "["), "]"), " ")
		if len(strs) == 0 {
			return []string{}, NewUpgradesNotAvailableError(fmt.Sprintf("no available upgrades for cluster %q, namespace %q", clusterName, namespace))
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
