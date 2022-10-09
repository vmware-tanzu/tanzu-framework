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
	// SkipStabilityValidation lets you skip the validation performed when activating an unstable feature.
	// Once set to true, cannot be set back to false
	SkipStabilityValidation bool `json:"skipStabilityValidation,omitempty"`
}

// FeatureGateSpec defines the desired state of FeatureGate
type FeatureGateSpec struct {
	// Features is a slice of FeatureReference to gate features.
	// The Feature resource specified may or may not be present in the system. If the Feature is present, the
	// FeatureGate controller and webhook sets the specified activation state only if the Feature is discoverable and
	// its immutability constraint is satisfied. If the Feature is not present, the activation intent is applied when
	// the Feature resource appears in the system. The actual activation state of the Feature is reported in the status.
	// +listType=map
	// +listMapKey=name
	Features []FeatureReference `json:"features,omitempty"`
}

// FeatureGateStatus defines the observed state of FeatureGate
type FeatureGateStatus struct {
	// Results represents the results of all the features specified in the FeatureGate spec.
	// +listType=map
	// +listMapKey=name
	Results []Result `json:"results"`
}

// Result represents the result of Feature.
type Result struct {
	// Name is the name of the feature.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// Status represents the outcome of the feature operation specified in the spec
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=applied;no-op;invalid
	// - applied: represents feature toggle has been successfully applied.
	// - no-op: represent no operation has been done, feature is already in the intended state.
	// - invalid: represents that the intended state of the feature is invalid.
	Status string `json:"status"`
	// Message represents the reason for invalid status
	// +optional
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// FeatureGate is the Schema for the featuregates API
type FeatureGate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification for gating features.
	Spec FeatureGateSpec `json:"spec,omitempty"`
	// Status reports activation state and availability of features in the system.
	Status FeatureGateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeatureGateList contains a list of FeatureGate
type FeatureGateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureGate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FeatureGate{}, &FeatureGateList{})
}
