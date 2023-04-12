// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReadinessProviderSpec defines the desired state of ReadinessProvider
type ReadinessProviderSpec struct {
	CheckRef                string                       `json:"checkName"`
	Repeatable              bool                         `json:"repeatable"`
	RepeatIntervalInSeconds int32                        `json:"repeatIntervalInSeconds"`
	Conditions              []ReadinessProviderCondition `json:"conditions"`
}

type ReadinessProviderCondition struct {
	Name              string                     `json:"name"`
	ResourceExistence ResourceExistenceCondition `json:"resourceExistence"`
}

type ResourceExistenceCondition struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// ReadinessProviderStatus defines the observed state of ReadinessProvider
type ReadinessProviderStatus struct {
	// +kubebuilder:validation:Enum=success;failure;inprogress
	State      string                     `json:"state"`
	Conditions []ReadinessConditionStatus `json:"conditions"`
}

type ReadinessConditionStatus struct {
	Name string `json:"name"`

	// +kubebuilder:validation:Enum=success;failure;inprogress
	State string `json:"state"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ReadinessProvider is the Schema for the readinessproviders API
type ReadinessProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReadinessProviderSpec   `json:"spec,omitempty"`
	Status ReadinessProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReadinessProviderList contains a list of ReadinessProvider
type ReadinessProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReadinessProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReadinessProvider{}, &ReadinessProviderList{})
}
