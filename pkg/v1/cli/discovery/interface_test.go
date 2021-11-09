// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

func Test_CreateDiscoveryFromV1alpha1(t *testing.T) {
	assert := assert.New(t)

	// When no discovery type is provided, it should throw error
	pd := v1alpha1.PluginDiscovery{}
	_, err := CreateDiscoveryFromV1alpha1(pd)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unknown plugin discovery source")

	// When OCI discovery is provided
	pd = v1alpha1.PluginDiscovery{
		OCI: &v1alpha1.OCIDiscovery{Name: "fake-oci", Image: "fake.repo.com/test:v1.0.0"},
	}
	discovery, err := CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeOCI, discovery.Type())
	assert.Equal("fake-oci", discovery.Name())

	// When Local discovery is provided
	pd = v1alpha1.PluginDiscovery{
		Local: &v1alpha1.LocalDiscovery{Name: "fake-local", Path: "test/path"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeLocal, discovery.Type())
	assert.Equal("fake-local", discovery.Name())

	// When GCP discovery is provided
	pd = v1alpha1.PluginDiscovery{
		GCP: &v1alpha1.GCPDiscovery{Name: "fake-gcp"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeGCP, discovery.Type())
	assert.Equal("fake-gcp", discovery.Name())

	// When K8s discovery is provided
	pd = v1alpha1.PluginDiscovery{
		Kubernetes: &v1alpha1.KubernetesDiscovery{Name: "fake-k8s"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeKubernetes, discovery.Type())
	assert.Equal("fake-k8s", discovery.Name())
}
