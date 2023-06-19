// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReadinessProviderState defines the current state of the provider
type ReadinessProviderState string

const (
	// ProviderSuccessState is a ReadinessProviderState that denotes success state
	ProviderSuccessState = ReadinessProviderState("success")

	// ProviderFailureState is a ReadinessProviderState that denotes failure state
	ProviderFailureState = ReadinessProviderState("failure")

	// ProviderInProgressState is a ReadinessProviderState that denotes in-progress state
	ProviderInProgressState = ReadinessProviderState("inprogress")
)

// ReadinessConditionState defines the state of indvidual conditions in a readiness provider
type ReadinessConditionState string

const (
	// ConditionSuccessState is a ReadinessConditionState that denotes success state
	ConditionSuccessState = ReadinessConditionState("success")

	// ConditionFailureState is a ReadinessConditionState that denotes failure state
	ConditionFailureState = ReadinessConditionState("failure")

	// ConditionInProgressState is a ReadinessConditionState that denotes in-progress state
	ConditionInProgressState = ReadinessConditionState("inprogress")
)

// ReadinessProviderSpec defines the desired state of ReadinessProvider
type ReadinessProviderSpec struct {
	// CheckRefs contains names of the checks that the current provider satisfies
	CheckRefs []string `json:"checkRefs"`

	// Conditions is the set of checks that must be evaluated to true to mark the provider as ready
	Conditions []ReadinessProviderCondition `json:"conditions"`

	// ServiceAccount represents the service account to be used
	// to make requests to the API server for evaluating conditions.
	// If not provided, it uses the default service account
	// of the readiness provider controller.
	//+kubebuilder:validation:Optional
	ServiceAccount *ServiceAccountSource `json:"serviceAccount"`
}

type ServiceAccountSource struct {
	// Namespace is the namespace containing the service account.
	Namespace string `json:"namespace"`

	// Name is the name of the service account to be used
	// to make requests to the API server for evaluating conditions.
	Name string `json:"name"`
}

// ReadinessProviderCondition defines the readiness provider condition
type ReadinessProviderCondition struct {
	// Name is the name of the condition
	Name string `json:"name"`

	// ResourceExistenceCondition is the condition that checks for the presence of a certain resource in the cluster
	//+kubebuilder:validation:Optional
	ResourceExistenceCondition *ResourceExistenceCondition `json:"resourceExistenceCondition"`
}

// ResourceExistenceCondition is a type of readiness provider condition that checks for existence of given resource
type ResourceExistenceCondition struct {
	// APIVersion is the API version of the resource that is being checked.
	// This should be provided in <group>/<version> format.
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources
	APIVersion string `json:"apiVersion"`

	// Kind is the API kind of the resource that is being checked
	// More info: More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind"`

	// Namespace is the namespace of the resource that is being checked; if the Namespace is nil,
	// the resource is assumed to be cluster scoped. Empty string for the namespace will throw error.
	//+kubebuilder:validation:Optional
	Namespace *string `json:"namespace"`
	Name      string  `json:"name"`
}

// ReadinessProviderStatus defines the observed state of ReadinessProvider
type ReadinessProviderStatus struct {
	// State is the computed state of the provider. The state will be success if all the conditions pass;
	// The state will be failure if any of the conditions fail. Otherwise, the state will be in-progress.
	// +kubebuilder:validation:Enum=success;failure;inprogress
	State ReadinessProviderState `json:"state"`

	// Message provides information about the ReadinessProvider state
	Message string `json:"message"`

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
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Checks",priority=1,type=string,JSONPath=`.spec.checkRefs`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
