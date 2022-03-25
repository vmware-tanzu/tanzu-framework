// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

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

// ClusterBootstrapStatus defines the observed state of ClusterBootstrap
type ClusterBootstrapStatus struct {
	ResolvedTKR string `json:"resolvedTKR,omitempty"`

	Conditions clusterapiv1beta1.Conditions `json:"conditions,omitempty"`
}

// GetConditions returns the set of conditions for this object. implements Setter interface
func (c *ClusterBootstrap) GetConditions() clusterapiv1beta1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object. implements Setter interface
func (c *ClusterBootstrap) SetConditions(conditions clusterapiv1beta1.Conditions) {
	c.Status.Conditions = conditions
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
