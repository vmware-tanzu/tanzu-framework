// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapabilitySpec defines the desired state of Capability.
type CapabilitySpec struct {
	// ServiceAccountName is the name of the service account with which requests
	// are made to the API server for evaluating queries.
	// Service account should exist in the same namespace as this resource.
	// When this field is not specified, a default service account with only the
	// ability to execute GVR queries is used.
	// +optional
	ServiceAccountName string `json:"serviceAccountName"`
	// Queries specifies set of queries that are evaluated.
	// +listType=map
	// +listMapKey=name
	Queries []Query `json:"queries"`
}

// Query is a logical grouping of GVR, Object and PartialSchema queries.
type Query struct {
	// Name is the unique name of the query.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// GroupVersionResources evaluates a slice of GVR queries.
	// +listType=map
	// +listMapKey=name
	// +optional
	GroupVersionResources []QueryGVR `json:"groupVersionResources,omitempty"`
	// Objects evaluates a slice of Object queries.
	// +listType=map
	// +listMapKey=name
	// +optional
	Objects []QueryObject `json:"objects,omitempty"`
	// PartialSchemas evaluates a slice of PartialSchema queries.
	// +listType=map
	// +listMapKey=name
	// +optional
	PartialSchemas []QueryPartialSchema `json:"partialSchemas,omitempty"`
}

// QueryObject represents any runtime.Object that could exist in a cluster with the ability to check for annotations.
type QueryObject struct {
	// Name is the unique name of the query.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// ObjectReference is the ObjectReference to check for in the cluster.
	// +kubebuilder:validation:Required
	ObjectReference corev1.ObjectReference `json:"objectReference"`
	// WithAnnotations are the annotations whose presence is checked in the object.
	// The query succeeds only if all the annotations specified exists.
	// +optional
	WithAnnotations map[string]string `json:"withAnnotations,omitempty"`
	// WithAnnotations are the annotations whose absence is checked in the object.
	// The query succeeds only if all the annotations specified do not exist.
	// +optional
	WithoutAnnotations map[string]string `json:"withoutAnnotations,omitempty"`
}

// QueryGVR queries for an API group with the optional ability to check for API versions and resource.
type QueryGVR struct {
	// Name is the unique name of the query.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// Group is the API group to check for in the cluster.
	// +optional
	Group string `json:"group"`
	// Versions is the slice of versions to check for in the specified API group.
	// +optional
	Versions []string `json:"versions,omitempty"`
	// Resource is the API resource to check for given an API group and a slice of versions.
	// Specifying a Resource requires at least one version to be specified in Versions.
	// +optional
	Resource string `json:"resource,omitempty"`
}

// QueryPartialSchema queries for any OpenAPI schema that may exist on a cluster.
type QueryPartialSchema struct {
	// Name is the unique name of the query.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// PartialSchema is the partial OpenAPI schema that will be matched in a cluster.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	PartialSchema string `json:"partialSchema"`
}

// CapabilityStatus defines the observed state of Capability
type CapabilityStatus struct {
	// Results represents the results of all the queries specified in the spec.
	// +listType=map
	// +listMapKey=name
	Results []Result `json:"results"`
}

// QueryResult represents the result of a single query.
type QueryResult struct {
	// Name is the name of the query in spec whose result this struct represents.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// Found is a boolean which indicates if the query condition succeeded.
	// +optional
	Found bool `json:"found"`
	// Error indicates if an error occurred while processing the query.
	// +optional
	Error bool `json:"error,omitempty"`
	// ErrorDetail represents the error detail, if an error occurred.
	// +optional
	ErrorDetail string `json:"errorDetail,omitempty"`
	// NotFoundReason provides the reason if the query condition fails.
	// This is non-empty when Found is false.
	// +optional
	NotFoundReason string `json:"notFoundReason,omitempty"`
}

// Result represents the results of queries in Query.
type Result struct {
	// Name is the unique name of the query.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
	// GroupVersionResources represents results of GVR queries in spec.
	// +listType=map
	// +listMapKey=name
	// +optional
	GroupVersionResources []QueryResult `json:"groupVersionResources,omitempty"`
	// Objects represents results of Object queries in spec.
	// +listType=map
	// +listMapKey=name
	// +optional
	Objects []QueryResult `json:"objects,omitempty"`
	// PartialSchemas represents results of PartialSchema queries in spec.
	// +listType=map
	// +listMapKey=name
	// +optional
	PartialSchemas []QueryResult `json:"partialSchemas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Capability is the Schema for the capabilities API
type Capability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the capability spec that has cluster queries.
	Spec CapabilitySpec `json:"spec,omitempty"`
	// Status is the capability status that has results of cluster queries.
	Status CapabilityStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CapabilityList contains a list of Capability
type CapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Capability `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Capability{}, &CapabilityList{})
}
