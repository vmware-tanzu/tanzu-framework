// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package data provides data types for the TKR Resolver.
package data

import (
	"k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// Query sets constraints for resolution of TKRs. Its structure reflects Cluster API cluster topology.
type Query struct {
	// ControlPlane specifies the Query for the control plane.
	ControlPlane OSImageQuery

	// ControlPlane specifies the OSImageQueries for worker machine deployments.
	MachineDeployments map[string]OSImageQuery
}

// OSImageQuery sets constraints for resolution of OSImages for the control plane or a machine deployment of a cluster.
type OSImageQuery struct {
	// K8sVersionPrefix is a version prefix for matching Kubernetes versions.
	K8sVersionPrefix string

	// TKRSelector is a label selector that resolved TKRs must satisfy.
	TKRSelector labels.Selector

	// OSImageSelector is a label selector that resolved OSImages must satisfy.
	OSImageSelector labels.Selector
}

// Result carries the results of TKR resolution. Its structure reflects Cluster API cluster topology.
type Result struct {
	// ControlPlane carries the Result for the  control plane.
	ControlPlane OSImageResult

	// ControlPlane carries the Result for worker machine deployments.
	MachineDeployments map[string]OSImageResult
}

// OSImageResult carries the results of OSImage resolution for the control plane or a machine deployment of a cluster.
type OSImageResult struct {
	// K8sVersion is the latest conforming K8s version. If empty, then no K8s version satisfied the query.
	K8sVersion string

	// TKRName is the latest conforming TKR name. If empty, then no TKRs satisfied the query.
	TKRName string

	// TKRsByK8sVersion maps resolved K8s versions to TKRs (sorted "latest first").
	TKRsByK8sVersion map[string]TKRs

	// OSImagesByTKR maps resolved TKR names to OSImages (unsorted).
	OSImagesByTKR map[string]OSImages
}

// TKRs is a set of TanzuKubernetesRelease objects implemented as a map tkr.Name -> tkr.
type TKRs map[string]*v1alpha3.TanzuKubernetesRelease

// OSImages is a set of OSImage objects implemented as a map osImage.Name -> osImage.
type OSImages map[string]*v1alpha3.OSImage

// Filter returns a subset of TKRs satisfying the predicate f.
func (tkrs TKRs) Filter(f func(tkr *v1alpha3.TanzuKubernetesRelease) bool) TKRs {
	result := make(TKRs, len(tkrs))
	for name, tkr := range tkrs {
		if f(tkr) {
			result[name] = tkr
		}
	}
	return result
}

// Filter returns a subset of OSImages satisfying the predicate f.
func (osImages OSImages) Filter(f func(osImage *v1alpha3.OSImage) bool) OSImages {
	result := make(OSImages, len(osImages))
	for name, osImage := range osImages {
		if f(osImage) {
			result[name] = osImage
		}
	}
	return result
}
