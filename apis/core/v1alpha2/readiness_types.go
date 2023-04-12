// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReadinessSpec defines the desired state of Readiness
type ReadinessSpec struct {
	Checks []Check `json:"checks"`
}

type Check struct {
	Name string `json:"name"`

	// +kubebuilder:validation:Enum=basic;composite
	Type string `json:"type"`

	Category string `json:"category"`
}

// ReadinessStatus defines the observed state of Readiness
type ReadinessStatus struct {
	CheckStatus      []CheckStatus `json:"checkStatus"`
	Ready            bool          `json:"ready"`
	LastComputedTime metav1.Time   `json:"lastComputedTime"`
}

type CheckStatus struct {
	Name            string   `json:"name"`
	Ready           bool     `json:"status"`
	Providers       []string `json:"providers"`
	ActiveProviders []string `json:"activeProviders"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Readiness is the Schema for the readinesses API
type Readiness struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReadinessSpec   `json:"spec,omitempty"`
	Status ReadinessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReadinessList contains a list of Readiness
type ReadinessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Readiness `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Readiness{}, &ReadinessList{})
}
