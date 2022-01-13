// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package discovery is implements discovery interface for plugin discovery
// Discovery is the interface to fetch the list of available plugins, their
// supported versions and how to download them either stand-alone or scoped to a server.
// A separate interface for discovery helps to decouple discovery (which is usually
// tied to a server or user identity) from distribution (which can be shared).
package discovery

import (
	"errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
)

// Discovery is the interface to fetch the list of available plugins
type Discovery interface {
	// Name of the repository.
	Name() string

	// List available plugins.
	List() ([]plugin.Discovered, error)

	// Describe a plugin.
	Describe(name string) (plugin.Discovered, error)

	// Type returns type of discovery.
	Type() string
}

// CreateDiscoveryFromV1alpha1 creates discovery interface from v1alpha1 API
func CreateDiscoveryFromV1alpha1(pd v1alpha1.PluginDiscovery) (Discovery, error) {
	switch {
	case pd.GCP != nil:
		return NewGCPDiscovery(pd.GCP.Bucket, pd.GCP.ManifestPath, pd.GCP.Name), nil
	case pd.OCI != nil:
		return NewOCIDiscovery(pd.OCI.Name, pd.OCI.Image), nil
	case pd.Local != nil:
		return NewLocalDiscovery(pd.Local.Name, pd.Local.Path), nil
	case pd.Kubernetes != nil:
		return NewKubernetesDiscovery(pd.Kubernetes.Name, pd.Kubernetes.Path, pd.Kubernetes.Context), nil
	case pd.REST != nil:
		return NewRESTDiscovery(pd.REST.Name, pd.REST.Endpoint, pd.REST.BasePath), nil
	}
	return nil, errors.New("unknown plugin discovery source")
}
