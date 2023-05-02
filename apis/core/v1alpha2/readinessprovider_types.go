// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReadinessProviderState string

const (
	ProviderSuccessState    = ReadinessProviderState("success")
	ProviderFailureState    = ReadinessProviderState("failure")
	ProviderInProgressState = ReadinessProviderState("inprogress")
)

type ReadinessConditionState string

const (
	ConditionSuccessState    = ReadinessConditionState("success")
	ConditionFailureState    = ReadinessConditionState("failure")
	ConditionInProgressState = ReadinessConditionState("inprogress")
)

// ReadinessProviderSpec defines the desired state of ReadinessProvider
type ReadinessProviderSpec struct {
	// CheckRef is the name of the check that the current provider satisfies
	CheckRefs []string `json:"checkRefs"`

	// RepeatInterval is the re-evaluation interval;
	// if RepeatInterval is not provided or nil,
	// the provider will be evaluated only once and the evaluation will not be repeated.
	// A valid value specifies a duration string, such as "1.5h", "1h10m" or "200s".
	// Refer: https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration
	//+kubebuilder:validation:Optional
	RepeatInterval *metav1.Duration `json:"repeatInterval"`

	// Conditions is the set of checks that must be evaluated to true to mark the provider as ready
	Conditions []ReadinessProviderCondition `json:"conditions"`
}

type ReadinessProviderCondition struct {
	// Name is the name of the condition
	Name string `json:"name"`

	// ResourceExistenceCondition is the condition that checks for the presence of a certain resource in the cluster
	ResourceExistenceCondition *ResourceExistenceCondition `json:"resourceExistenceCondition"`
}

type ResourceExistenceCondition struct {
	// APIVersion is the API version of the resource that is being checked
	APIVersion string `json:"apiVersion"`

	// Kind is the API kind of the resource that is being checked
	Kind string `json:"kind"`

	// Namespace is the namespace of the resource that is being checked; if the Namespace is nil,
	// the resource is assumed to be cluster scoped. Empty string for the namespace will throw error
	//+kubebuilder:validation:Optional
	Namespace *string `json:"namespace"`
	Name      string  `json:"name"`
}

// ReadinessProviderStatus defines the observed state of ReadinessProvider
type ReadinessProviderStatus struct {
	// State is the computed state of the provider. The state will be success if all the coditions pass;
	// The state will be failure if any of the conditions fail. Otherwise, the state will be in-progress
	// +kubebuilder:validation:Enum=success;failure;inprogress
	State ReadinessProviderState `json:"state"`

	// Conditions is the set of ReadinessConditions that are being evaluated
	Conditions []ReadinessConditionStatus `json:"conditions"`
}

type ReadinessConditionStatus struct {
	// Name is the name of the readiness condition
	Name string `json:"name"`

	// State is the computed state of the condition
	// +kubebuilder:validation:Enum=success;failure;inprogress
	State ReadinessConditionState `json:"state"`

	// Message is the field that provides information about the condition evaluation
	Message string `json:"message"`
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
