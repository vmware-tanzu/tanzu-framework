// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterBootstrapTemplateSpec defines the desired state of ClusterBootstrapTemplate
type ClusterBootstrapTemplateSpec struct {
	// Paused can be used to prevent controllers from processing the ClusterBootstrap and all its associated objects.
	// +optional
	// +kubebuilder:default:=false
	Paused bool `json:"paused,omitempty"`

	// TODO these are optional for temp testing, change when TKR v1alpha3 is available
	// +optional
	CNI *ClusterBootstrapPackage `json:"cni,omitempty"`
	// +optional
	CSI *ClusterBootstrapPackage `json:"csi,omitempty"`
	// +optional
	CPI *ClusterBootstrapPackage `json:"cpi,omitempty"`
	// +optional
	Kapp *ClusterBootstrapPackage `json:"kapp,omitempty"`
	// +optional
	AdditionalPackages []*ClusterBootstrapPackage `json:"additionalPackages,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clusterbootstraptemplates,shortName=cbt,scope=Namespaced
// +kubebuilder:printcolumn:name="CNI",type="string",JSONPath=".spec.cni.refName",description="CNI package name and version"
// +kubebuilder:printcolumn:name="CSI",type="string",JSONPath=".spec.csi.refName",description="CSI package name and version"
// +kubebuilder:printcolumn:name="CPI",type="string",JSONPath=".spec.cpi.refName",description="CPI package name and version"
// +kubebuilder:printcolumn:name="Kapp",type="string",JSONPath=".spec.kapp.refName",description="Kapp package name and version"
// +kubebuilder:printcolumn:name="Additional Packages",type="string",JSONPath=".spec.additionalPackages[*].refName",description="Additional packages",priority=10

// ClusterBootstrapTemplate is the Schema for the ClusterBootstraptemplates API
type ClusterBootstrapTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *ClusterBootstrapTemplateSpec `json:"spec"`
}

type ClusterBootstrapPackage struct {
	RefName string `json:"refName"`
	// +optional
	ValuesFrom *ValuesFrom `json:"valuesFrom,omitempty"`
}

type ValuesFrom struct {
	// +optional
	Inline string `json:"inline,omitempty"`
	// +optional
	SecretRef string `json:"secretRef,omitempty"`
	// +optional
	ProviderRef *corev1.TypedLocalObjectReference `json:"providerRef,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterBootstrapTemplateList contains a list of ClusterBootstrapTemplate
type ClusterBootstrapTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBootstrapTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterBootstrapTemplate{}, &ClusterBootstrapTemplateList{})
}
