// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregate

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	featureutil "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/util"
)

const (
	// TKGSystemFeatureGate is the FeatureGate resource for gating TKG features.
	TKGSystemFeatureGate = "tkg-system"
)

// TKGNamespaceSelector is a label selector which matches TKG-related namespaces.
var TKGNamespaceSelector = metav1.LabelSelector{
	MatchExpressions: []metav1.LabelSelectorRequirement{
		{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system-public"}},
	},
}

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

// FeatureActivatedInNamespace returns true only if all of the features specified are activated in the namespace.
func FeatureActivatedInNamespace(ctx context.Context, c client.Client, namespace, feature string) (bool, error) {
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{namespace}},
		},
	}
	return FeaturesActivatedInNamespacesMatchingSelector(ctx, c, selector, []string{feature})
}

// FeaturesActivatedInNamespacesMatchingSelector returns true only if all the features specified are activated in every namespace matched by the selector.
func FeaturesActivatedInNamespacesMatchingSelector(ctx context.Context, c client.Client, namespaceSelector metav1.LabelSelector, features []string) (bool, error) {
	namespaces, err := featureutil.NamespacesMatchingSelector(ctx, c, &namespaceSelector)
	if err != nil {
		return false, err
	}

	// If no namespaces are matched or no features specified, return false.
	if len(namespaces) == 0 || len(features) == 0 {
		return false, nil
	}

	featureGatesList := &configv1alpha1.FeatureGateList{}
	if err := c.List(ctx, featureGatesList); err != nil {
		return false, err
	}

	// Map of namespace to a set of features activated in that namespace.
	namespaceToActivatedFeatures := make(map[string]sets.String)
	for i := range featureGatesList.Items {
		fg := featureGatesList.Items[i]
		for _, namespace := range fg.Status.Namespaces {
			namespaceToActivatedFeatures[namespace] = sets.NewString(fg.Status.ActivatedFeatures...)
		}
	}

	for _, ns := range namespaces {
		activatedFeatures, found := namespaceToActivatedFeatures[ns]
		if !found {
			// Namespace has no features gated.
			return false, nil
		}
		// Feature is not activated in this namespace.
		if !activatedFeatures.HasAll(features...) {
			return false, nil
		}
	}
	return true, nil
}
