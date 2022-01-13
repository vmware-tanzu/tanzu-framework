// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"sort"

	"github.com/Masterminds/semver"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

// CheckDiscoveryName returns true if discovery name exists else return false
func CheckDiscoveryName(ds configv1alpha1.PluginDiscovery, dn string) bool {
	return (ds.GCP != nil && ds.GCP.Name == dn) ||
		(ds.Kubernetes != nil && ds.Kubernetes.Name == dn) ||
		(ds.Local != nil && ds.Local.Name == dn) ||
		(ds.REST != nil && ds.REST.Name == dn) ||
		(ds.OCI != nil && ds.OCI.Name == dn)
}

// CompareDiscoverySource returns true if both discovery source are same for the given type
func CompareDiscoverySource(ds1, ds2 configv1alpha1.PluginDiscovery, dsType string) bool {
	switch dsType {
	case common.DiscoveryTypeLocal:
		return compareLocalDiscoverySources(ds1, ds2)

	case common.DiscoveryTypeOCI:
		return compareOCIDiscoverySources(ds1, ds2)

	case common.DiscoveryTypeKubernetes:
		return compareK8sDiscoverySources(ds1, ds2)

	case common.DiscoveryTypeGCP:
		return compareGCPDiscoverySources(ds1, ds2)

	case common.DiscoveryTypeREST:
		return compareRESTDiscoverySources(ds1, ds2)
	}
	return false
}

func compareGCPDiscoverySources(ds1, ds2 configv1alpha1.PluginDiscovery) bool {
	return ds1.GCP != nil && ds2.GCP != nil &&
		ds1.GCP.Name == ds2.GCP.Name &&
		ds1.GCP.Bucket == ds2.GCP.Bucket &&
		ds1.GCP.ManifestPath == ds2.GCP.ManifestPath
}

func compareLocalDiscoverySources(ds1, ds2 configv1alpha1.PluginDiscovery) bool {
	return ds1.Local != nil && ds2.Local != nil &&
		ds1.Local.Name == ds2.Local.Name &&
		ds1.Local.Path == ds2.Local.Path
}

func compareOCIDiscoverySources(ds1, ds2 configv1alpha1.PluginDiscovery) bool {
	return ds1.OCI != nil && ds2.OCI != nil &&
		ds1.OCI.Name == ds2.OCI.Name &&
		ds1.OCI.Image == ds2.OCI.Image
}

func compareK8sDiscoverySources(ds1, ds2 configv1alpha1.PluginDiscovery) bool {
	return ds1.Kubernetes != nil && ds2.Kubernetes != nil &&
		ds1.Kubernetes.Name == ds2.Kubernetes.Name &&
		ds1.Kubernetes.Path == ds2.Kubernetes.Path &&
		ds1.Kubernetes.Context == ds2.Kubernetes.Context
}

func compareRESTDiscoverySources(ds1, ds2 configv1alpha1.PluginDiscovery) bool {
	return ds1.REST != nil && ds2.REST != nil &&
		ds1.REST.Name == ds2.REST.Name &&
		ds1.REST.BasePath == ds2.REST.BasePath &&
		ds1.REST.Endpoint == ds2.REST.Endpoint
}

// SortVersions sorts the supported version strings in semver 2.0 order.
func SortVersions(vStrArr []string) error {
	vArr := make([]*semver.Version, len(vStrArr))
	for i, vStr := range vStrArr {
		v, err := semver.NewVersion(vStr)
		if err != nil {
			return err
		}
		vArr[i] = v
	}
	sort.Sort(semver.Collection(vArr))
	for i, v := range vArr {
		vStrArr[i] = v.Original()
	}
	return nil
}
