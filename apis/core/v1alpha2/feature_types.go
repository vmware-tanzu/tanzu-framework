// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StabilityLevel indicates stability level of the feature.
type StabilityLevel string

const (
	WorkInProgress   StabilityLevel = "Work In Progress"
	Experimental     StabilityLevel = "Experimental"
	TechnicalPreview StabilityLevel = "Technical Preview"
	Stable           StabilityLevel = "Stable"
	Deprecated       StabilityLevel = "Deprecated"
)

// FeatureSpec defines the desired state of Feature
type FeatureSpec struct {
	// Description of the feature.
	Description string `json:"description"`
	// Stability indicates stability level of the feature.
	// Stability levels are Work In Progress, Experimental, Technical Preview, Stable and Deprecated.
	// +kubebuilder:validation:Enum=Work In Progress;Experimental;Technical Preview;Stable;Deprecated
	// - Work In Progress: Feature is still under development. It is not ready to be used, except by the team working on it. Activating this feature is not recommended under any circumstances.
	// - Experimental: Feature is not ready, but may be used in pre-production environments. However, if an experimental feature has ever been used in an environment, that environment will not be supported. Activating an experimental feature requires you to permanently, irrevocably void all support guarantees for this environment by setting permanentlyVoidAllSupportGuarantees in feature reference in featuregate spec to true. You will need to recreate the environment to return to a supported state.
	// - Technical Preview: Feature is not ready, but is not believed to be dangerous. The feature itself is unsupported, but activating a technical preview feature does not affect the support status of the environment.
	// - Stable: Feature is ready and fully supported
	// - Deprecated: Feature is destined for removal, usage is discouraged. Deactivate this feature prior to upgrading to a release which has removed it to validate that you are not still using it and to prevent users from introducing new usage of it.
	Stability StabilityLevel `json:"stability"`
}

// FeatureStatus defines the observed state of Feature
type FeatureStatus struct {
	// Activated is a boolean which indicates whether a feature is activated or not.
	Activated bool `json:"activated"`
}

// Feature is the Schema for the features API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=.spec.description
// +kubebuilder:printcolumn:name="Stability",type=string,JSONPath=.spec.stability
// +kubebuilder:printcolumn:name="Activated?",type=string,JSONPath=.status.activated
type Feature struct {
	Status            FeatureStatus `json:"status,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FeatureSpec `json:"spec,omitempty"`
	metav1.TypeMeta   `json:",inline"`
}

// +kubebuilder:object:root=true

// FeatureList contains a list of Feature
type FeatureList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Feature `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Feature{}, &FeatureList{})
}
