// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:generate=false

// TanzuKubernetesCluster defines schema for TanzuKubernetesCluster
type TanzuKubernetesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   map[string]interface{} `json:"spec,omitempty"`
	Status map[string]interface{} `json:"status,omitempty"`
}

// TanzuKubernetesClusterList contains a list of TanzuKubernetesCluster
//
// +kubebuilder:object:root=true
type TanzuKubernetesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuKubernetesCluster `json:"items"`
}

// TODO: Remove this DeepCopy functions from this file once the full API is defined

// DeepCopyInto is an deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesCluster) DeepCopyInto(out *TanzuKubernetesCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = runtime.DeepCopyJSON(in.Spec)
	out.Status = runtime.DeepCopyJSON(in.Status)
}

// DeepCopy is an deepcopy function, copying the receiver, creating a new TanzuKubernetesCluster.
func (in *TanzuKubernetesCluster) DeepCopy() *TanzuKubernetesCluster {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TanzuKubernetesCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&TanzuKubernetesCluster{}, &TanzuKubernetesClusterList{})
}
