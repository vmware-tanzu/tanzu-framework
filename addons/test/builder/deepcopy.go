// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// This file is manually generated because interface{} is not supported by deepcopy-gen.

package builder

import (
	v1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterBootstrapBuilder) DeepCopyInto(out *ClusterBootstrapBuilder) {
	*out = *in
	if in.cni != nil {
		in, out := &in.cni, &out.cni
		*out = new(v1alpha3.ClusterBootstrapPackage)
		(*in).DeepCopyInto(*out)
	}
	if in.csi != nil {
		in, out := &in.csi, &out.csi
		*out = new(v1alpha3.ClusterBootstrapPackage)
		(*in).DeepCopyInto(*out)
	}
	if in.cpi != nil {
		in, out := &in.cpi, &out.cpi
		*out = new(v1alpha3.ClusterBootstrapPackage)
		(*in).DeepCopyInto(*out)
	}
	if in.kapp != nil {
		in, out := &in.kapp, &out.kapp
		*out = new(v1alpha3.ClusterBootstrapPackage)
		(*in).DeepCopyInto(*out)
	}
	if in.additionalPackages != nil {
		in, out := &in.additionalPackages, &out.additionalPackages
		*out = make([]*v1alpha3.ClusterBootstrapPackage, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha3.ClusterBootstrapPackage)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new ClusterBootstrapBuilder.
func (in *ClusterBootstrapBuilder) DeepCopy() *ClusterBootstrapBuilder {
	if in == nil {
		return nil
	}
	out := new(ClusterBootstrapBuilder)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterBootstrapPackageBuilder) DeepCopyInto(out *ClusterBootstrapPackageBuilder) {
	*out = *in
	if in.inline != nil {
		in, out := &in.inline, &out.inline
		*out = make(map[string]interface{}, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.providerRef != nil {
		in, out := &in.providerRef, &out.providerRef
		*out = new(v1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new ClusterBootstrapPackageBuilder.
func (in *ClusterBootstrapPackageBuilder) DeepCopy() *ClusterBootstrapPackageBuilder {
	if in == nil {
		return nil
	}
	out := new(ClusterBootstrapPackageBuilder)
	in.DeepCopyInto(out)
	return out
}
