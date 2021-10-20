// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
)

// OCIDiscovery is a artifact discovery endpoint utilizing OCI image
type OCIDiscovery struct {
	// name is a name of the discovery
	name string
	// image is an OCI compliant image. Which include DNS-compatible registry name,
	// a valid URI path(MAY contain zero or more ‘/’) and a valid tag
	// E.g., harbor.my-domain.local/tanzu-cli/plugins-manifest:latest
	// Contains a directory containing YAML files, each of which contains single
	// CLIPlugin API resource.
	image string
}

// NewOCIDiscovery returns a new local repository.
func NewOCIDiscovery(name, image string) Discovery {
	return &OCIDiscovery{
		name:  name,
		image: image,
	}
}

// List available plugins.
func (od *OCIDiscovery) List() (plugins []plugin.Discovered, err error) {
	// TODO(anujc25): implement OCI discovery plugin list
	return
}

// Describe a plugin.
func (od *OCIDiscovery) Describe(name string) (p plugin.Discovered, err error) {
	// TODO(anujc25): implement OCI discovery plugin describe
	return
}

// Name of the repository.
func (od *OCIDiscovery) Name() string {
	return od.name
}

// Type of the discovery.
func (od *OCIDiscovery) Type() string {
	return "OCI"
}
