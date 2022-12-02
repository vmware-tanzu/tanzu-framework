// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FeatureReference refers to a Feature resource and specifies its intended activation state.
type FeatureReference struct {
	// Name is the name of the Feature resource, which represents a feature the system offers.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Activate indicates the activation intent for the feature.
	Activate bool `json:"activate,omitempty"`
	// PermanentlyVoidAllSupportGuarantees when set to true permanently voids all support guarantees.
	// Once set to true, cannot be set back to false
	PermanentlyVoidAllSupportGuarantees bool `json:"permanentlyVoidAllSupportGuarantees,omitempty"`
}

// FeatureGateSpec defines the desired state of FeatureGate
type FeatureGateSpec struct {
	// Features is a slice of FeatureReference to gate features.
	// Feature controller sets the specified activation state only if the Feature policy is satisfied.
	// +listType=map
	// +listMapKey=name
	Features []FeatureReference `json:"features,omitempty"`
}

// FeatureGateStatus defines the observed state of FeatureGate
type FeatureGateStatus struct {
	// FeatureReferenceResult represents the results of all the features specified in the FeatureGate spec.
	// +listType=map
	// +listMapKey=name
	FeatureReferenceResults []FeatureReferenceResult `json:"featureReferenceResults"`
}

// FeatureReferenceStatus represents the status of the feature reference in the FeatureGate spec
type FeatureReferenceStatus string

const (
	AppliedReferenceStatus FeatureReferenceStatus = "Applied"
	InvalidReferenceStatus FeatureReferenceStatus = "Invalid"
)

// FeatureReferenceResult represents the result of FeatureReference.
type FeatureReferenceResult struct {
	// Name is the name of the feature.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// Status represents the outcome of the feature reference operation specified in the FeatureGate spec
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Applied;Invalid
	// - Applied: represents feature toggle has been successfully applied.
	// - Invalid: represents that the intended state of the feature is invalid.
	Status FeatureReferenceStatus `json:"status"`
	// Message represents the reason for status
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:storageversion

// FeatureGate is the Schema for the featuregates API
type FeatureGate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification for gating features.
	Spec FeatureGateSpec `json:"spec,omitempty"`
	// Status reports activation state and availability of features in the system.
	Status FeatureGateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FeatureGateList contains a list of FeatureGate
type FeatureGateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureGate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FeatureGate{}, &FeatureGateList{})
}
