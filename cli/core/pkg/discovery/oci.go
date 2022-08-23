// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"strings"

	"github.com/pkg/errors"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
)

// OCIDiscovery is an artifact discovery endpoint utilizing OCI image
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
	return od.Manifest()
}

// Describe a plugin.
func (od *OCIDiscovery) Describe(name string) (p plugin.Discovered, err error) {
	plugins, err := od.Manifest()
	if err != nil {
		return
	}

	for i := range plugins {
		if plugins[i].Name == name {
			p = plugins[i]
			return
		}
	}
	err = errors.Errorf("cannot find plugin with name '%v'", name)
	return
}

// Name of the repository.
func (od *OCIDiscovery) Name() string {
	return od.name
}

// Type of the discovery.
func (od *OCIDiscovery) Type() string {
	return common.DiscoveryTypeOCI
}

// Manifest returns the manifest for a local repository.
func (od *OCIDiscovery) Manifest() ([]plugin.Discovered, error) {
	outputData, err := carvelhelpers.ProcessCarvelPackage(od.image)
	if err != nil {
		return nil, errors.Wrap(err, "error while processing package")
	}

	return processDiscoveryManifestData(outputData, od.name)
}

func processDiscoveryManifestData(data []byte, discoveryName string) ([]plugin.Discovered, error) {
	plugins := make([]plugin.Discovered, 0)

	for _, resourceYAML := range strings.Split(string(data), "---") {
		scheme, err := cliv1alpha1.SchemeBuilder.Build()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create scheme")
		}
		s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
			apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
		var p cliv1alpha1.CLIPlugin
		_, _, err = s.Decode([]byte(resourceYAML), nil, &p)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode discovery manifests")
		}

		dp, err := DiscoveredFromK8sV1alpha1(&p)
		if err != nil {
			return nil, err
		}
		dp.Source = discoveryName
		dp.DiscoveryType = common.DiscoveryTypeOCI
		if dp.Name != "" {
			plugins = append(plugins, dp)
		}
	}
	return plugins, nil
}
