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
func (in *CSIConfig) DeepCopyInto(out *CSIConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CSIConfig.
func (in *CSIConfig) DeepCopy() *CSIConfig {
	if in == nil {
		return nil
	}
	out := new(CSIConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CSIConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CSIConfigList) DeepCopyInto(out *CSIConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CSIConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CSIConfigList.
func (in *CSIConfigList) DeepCopy() *CSIConfigList {
	if in == nil {
		return nil
	}
	out := new(CSIConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CSIConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CSIConfigSpec) DeepCopyInto(out *CSIConfigSpec) {
	*out = *in
	in.VSphereCSI.DeepCopyInto(&out.VSphereCSI)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CSIConfigSpec.
func (in *CSIConfigSpec) DeepCopy() *CSIConfigSpec {
	if in == nil {
		return nil
	}
	out := new(CSIConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CSIConfigStatus) DeepCopyInto(out *CSIConfigStatus) {
	*out = *in
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CSIConfigStatus.
func (in *CSIConfigStatus) DeepCopy() *CSIConfigStatus {
	if in == nil {
		return nil
	}
	out := new(CSIConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NonParavirtualConfig) DeepCopyInto(out *NonParavirtualConfig) {
	*out = *in
	if in.Region != nil {
		in, out := &in.Region, &out.Region
		*out = new(string)
		**out = **in
	}
	if in.Zone != nil {
		in, out := &in.Zone, &out.Zone
		*out = new(string)
		**out = **in
	}
	if in.UseTopologyCategories != nil {
		in, out := &in.UseTopologyCategories, &out.UseTopologyCategories
		*out = new(bool)
		**out = **in
	}
	if in.ProvisionTimeout != nil {
		in, out := &in.ProvisionTimeout, &out.ProvisionTimeout
		*out = new(string)
		**out = **in
	}
	if in.AttachTimeout != nil {
		in, out := &in.AttachTimeout, &out.AttachTimeout
		*out = new(string)
		**out = **in
	}
	if in.ResizerTimeout != nil {
		in, out := &in.ResizerTimeout, &out.ResizerTimeout
		*out = new(string)
		**out = **in
	}
	if in.VSphereVersion != nil {
		in, out := &in.VSphereVersion, &out.VSphereVersion
		*out = new(string)
		**out = **in
	}
	if in.HttpProxy != nil {
		in, out := &in.HttpProxy, &out.HttpProxy
		*out = new(string)
		**out = **in
	}
	if in.HttpsProxy != nil {
		in, out := &in.HttpsProxy, &out.HttpsProxy
		*out = new(string)
		**out = **in
	}
	if in.NoProxy != nil {
		in, out := &in.NoProxy, &out.NoProxy
		*out = new(string)
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
func (in *ParavirtualConfig) DeepCopyInto(out *ParavirtualConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ParavirtualConfig.
func (in *ParavirtualConfig) DeepCopy() *ParavirtualConfig {
	if in == nil {
		return nil
	}
	out := new(ParavirtualConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereCSI) DeepCopyInto(out *VSphereCSI) {
	*out = *in
	if in.ParavirtualConfig != nil {
		in, out := &in.ParavirtualConfig, &out.ParavirtualConfig
		*out = new(ParavirtualConfig)
		**out = **in
	}
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
