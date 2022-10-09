// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FeatureSpec defines the desired state of Feature
type FeatureSpec struct {
	// Description of the feature.
	Description string `json:"description,omitempty"`
	// Immutable indicates this feature cannot be toggled once set
	// If set at creation time, this state will persist for the life of the cluster
	Immutable bool `json:"immutable"`
	// Stability indicates stability level of this feature.
	// Stability levels are Work In Progress, Experimental, Technical Preview, Stable and Deprecated.
	// +kubebuilder:validation:Enum=Work In Progress;Experimental;Technical Preview;Stable;Deprecated
	// - Work In Progress: the default for new resources, represents local dev. intended to be hidden and deactivated
	// - Experimental: the first milestone meant for limited wider consumption, discoverable and defaults to deactivated
	// - Technical Preview: greater visibility for proven designs, discoverable and defaults to activated
	// - Stable: intended to be part of the mainline codebase, non-optional
	// - Deprecated: destined for future removal
	Stability string `json:"stability"`
}

// FeatureStatus defines the observed state of Feature
type FeatureStatus struct {
	// Activated is a boolean which indicates whether a feature is activated or not.
	Activated bool `json:"activated,omitempty"`
}

// Feature is the Schema for the features API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Activated",type=boolean,JSONPath=.spec.activated
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=.spec.description
// +kubebuilder:printcolumn:name="Immutable",type=boolean,JSONPath=.spec.immutable
// +kubebuilder:printcolumn:name="Stability",type=string,JSONPath=.spec.maturity
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
