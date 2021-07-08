// +build !ignore_autogenerated

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIEndpoint) DeepCopyInto(out *APIEndpoint) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIEndpoint.
func (in *APIEndpoint) DeepCopy() *APIEndpoint {
	if in == nil {
		return nil
	}
	out := new(APIEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddonStatus) DeepCopyInto(out *AddonStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddonStatus.
func (in *AddonStatus) DeepCopy() *AddonStatus {
	if in == nil {
		return nil
	}
	out := new(AddonStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddonsStatuses) DeepCopyInto(out *AddonsStatuses) {
	*out = *in
	if in.DNS != nil {
		in, out := &in.DNS, &out.DNS
		*out = new(AddonStatus)
		**out = **in
	}
	if in.Proxy != nil {
		in, out := &in.Proxy, &out.Proxy
		*out = new(AddonStatus)
		**out = **in
	}
	if in.PSP != nil {
		in, out := &in.PSP, &out.PSP
		*out = new(AddonStatus)
		**out = **in
	}
	if in.CNI != nil {
		in, out := &in.CNI, &out.CNI
		*out = new(AddonStatus)
		**out = **in
	}
	if in.CSI != nil {
		in, out := &in.CSI, &out.CSI
		*out = new(AddonStatus)
		**out = **in
	}
	if in.CloudProvider != nil {
		in, out := &in.CloudProvider, &out.CloudProvider
		*out = new(AddonStatus)
		**out = **in
	}
	if in.AuthService != nil {
		in, out := &in.AuthService, &out.AuthService
		*out = new(AddonStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddonsStatuses.
func (in *AddonsStatuses) DeepCopy() *AddonsStatuses {
	if in == nil {
		return nil
	}
	out := new(AddonsStatuses)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CNIConfiguration) DeepCopyInto(out *CNIConfiguration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CNIConfiguration.
func (in *CNIConfiguration) DeepCopy() *CNIConfiguration {
	if in == nil {
		return nil
	}
	out := new(CNIConfiguration)
	in.DeepCopyInto(out)
	return out
}

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
	in.Queries.DeepCopyInto(&out.Queries)
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
	in.Results.DeepCopyInto(&out.Results)
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
func (in *ClusterAPIStatus) DeepCopyInto(out *ClusterAPIStatus) {
	*out = *in
	if in.APIEndpoints != nil {
		in, out := &in.APIEndpoints, &out.APIEndpoints
		*out = make([]APIEndpoint, len(*in))
		copy(*out, *in)
	}
	if in.ErrorReason != nil {
		in, out := &in.ErrorReason, &out.ErrorReason
		*out = new(ClusterAPIStatusError)
		**out = **in
	}
	if in.ErrorMessage != nil {
		in, out := &in.ErrorMessage, &out.ErrorMessage
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAPIStatus.
func (in *ClusterAPIStatus) DeepCopy() *ClusterAPIStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterAPIStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerImage) DeepCopyInto(out *ContainerImage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerImage.
func (in *ContainerImage) DeepCopy() *ContainerImage {
	if in == nil {
		return nil
	}
	out := new(ContainerImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Distribution) DeepCopyInto(out *Distribution) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Distribution.
func (in *Distribution) DeepCopy() *Distribution {
	if in == nil {
		return nil
	}
	out := new(Distribution)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Error) DeepCopyInto(out *Error) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Error.
func (in *Error) DeepCopy() *Error {
	if in == nil {
		return nil
	}
	out := new(Error)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Network) DeepCopyInto(out *Network) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = new(NetworkRanges)
		(*in).DeepCopyInto(*out)
	}
	if in.Pods != nil {
		in, out := &in.Pods, &out.Pods
		*out = new(NetworkRanges)
		(*in).DeepCopyInto(*out)
	}
	if in.CNI != nil {
		in, out := &in.CNI, &out.CNI
		*out = new(CNIConfiguration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Network.
func (in *Network) DeepCopy() *Network {
	if in == nil {
		return nil
	}
	out := new(Network)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkRanges) DeepCopyInto(out *NetworkRanges) {
	*out = *in
	if in.CIDRBlocks != nil {
		in, out := &in.CIDRBlocks, &out.CIDRBlocks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkRanges.
func (in *NetworkRanges) DeepCopy() *NetworkRanges {
	if in == nil {
		return nil
	}
	out := new(NetworkRanges)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Queries) DeepCopyInto(out *Queries) {
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Queries.
func (in *Queries) DeepCopy() *Queries {
	if in == nil {
		return nil
	}
	out := new(Queries)
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
func (in *Results) DeepCopyInto(out *Results) {
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Results.
func (in *Results) DeepCopy() *Results {
	if in == nil {
		return nil
	}
	out := new(Results)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Settings) DeepCopyInto(out *Settings) {
	*out = *in
	if in.Network != nil {
		in, out := &in.Network, &out.Network
		*out = new(Network)
		(*in).DeepCopyInto(*out)
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(Storage)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Settings.
func (in *Settings) DeepCopy() *Settings {
	if in == nil {
		return nil
	}
	out := new(Settings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Storage) DeepCopyInto(out *Storage) {
	*out = *in
	if in.Classes != nil {
		in, out := &in.Classes, &out.Classes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Storage.
func (in *Storage) DeepCopy() *Storage {
	if in == nil {
		return nil
	}
	out := new(Storage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesCluster) DeepCopyInto(out *TanzuKubernetesCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesCluster.
func (in *TanzuKubernetesCluster) DeepCopy() *TanzuKubernetesCluster {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TanzuKubernetesCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesClusterList) DeepCopyInto(out *TanzuKubernetesClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TanzuKubernetesCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesClusterList.
func (in *TanzuKubernetesClusterList) DeepCopy() *TanzuKubernetesClusterList {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TanzuKubernetesClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesClusterSpec) DeepCopyInto(out *TanzuKubernetesClusterSpec) {
	*out = *in
	out.Topology = in.Topology
	out.Distribution = in.Distribution
	if in.Settings != nil {
		in, out := &in.Settings, &out.Settings
		*out = new(Settings)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesClusterSpec.
func (in *TanzuKubernetesClusterSpec) DeepCopy() *TanzuKubernetesClusterSpec {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesClusterStatus) DeepCopyInto(out *TanzuKubernetesClusterStatus) {
	*out = *in
	if in.ClusterAPIStatus != nil {
		in, out := &in.ClusterAPIStatus, &out.ClusterAPIStatus
		*out = new(ClusterAPIStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.VMStatus != nil {
		in, out := &in.VMStatus, &out.VMStatus
		*out = make(map[string]VirtualMachineState, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.NodeStatus != nil {
		in, out := &in.NodeStatus, &out.NodeStatus
		*out = make(map[string]NodeStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Upgrade != nil {
		in, out := &in.Upgrade, &out.Upgrade
		*out = new(UpgradeStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.Addons != nil {
		in, out := &in.Addons, &out.Addons
		*out = new(AddonsStatuses)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesClusterStatus.
func (in *TanzuKubernetesClusterStatus) DeepCopy() *TanzuKubernetesClusterStatus {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesRelease) DeepCopyInto(out *TanzuKubernetesRelease) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesRelease.
func (in *TanzuKubernetesRelease) DeepCopy() *TanzuKubernetesRelease {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesRelease)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TanzuKubernetesRelease) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesReleaseList) DeepCopyInto(out *TanzuKubernetesReleaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TanzuKubernetesRelease, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesReleaseList.
func (in *TanzuKubernetesReleaseList) DeepCopy() *TanzuKubernetesReleaseList {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesReleaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TanzuKubernetesReleaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesReleaseSpec) DeepCopyInto(out *TanzuKubernetesReleaseSpec) {
	*out = *in
	if in.Images != nil {
		in, out := &in.Images, &out.Images
		*out = make([]ContainerImage, len(*in))
		copy(*out, *in)
	}
	if in.NodeImageRef != nil {
		in, out := &in.NodeImageRef, &out.NodeImageRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesReleaseSpec.
func (in *TanzuKubernetesReleaseSpec) DeepCopy() *TanzuKubernetesReleaseSpec {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesReleaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TanzuKubernetesReleaseStatus) DeepCopyInto(out *TanzuKubernetesReleaseStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1alpha3.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TanzuKubernetesReleaseStatus.
func (in *TanzuKubernetesReleaseStatus) DeepCopy() *TanzuKubernetesReleaseStatus {
	if in == nil {
		return nil
	}
	out := new(TanzuKubernetesReleaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Topology) DeepCopyInto(out *Topology) {
	*out = *in
	out.ControlPlane = in.ControlPlane
	out.Workers = in.Workers
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Topology.
func (in *Topology) DeepCopy() *Topology {
	if in == nil {
		return nil
	}
	out := new(Topology)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TopologySettings) DeepCopyInto(out *TopologySettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TopologySettings.
func (in *TopologySettings) DeepCopy() *TopologySettings {
	if in == nil {
		return nil
	}
	out := new(TopologySettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpgradeStatus) DeepCopyInto(out *UpgradeStatus) {
	*out = *in
	if in.Errors != nil {
		in, out := &in.Errors, &out.Errors
		*out = make([]Error, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpgradeStatus.
func (in *UpgradeStatus) DeepCopy() *UpgradeStatus {
	if in == nil {
		return nil
	}
	out := new(UpgradeStatus)
	in.DeepCopyInto(out)
	return out
}
