//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureFileCSI) DeepCopyInto(out *AzureFileCSI) {
	*out = *in
	if in.DeploymentReplicas != nil {
		in, out := &in.DeploymentReplicas, &out.DeploymentReplicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureFileCSI.
func (in *AzureFileCSI) DeepCopy() *AzureFileCSI {
	if in == nil {
		return nil
	}
	out := new(AzureFileCSI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureFileCSIConfig) DeepCopyInto(out *AzureFileCSIConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureFileCSIConfig.
func (in *AzureFileCSIConfig) DeepCopy() *AzureFileCSIConfig {
	if in == nil {
		return nil
	}
	out := new(AzureFileCSIConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AzureFileCSIConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureFileCSIConfigList) DeepCopyInto(out *AzureFileCSIConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AzureFileCSIConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureFileCSIConfigList.
func (in *AzureFileCSIConfigList) DeepCopy() *AzureFileCSIConfigList {
	if in == nil {
		return nil
	}
	out := new(AzureFileCSIConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AzureFileCSIConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureFileCSIConfigSpec) DeepCopyInto(out *AzureFileCSIConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureFileCSIConfigSpec.
func (in *AzureFileCSIConfigSpec) DeepCopy() *AzureFileCSIConfigSpec {
	if in == nil {
		return nil
	}
	out := new(AzureFileCSIConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureFileCSIConfigStatus) DeepCopyInto(out *AzureFileCSIConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureFileCSIConfigStatus.
func (in *AzureFileCSIConfigStatus) DeepCopy() *AzureFileCSIConfigStatus {
	if in == nil {
		return nil
	}
	out := new(AzureFileCSIConfigStatus)
	in.DeepCopyInto(out)
	return out
}
