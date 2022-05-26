// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CalicoConfigSpec defines the desired state of CalicoConfig.
type CalicoConfigSpec struct {
	Calico Calico `json:"calico,omitempty"`
}

// Calico stores the configurations for Calico.
type Calico struct {
	Config CalicoConfigDataValue `json:"config,omitempty"`
}

// CalicoConfigDataValue contains the specific configurations for the Calico package.
type CalicoConfigDataValue struct {
	// VethMTU defines maximum transmission unit. "0" as default means MTU will be auto detected.
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=0
	VethMTU int64 `json:"vethMTU,omitempty"`

	// SkipCNIBinaries allows to skip the cni plugin binaries installation.
	// Default to false. Set to true for providers who already installed
	// cni plugin binaries in their OVAs and do not want Calico to overwrite them.
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	SkipCNIBinaries bool `json:"skipCNIBinaries,omitempty"`
}

// CalicoConfigStatus defines the observed state of CalicoConfig.
type CalicoConfigStatus struct {
	// SecretRef is the name of the data value secret created by calico controller.
	//+ kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=calicoconfigs,shortName=calicoconf,scope=Namespaced
//+kubebuilder:printcolumn:name="VethMTU",type="integer",JSONPath=".spec.calico.config.vethMTU",description="Maximum transmission unit setting. '0' as default means MTU will be auto detected"
//+kubebuilder:printcolumn:name="SecretRef",type="string",JSONPath=".status.secretRef",description="Name of the Calico data values secret"

// CalicoConfig is the Schema for the calicoconfigs API.
type CalicoConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CalicoConfigSpec   `json:"spec"`
	Status CalicoConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CalicoConfigList contains a list of CalicoConfig.
type CalicoConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CalicoConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CalicoConfig{}, &CalicoConfigList{})
}
