// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

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
	// Discoverable indicates if clients should include consider the Feature available for their use
	// Allowing clients to control discoverability is one of the ways the API allows gradual rollout of functionality
	Discoverable bool `json:"discoverable"`
	// Activated defines the default state of the features activation
	Activated bool `json:"activated"`
	// Maturity indicates maturity level of this feature.
	// Maturity levels are Dev, Alpha, Beta, GA and Deprecated.
	// +kubebuilder:validation:Enum=dev;alpha;beta;ga;deprecated
	// - dev: the default for new resources, represents local dev. intended to be hidden and deactivated
	// - alpha: the first milestone meant for limited wider consumption, discoverable and defaults to deactivated
	// - beta: greater visibility for proven designs, discoverable and defaults to activated
	// - ga: intended to be part of the mainline codebase, non-optional
	// - deprecated: destined for future removal
	Maturity string `json:"maturity"`
}

// FeatureStatus defines the observed state of Feature
type FeatureStatus struct{}

// Feature is the Schema for the features API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Activated",type=boolean,JSONPath=.spec.activated
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=.spec.description
// +kubebuilder:printcolumn:name="Immutable",type=boolean,JSONPath=.spec.immutable
// +kubebuilder:printcolumn:name="Maturity",type=string,JSONPath=.spec.maturity
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
