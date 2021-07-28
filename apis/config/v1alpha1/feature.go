// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OverrideFeatures allows a slice of features to be overridden by the gate
func (s *FeatureGate) OverrideFeatureActivation(features []Feature) []Feature {
	updated := []Feature{}
	copy(features, updated)

OUTER:
	for _, feature := range updated {
		for _, f := range s.Status.ActivatedFeatures {
			if f == feature.Name {
				feature.Spec.Activated = true
				continue OUTER
			}
		}
		for _, f := range s.Status.DeactivatedFeatures {
			if f == feature.Name {
				feature.Spec.Activated = false
				continue OUTER
			}
		}
	}
	return updated
}

func (s *FeatureGate) OverridesFeature(feature Feature) bool {
	for _, f := range s.Status.ActivatedFeatures {
		if f == feature.Name {
			return true
		}
	}
	for _, f := range s.Status.DeactivatedFeatures {
		if f == feature.Name {
			return true
		}
	}
	return false
}
