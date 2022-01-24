// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TanzuClusterBootstrapTemplateSpec defines the desired state of TanzuClusterBootstrapTemplate
type TanzuClusterBootstrapTemplateSpec struct {
	// TODO these are optional for temp testing, change when TKR v1alpha3 is available
	// +optional
	CNI *TanzuClusterBootstrapPackage `json:"cni,omitempty"`
	// +optional
	CSI *TanzuClusterBootstrapPackage `json:"csi,omitempty"`
	// +optional
	CPI *TanzuClusterBootstrapPackage `json:"cpi,omitempty"`
	// +optional
	Kapp *TanzuClusterBootstrapPackage `json:"kapp,omitempty"`
	// +optional
	AdditionalPackages []*TanzuClusterBootstrapPackage `json:"additionalPackages,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tanzuclusterbootstraptemplates,shortName=tcbt,scope=Namespaced
// +kubebuilder:printcolumn:name="CNI",type="string",JSONPath=".spec.cni.refName",description="CNI package name and version"
// +kubebuilder:printcolumn:name="CSI",type="string",JSONPath=".spec.csi.refName",description="CSI package name and version"
// +kubebuilder:printcolumn:name="CPI",type="string",JSONPath=".spec.cpi.refName",description="CPI package name and version"
// +kubebuilder:printcolumn:name="Kapp",type="string",JSONPath=".spec.kapp.refName",description="Kapp package name and version"
// +kubebuilder:printcolumn:name="Additional Packages",type="string",JSONPath=".spec.additionalPackages[*].refName",description="Additional packages",priority=10

// TanzuClusterBootstrapTemplate is the Schema for the tanzuclusterbootstraptemplates API
type TanzuClusterBootstrapTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *TanzuClusterBootstrapTemplateSpec `json:"spec"`
}

type TanzuClusterBootstrapPackage struct {
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

// TanzuClusterBootstrapTemplateList contains a list of TanzuClusterBootstrapTemplate
type TanzuClusterBootstrapTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuClusterBootstrapTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TanzuClusterBootstrapTemplate{}, &TanzuClusterBootstrapTemplateList{})
}
