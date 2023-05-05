// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReadinessSpec defines the desired state of Readiness
type ReadinessSpec struct {
	// Checks is the set of checks that are required to mark the readiness
	Checks []Check `json:"checks"`
}

type Check struct {
	// Name is the name of the check
	Name string `json:"name"`

	// Type is the type of the check. the Type can be either basic or composite
	// The basic checks depend on its providers to be ready
	// The composite checks depend on the basic checks for their readiness
	// +kubebuilder:validation:Enum=basic;composite
	Type string `json:"type"`

	// Category is the category of the check. Examples of catagories are availability and security.
	Category string `json:"category"`
}

// ReadinessStatus defines the observed state of Readiness
type ReadinessStatus struct {
	// CheckStatus presents the status of check defined in the spec
	CheckStatus []CheckStatus `json:"checkStatus"`

	// Ready is the flag that denotes if the defined readiness is ready
	// The readiness is marked ready if all the checks are satisfied
	// The time at which this field is evaluated is given by LastComputedTime
	Ready bool `json:"ready"`

	// LastUpdatedTime is the field that denotes the time at which the readiness is updated.
	LastUpdatedTime *metav1.Time `json:"lastComputedTime"`
}

type CheckStatus struct {
	// Name is the name of the check
	Name string `json:"name"`

	// Ready is the boolean flag indicating if the check is ready
	Ready bool `json:"status"`

	// Providers is the list of providers available for the given check
	Providers []Provider `json:"providers"`
}

type Provider struct {
	// Name is the name of the provider
	Name string `json:"name"`

	// IsActive is the boolean flag indicating if the provider is active
	IsActive bool `json:"isActive"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
