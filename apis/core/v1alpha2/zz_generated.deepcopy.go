//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Capability) DeepCopyInto(out *Capability) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Capability.
func (in *Capability) DeepCopy() *Capability {
	if in == nil {
		return nil
	}
	out := new(Capability)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Capability) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapabilityList) DeepCopyInto(out *CapabilityList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Capability, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapabilityList.
func (in *CapabilityList) DeepCopy() *CapabilityList {
	if in == nil {
		return nil
	}
	out := new(CapabilityList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CapabilityList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapabilitySpec) DeepCopyInto(out *CapabilitySpec) {
	*out = *in
	if in.Queries != nil {
		in, out := &in.Queries, &out.Queries
		*out = make([]Query, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapabilitySpec.
func (in *CapabilitySpec) DeepCopy() *CapabilitySpec {
	if in == nil {
		return nil
	}
	out := new(CapabilitySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapabilityStatus) DeepCopyInto(out *CapabilityStatus) {
	*out = *in
	if in.Results != nil {
		in, out := &in.Results, &out.Results
		*out = make([]Result, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapabilityStatus.
func (in *CapabilityStatus) DeepCopy() *CapabilityStatus {
	if in == nil {
		return nil
	}
	out := new(CapabilityStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Query) DeepCopyInto(out *Query) {
	*out = *in
	if in.GroupVersionResources != nil {
		in, out := &in.GroupVersionResources, &out.GroupVersionResources
		*out = make([]QueryGVR, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Objects != nil {
		in, out := &in.Objects, &out.Objects
		*out = make([]QueryObject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PartialSchemas != nil {
		in, out := &in.PartialSchemas, &out.PartialSchemas
		*out = make([]QueryPartialSchema, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Query.
func (in *Query) DeepCopy() *Query {
	if in == nil {
		return nil
	}
	out := new(Query)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QueryGVR) DeepCopyInto(out *QueryGVR) {
	*out = *in
	if in.Versions != nil {
		in, out := &in.Versions, &out.Versions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QueryGVR.
func (in *QueryGVR) DeepCopy() *QueryGVR {
	if in == nil {
		return nil
	}
	out := new(QueryGVR)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QueryObject) DeepCopyInto(out *QueryObject) {
	*out = *in
	out.ObjectReference = in.ObjectReference
	if in.WithAnnotations != nil {
		in, out := &in.WithAnnotations, &out.WithAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.WithoutAnnotations != nil {
		in, out := &in.WithoutAnnotations, &out.WithoutAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QueryObject.
func (in *QueryObject) DeepCopy() *QueryObject {
	if in == nil {
		return nil
	}
	out := new(QueryObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QueryPartialSchema) DeepCopyInto(out *QueryPartialSchema) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QueryPartialSchema.
func (in *QueryPartialSchema) DeepCopy() *QueryPartialSchema {
	if in == nil {
		return nil
	}
	out := new(QueryPartialSchema)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QueryResult) DeepCopyInto(out *QueryResult) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QueryResult.
func (in *QueryResult) DeepCopy() *QueryResult {
	if in == nil {
		return nil
	}
	out := new(QueryResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Result) DeepCopyInto(out *Result) {
	*out = *in
	if in.GroupVersionResources != nil {
		in, out := &in.GroupVersionResources, &out.GroupVersionResources
		*out = make([]QueryResult, len(*in))
		copy(*out, *in)
	}
	if in.Objects != nil {
		in, out := &in.Objects, &out.Objects
		*out = make([]QueryResult, len(*in))
		copy(*out, *in)
	}
	if in.PartialSchemas != nil {
		in, out := &in.PartialSchemas, &out.PartialSchemas
		*out = make([]QueryResult, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Result.
func (in *Result) DeepCopy() *Result {
	if in == nil {
		return nil
	}
	out := new(Result)
	in.DeepCopyInto(out)
	return out
}