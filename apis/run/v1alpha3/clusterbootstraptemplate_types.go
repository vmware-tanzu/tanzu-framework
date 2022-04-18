// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterBootstrapTemplateSpec defines the desired state of ClusterBootstrapTemplate
type ClusterBootstrapTemplateSpec struct {
	// Paused can be used to prevent controllers from processing the ClusterBootstrap and all its associated objects.
	// +optional
	// +kubebuilder:default:=false
	Paused bool `json:"paused,omitempty"`

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

// ValuesFrom specifies how values for package install are retrieved from
// +kubebuilder:object:generate=false
type ValuesFrom struct {
	// +optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Inline map[string]interface{} `json:"inline,omitempty"`
	// +optional
	SecretRef string `json:"secretRef,omitempty"`
	// +optional
	ProviderRef *corev1.TypedLocalObjectReference `json:"providerRef,omitempty"`
}

func (in *ValuesFrom) CountFields() int {
	if in == nil {
		return 0
	}
	counterFunc := func(flag bool) int {
		if flag {
			return 1
		}
		return 0
	}
	return counterFunc(in.Inline != nil) + counterFunc(in.SecretRef != "") + counterFunc(in.ProviderRef != nil)
}

// +kubebuilder:object:root=true

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValuesFrom) DeepCopyInto(out *ValuesFrom) {
	*out = *in
	if in.Inline != nil {
		out.Inline = make(map[string]interface{}, len(in.Inline))
		refBytes, _ := json.Marshal(in.Inline)    // ignoring error: the original data is a JSON object
		_ = json.Unmarshal(refBytes, &out.Inline) // ignoring error: the original data is a JSON object
	}
	if in.ProviderRef != nil {
		in, out := &in.ProviderRef, &out.ProviderRef
		*out = new(corev1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new MachineImageInfo.
func (in *ValuesFrom) DeepCopy() *ValuesFrom {
	if in == nil {
		return nil
	}
	out := new(ValuesFrom)
	in.DeepCopyInto(out)
	return out
}

// +kubebuilder:object:root=true

// ClusterBootstrapTemplateList contains a list of ClusterBootstrapTemplate
type ClusterBootstrapTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBootstrapTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterBootstrapTemplate{}, &ClusterBootstrapTemplateList{})
}
