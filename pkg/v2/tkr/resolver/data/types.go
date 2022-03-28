// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package data provides data types for the TKR Resolver.
package data

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// Query sets constraints for resolution of TKRs. Its structure reflects Cluster API cluster topology.
type Query struct {
	// ControlPlane specifies the Query for the control plane.
	// Set to nil if we want to skip resolving the control plane part.
	ControlPlane *OSImageQuery

	// MachineDeployments specifies the OSImageQueries for worker machine deployments.
	// An individual machine deployment query part may be set to nil if we want to skip resolving it.
	MachineDeployments []*OSImageQuery
}

func (q Query) String() string {
	return fmt.Sprintf("{controlPlane: %s, machineDeployments: %s}", q.ControlPlane, q.MachineDeployments)
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

const strNil = "nil"

func (q *OSImageQuery) String() string {
	if q == nil {
		return strNil
	}
	return fmt.Sprintf("{k8sVersionPrefix: '%s', tkrSelector: '%s', osImageSelector: '%s'}", q.K8sVersionPrefix, q.TKRSelector, q.OSImageSelector)
}

// Result carries the results of TKR resolution. Its structure reflects Cluster API cluster topology.
type Result struct {
	// ControlPlane carries the Result for the  control plane.
	// It is set to nil if resolving the control plane part was skipped.
	ControlPlane *OSImageResult

	// ControlPlane carries the Result for worker machine deployments.
	// An individual machine deployment result is set to nil if resolving it was skipped.
	MachineDeployments []*OSImageResult
}

func (r Result) String() string {
	return fmt.Sprintf("{controlPlane: %s, machineDeployments: %s}", r.ControlPlane, r.MachineDeployments)
}

// OSImageResult carries the results of OSImage resolution for the control plane or a machine deployment of a cluster.
type OSImageResult struct {
	// K8sVersion is the latest conforming K8s version. If empty, then no K8s version satisfied the query.
	K8sVersion string

	// TKRName is the latest conforming TKR name. If empty, then no TKRs satisfied the query.
	TKRName string

	// TKRsByK8sVersion maps resolved K8s versions to TKRs.
	TKRsByK8sVersion map[string]TKRs

	// OSImagesByTKR maps resolved TKR names to OSImages.
	OSImagesByTKR map[string]OSImages
}

func (r *OSImageResult) String() string {
	if r == nil {
		return strNil
	}
	return fmt.Sprintf("{k8sVersion: '%s', tkrName: '%s', osImagesByTKR: %s}", r.K8sVersion, r.TKRName, r.OSImagesByTKR)
}

// TKRs is a set of TanzuKubernetesRelease objects implemented as a map tkr.Name -> tkr.
type TKRs map[string]*runv1.TanzuKubernetesRelease

// OSImages is a set of OSImage objects implemented as a map osImage.Name -> osImage.
type OSImages map[string]*runv1.OSImage

// Filter returns a subset of TKRs satisfying the predicate f.
func (tkrs TKRs) Filter(f func(tkr *runv1.TanzuKubernetesRelease) bool) TKRs {
	result := make(TKRs, len(tkrs))
	for name, tkr := range tkrs {
		if f(tkr) {
			result[name] = tkr
		}
	}
	return result
}

// Filter returns a subset of OSImages satisfying the predicate f.
func (osImages OSImages) Filter(f func(osImage *runv1.OSImage) bool) OSImages {
	result := make(OSImages, len(osImages))
	for name, osImage := range osImages {
		if f(osImage) {
			result[name] = osImage
		}
	}
	return result
}

func (osImages OSImages) String() string {
	names := make([]string, 0, len(osImages))
	for name := range osImages {
		names = append(names, name)
	}
	return fmt.Sprintf("%v", names)
}
