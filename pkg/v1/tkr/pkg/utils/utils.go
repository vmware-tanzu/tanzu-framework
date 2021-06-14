// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package utils provides some utility functionalities for TKR.
package utils

import (
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
)

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
