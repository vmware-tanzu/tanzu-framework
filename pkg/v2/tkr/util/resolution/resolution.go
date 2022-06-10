// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package resolution provides helper functions for TKR resolution, e.g. ConstructQuery().
package resolution

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
)

// ConstructQuery returns data.Query to be used with the TKR Resolver.
func ConstructQuery(versionPrefix string, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass) (*data.Query, error) {
	tkrSelector, err := selectorFromAnnotation(cluster.Annotations, clusterClass.Annotations, runv1.AnnotationResolveTKR)
	if tkrSelector == nil || cluster.Spec.Topology == nil {
		return nil, err // err may be nil too
	}

	osImageSelector, err := selectorFromAnnotation(
		cluster.Spec.Topology.ControlPlane.Metadata.Annotations,
		clusterClass.Spec.ControlPlane.Metadata.Annotations,
		runv1.AnnotationResolveOSImage)
	if err != nil {
		return nil, err
	}
	if osImageSelector == nil {
		osImageSelector = labels.Everything() // default to empty selector (matches all) for OSImages
	}

	cpQuery := &data.OSImageQuery{
		K8sVersionPrefix: versionPrefix,
		TKRSelector:      tkrSelector,
		OSImageSelector:  osImageSelector,
	}

	if cluster.Spec.Topology.Workers == nil {
		return &data.Query{ControlPlane: cpQuery}, nil
	}

	mdQueries, err := constructMDOSImageQueries(versionPrefix, cluster, clusterClass, tkrSelector)
	if err != nil {
		return nil, err
	}

	return &data.Query{ControlPlane: cpQuery, MachineDeployments: mdQueries}, nil
}

// selectorFromAnnotation produces a selector from the value of the specified annotation.
func selectorFromAnnotation(cAnnots, ccAnnots map[string]string, annotation string) (labels.Selector, error) {
	var selectorStr *string
	if selectorStr = getAnnotation(cAnnots, annotation); selectorStr == nil {
		if selectorStr = getAnnotation(ccAnnots, annotation); selectorStr == nil {
			return nil, nil
		}
	}
	selector, err := labels.Parse(*selectorStr)
	return selector, errors.Wrapf(err, "error parsing selector: '%s'", *selectorStr)
}

// getAnnotation gets the value of the annotation specified by name.
// Returns nil if such annotation is not found.
func getAnnotation(annotations map[string]string, name string) *string {
	if annotations == nil {
		return nil
	}
	value, exists := annotations[name]
	if !exists {
		return nil
	}
	return &value
}

func constructMDOSImageQueries(versionPrefix string, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass, tkrSelector labels.Selector) ([]*data.OSImageQuery, error) {
	mdOSImageQueries := make([]*data.OSImageQuery, len(cluster.Spec.Topology.Workers.MachineDeployments))

	for i := range cluster.Spec.Topology.Workers.MachineDeployments {
		md := &cluster.Spec.Topology.Workers.MachineDeployments[i]
		mdClass := getMDClass(clusterClass, md.Class)
		if mdClass == nil {
			return nil, errors.Errorf("machineDeployment refers to non-existent MD class '%s'", md.Class)
		}
		osImageSelector, err := selectorFromAnnotation(
			md.Metadata.Annotations,
			mdClass.Template.Metadata.Annotations,
			runv1.AnnotationResolveOSImage)
		if err != nil {
			return nil, err
		}
		if osImageSelector == nil {
			osImageSelector = labels.Everything() // default to empty selector (matches all)
		}

		mdOSImageQueries[i] = &data.OSImageQuery{
			K8sVersionPrefix: versionPrefix,
			TKRSelector:      tkrSelector,
			OSImageSelector:  osImageSelector,
		}
	}
	return mdOSImageQueries, nil
}

func getMDClass(clusterClass *clusterv1.ClusterClass, mdClassName string) *clusterv1.MachineDeploymentClass {
	for i := range clusterClass.Spec.Workers.MachineDeployments {
		md := &clusterClass.Spec.Workers.MachineDeployments[i]
		if md.Class == mdClassName {
			return md
		}
	}
	return nil
}
