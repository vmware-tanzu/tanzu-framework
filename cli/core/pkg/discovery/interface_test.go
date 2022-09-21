// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func Test_CreateDiscoveryFromV1alpha1(t *testing.T) {
	assert := assert.New(t)

	// When no discovery type is provided, it should throw error
	pd := configapi.PluginDiscovery{}
	_, err := CreateDiscoveryFromV1alpha1(pd)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unknown plugin discovery source")

	// When OCI discovery is provided
	pd = configapi.PluginDiscovery{
		OCI: &configapi.OCIDiscovery{Name: "fake-oci", Image: "fake.repo.com/test:v1.0.0"},
	}
	discovery, err := CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeOCI, discovery.Type())
	assert.Equal("fake-oci", discovery.Name())

	// When Local discovery is provided
	pd = configapi.PluginDiscovery{
		Local: &configapi.LocalDiscovery{Name: "fake-local", Path: "test/path"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeLocal, discovery.Type())
	assert.Equal("fake-local", discovery.Name())

	// When GCP discovery is provided
	pd = configapi.PluginDiscovery{
		GCP: &configapi.GCPDiscovery{Name: "fake-gcp"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeGCP, discovery.Type())
	assert.Equal("fake-gcp", discovery.Name())

	// When K8s discovery is provided
	pd = configapi.PluginDiscovery{
		Kubernetes: &configapi.KubernetesDiscovery{Name: "fake-k8s"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeKubernetes, discovery.Type())
	assert.Equal("fake-k8s", discovery.Name())

	// When REST discovery is provided
	pd = configapi.PluginDiscovery{
		REST: &configapi.GenericRESTDiscovery{Name: "fake-rest"},
	}
	discovery, err = CreateDiscoveryFromV1alpha1(pd)
	assert.Nil(err)
	assert.Equal(common.DiscoveryTypeREST, discovery.Type())
	assert.Equal("fake-rest", discovery.Name())
}
