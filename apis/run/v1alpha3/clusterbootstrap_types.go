// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterBootstrapStatus defines the observed state of ClusterBootstrap
type ClusterBootstrapStatus struct {
	ResolvedTKR string `json:"resolvedTKR,omitempty"`

	Conditions Conditions `json:"conditions,omitempty"`
}

type Conditions []Condition
type ConditionType string

type Condition struct {
	// Type of condition in CamelCase
	Type ConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown
	Status corev1.ConditionStatus `json:"status"`

	// A human readable error message. Will be empty in case no error happened
	UsefulErrorMessage string `json:"usefulErrorMessage,omitempty"`

	// Last time the condition transitioned from one status to another.
	// It reflects the time any other fields in the condition changed
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterbootstraps,shortName=cb,scope=Namespaced
// +kubebuilder:printcolumn:name="CNI",type="string",JSONPath=".spec.cni.refName",description="CNI package name and version"
// +kubebuilder:printcolumn:name="CSI",type="string",JSONPath=".spec.csi.refName",description="CSI package name and version"
// +kubebuilder:printcolumn:name="CPI",type="string",JSONPath=".spec.cpi.refName",description="CPI package name and version"
// +kubebuilder:printcolumn:name="Kapp",type="string",JSONPath=".spec.kapp.refName",description="Kapp package name and version"
// +kubebuilder:printcolumn:name="Additional Packages",type="string",JSONPath=".spec.additionalPackages[*].refName",description="Additional packages",priority=10
// +kubebuilder:printcolumn:name="Resolved_TKR",type="string",JSONPath=".status.resolvedTKR",description="Resolved TKR name"

// ClusterBootstrap is the Schema for the ClusterBootstraps API
type ClusterBootstrap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *ClusterBootstrapTemplateSpec `json:"spec"`
	Status ClusterBootstrapStatus        `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterBootstrapList contains a list of ClusterBootstrap
type ClusterBootstrapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBootstrap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterBootstrap{}, &ClusterBootstrapList{})
}
