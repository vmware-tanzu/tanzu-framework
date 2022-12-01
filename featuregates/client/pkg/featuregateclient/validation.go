// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregateclient

import (
	"fmt"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

// validateFeatureActivationToggle ensures the given Feature can be activated.
func validateFeatureActivationToggle(gates *corev1alpha2.FeatureGateList, feature *corev1alpha2.Feature) error {
	if err := featureExistsInOneAndOnlyOneFeaturegate(gates, feature.Name); err != nil {
		return fmt.Errorf("could not validate Feature changing activation set point: %w", err)
	}

	if err := featureActivationToggleAllowed(feature); err != nil {
		return fmt.Errorf("could not validate Feature changing activation set point: %w", err)
	}

	return nil
}

// featureExistsInOneFeaturegate checks that the Feature exists in one and only one FeatureGate.
func featureExistsInOneAndOnlyOneFeaturegate(gates *corev1alpha2.FeatureGateList, featureName string) error {
	n := qtyFeatureGatesContainingFeature(gates, featureName)
	if n > 1 {
		return fmt.Errorf("the Feature %s was found in more than one FeatureGate: %w", featureName, ErrTypeTooMany)
	}
	if n == 0 {
		return fmt.Errorf("the Feature %s must exist in one FeatureGate: %w", featureName, ErrTypeNotFound)
	}
	return nil
}

// qtyFeatureGatesContainingFeature counts the number of FeatureGates that
// references the provided Feature.
func qtyFeatureGatesContainingFeature(gates *corev1alpha2.FeatureGateList, featureName string) int {
	var n int
	for i := range gates.Items {
		for _, ref := range gates.Items[i].Spec.Features {
			if ref.Name == featureName {
				n++
			}
		}
	}
	return n
}

// featureActivationToggleAllowed checks if a Feature is considered immutable by its stability
// level and associated policy. Immutable means a Feature's activation setting cannot be toggled.
func featureActivationToggleAllowed(feature *corev1alpha2.Feature) error {
	stability := feature.Spec.Stability
	policy := corev1alpha2.GetPolicyForStabilityLevel(stability)

	if policy.Immutable {
		return fmt.Errorf("activation setting for Feature %s cannot be toggled as its stability level is %s: %w", feature.Name, stability, ErrTypeForbidden)
	}
	return nil
}
