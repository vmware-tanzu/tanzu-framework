//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
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
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
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
	in.AzureFileCSI.DeepCopyInto(&out.AzureFileCSI)
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
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(string)
		**out = **in
	}
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NonParavirtualConfig) DeepCopyInto(out *NonParavirtualConfig) {
	*out = *in
	if in.VSphereCredentialLocalObjRef != nil {
		in, out := &in.VSphereCredentialLocalObjRef, &out.VSphereCredentialLocalObjRef
		*out = new(v1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
	if in.InsecureFlag != nil {
		in, out := &in.InsecureFlag, &out.InsecureFlag
		*out = new(bool)
		**out = **in
	}
	if in.UseTopologyCategories != nil {
		in, out := &in.UseTopologyCategories, &out.UseTopologyCategories
		*out = new(bool)
		**out = **in
	}
	if in.DeploymentReplicas != nil {
		in, out := &in.DeploymentReplicas, &out.DeploymentReplicas
		*out = new(int32)
		**out = **in
	}
	if in.WindowsSupport != nil {
		in, out := &in.WindowsSupport, &out.WindowsSupport
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NonParavirtualConfig.
func (in *NonParavirtualConfig) DeepCopy() *NonParavirtualConfig {
	if in == nil {
		return nil
	}
	out := new(NonParavirtualConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereCSI) DeepCopyInto(out *VSphereCSI) {
	*out = *in
	if in.NonParavirtualConfig != nil {
		in, out := &in.NonParavirtualConfig, &out.NonParavirtualConfig
		*out = new(NonParavirtualConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereCSI.
func (in *VSphereCSI) DeepCopy() *VSphereCSI {
	if in == nil {
		return nil
	}
	out := new(VSphereCSI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereCSIConfig) DeepCopyInto(out *VSphereCSIConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereCSIConfig.
func (in *VSphereCSIConfig) DeepCopy() *VSphereCSIConfig {
	if in == nil {
		return nil
	}
	out := new(VSphereCSIConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereCSIConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereCSIConfigList) DeepCopyInto(out *VSphereCSIConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VSphereCSIConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereCSIConfigList.
func (in *VSphereCSIConfigList) DeepCopy() *VSphereCSIConfigList {
	if in == nil {
		return nil
	}
	out := new(VSphereCSIConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereCSIConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereCSIConfigSpec) DeepCopyInto(out *VSphereCSIConfigSpec) {
	*out = *in
	in.VSphereCSI.DeepCopyInto(&out.VSphereCSI)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereCSIConfigSpec.
func (in *VSphereCSIConfigSpec) DeepCopy() *VSphereCSIConfigSpec {
	if in == nil {
		return nil
	}
	out := new(VSphereCSIConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereCSIConfigStatus) DeepCopyInto(out *VSphereCSIConfigStatus) {
	*out = *in
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereCSIConfigStatus.
func (in *VSphereCSIConfigStatus) DeepCopy() *VSphereCSIConfigStatus {
	if in == nil {
		return nil
	}
	out := new(VSphereCSIConfigStatus)
	in.DeepCopyInto(out)
	return out
}
