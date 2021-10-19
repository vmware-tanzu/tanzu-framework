// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import "github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"

// GCPDiscovery is a artifact discovery endpoing utilizing a GCP bucket.
type GCPDiscovery struct {
	bucketName   string
	manifestPath string
	name         string
}

// NewGCPDiscovery returns a new GCP bucket repository.
func NewGCPDiscovery(bucket, manifestPath, name string) Discovery {
	return &GCPDiscovery{
		bucketName:   bucket,
		manifestPath: manifestPath,
		name:         name,
	}
}

// List available plugins.
func (g *GCPDiscovery) List() (plugins []plugin.Discovered, err error) {
	// TODO: implement GCP discovery plugin list
	return
}

// Describe a plugin.
func (g *GCPDiscovery) Describe(name string) (p plugin.Discovered, err error) {
	// TODO: implement GCP discovery plugin describe
	return p, err
}

// Name of the repository.
func (g *GCPDiscovery) Name() string {
	return g.name
}

// Type of the discovery.
func (g *GCPDiscovery) Type() string {
	return "GCP"
}
