// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregate

import (
	"k8s.io/apimachinery/pkg/util/sets"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// ComputeFeatureStates takes a FeatureGate spec and computes the actual state (activated, deactivated or unavailable)
// of the features in the gate by referring to a list of Feature resources.
func ComputeFeatureStates(featureGateSpec configv1alpha1.FeatureGateSpec, features []configv1alpha1.Feature) (activated, deactivated, unavailable []string) {
	// Collect features to be activated/deactivated in the spec.
	toActivate := sets.String{}
	toDeactivate := sets.String{}
	for _, f := range featureGateSpec.Features {
		if f.Activate {
			toActivate.Insert(f.Name)
		} else {
			toDeactivate.Insert(f.Name)
		}
	}

	// discovered is set a set of available features that are discoverable.
	discovered := sets.String{}
	// discoveredDefaultActivated is a set of available features that are discoverable and activated by default.
	discoveredDefaultActivated := sets.String{}
	// discoveredDefaultDeactivated is a set of available features that are discoverable and deactivated by default.
	discoveredDefaultDeactivated := sets.String{}
	for i := range features {
		feature := features[i]
		if !feature.Spec.Discoverable {
			continue
		}
		discovered.Insert(feature.Name)
		if feature.Spec.Activated {
			discoveredDefaultActivated.Insert(feature.Name)
		} else {
			discoveredDefaultDeactivated.Insert(feature.Name)
		}
	}

	// activate is all the features that the spec intends to be activated and features that are default activated.
	activate := discoveredDefaultActivated.Union(toActivate)
	// activationCandidates are features that are discovered, but are explicitly set *not* to be activated in this feature gate.
	// Only these features can be activated regardless of what the intent in the spec is.
	activationCandidates := discovered.Difference(toDeactivate)
	// Intersection gives us the actual activated features.
	activated = activationCandidates.Intersection(activate).List()

	// deactivate is all the features that the spec intends to be deactivated and features that are default deactivated.
	deactivate := discoveredDefaultDeactivated.Union(toDeactivate)
	// deactivationCandidates are features that are discovered, but are explicitly set *not* to be deactivated in this feature gate.
	// Only these features can be deactivated regardless of what the intent in the spec is.
	deactivationCandidates := discovered.Difference(toActivate)
	// Intersection gives us the actual deactivated features.
	deactivated = deactivationCandidates.Intersection(deactivate).List()

	// Set of all features specified in the current spec.
	allFeaturesInSpec := toActivate.Union(toDeactivate)
	// Set difference with all the discovered features gives unavailable features.
	unavailable = allFeaturesInSpec.Difference(discovered).List()

	return activated, deactivated, unavailable
}
