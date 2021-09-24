// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

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
}

// FeatureGateSpec defines the desired state of FeatureGate
type FeatureGateSpec struct {
	// NamespaceSelector is a selector to specify namespaces for which this feature gate applies.
	// Use an empty LabelSelector to match all namespaces.
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector"`
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
	// Namespaces lists the existing namespaces for which this feature gate applies. This is obtained from listing all
	// namespaces and applying the NamespaceSelector specified in spec.
	Namespaces []string `json:"namespaces,omitempty"`
	// ActivatedFeatures lists the discovered features that are activated for the namespaces specified in the spec.
	// This can include features that are not explicitly gated in the spec, but are already available in the system as
	// Feature resources.
	ActivatedFeatures []string `json:"activatedFeatures,omitempty"`
	// DeactivatedFeatures lists the discovered features that are deactivated for the namespaces specified in the spec.
	// This can include features that are not explicitly gated in the spec, but are already available in the system as
	// Feature resources.
	DeactivatedFeatures []string `json:"deactivatedFeatures,omitempty"`
	// UnavailableFeatures lists the features that are gated in the spec, but are not available in the system as
	// Feature resources.
	UnavailableFeatures []string `json:"unavailableFeatures,omitempty"`
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
