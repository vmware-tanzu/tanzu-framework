// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func Test_CheckDiscoveryName(t *testing.T) {
	assert := assert.New(t)

	gcpDiscovery := configapi.PluginDiscovery{GCP: &configapi.GCPDiscovery{Name: "gcp-test"}}
	result := CheckDiscoveryName(gcpDiscovery, "gcp-test")
	assert.True(result)
	result = CheckDiscoveryName(gcpDiscovery, "test")
	assert.False(result)

	ociDiscovery := configapi.PluginDiscovery{OCI: &configapi.OCIDiscovery{Name: "oci-test"}}
	result = CheckDiscoveryName(ociDiscovery, "oci-test")
	assert.True(result)
	result = CheckDiscoveryName(ociDiscovery, "test")
	assert.False(result)

	k8sDiscovery := configapi.PluginDiscovery{Kubernetes: &configapi.KubernetesDiscovery{Name: "k8s-test"}}
	result = CheckDiscoveryName(k8sDiscovery, "k8s-test")
	assert.True(result)
	result = CheckDiscoveryName(k8sDiscovery, "test")
	assert.False(result)

	localDiscovery := configapi.PluginDiscovery{Local: &configapi.LocalDiscovery{Name: "local-test"}}
	result = CheckDiscoveryName(localDiscovery, "local-test")
	assert.True(result)
	result = CheckDiscoveryName(localDiscovery, "test")
	assert.False(result)

	restDiscovery := configapi.PluginDiscovery{REST: &configapi.GenericRESTDiscovery{Name: "rest-test"}}
	result = CheckDiscoveryName(restDiscovery, "rest-test")
	assert.True(result)
	result = CheckDiscoveryName(restDiscovery, "test")
	assert.False(result)
}

func Test_CompareDiscoverySource(t *testing.T) {
	assert := assert.New(t)

	ds1 := configapi.PluginDiscovery{GCP: &configapi.GCPDiscovery{Name: "gcp-test", Bucket: "bucket1", ManifestPath: "manifest1"}}
	ds2 := configapi.PluginDiscovery{GCP: &configapi.GCPDiscovery{Name: "gcp-test", Bucket: "bucket1", ManifestPath: "manifest1"}}
	result := CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeGCP)
	assert.True(result)
	ds2 = configapi.PluginDiscovery{GCP: &configapi.GCPDiscovery{Name: "gcp-test", Bucket: "bucket2", ManifestPath: "manifest1"}}
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeGCP)
	assert.False(result)

	ds1 = configapi.PluginDiscovery{Local: &configapi.LocalDiscovery{Name: "gcp-test", Path: "path1"}}
	ds2 = configapi.PluginDiscovery{Local: &configapi.LocalDiscovery{Name: "gcp-test", Path: "path1"}}
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeLocal)
	assert.True(result)
	ds2 = configapi.PluginDiscovery{Local: &configapi.LocalDiscovery{Name: "gcp-test", Path: "path2"}}
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeLocal)
	assert.False(result)

	ds1 = configapi.PluginDiscovery{OCI: &configapi.OCIDiscovery{Name: "gcp-test", Image: "image1"}}
	ds2 = configapi.PluginDiscovery{OCI: &configapi.OCIDiscovery{Name: "gcp-test", Image: "image1"}}
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeOCI)
	assert.True(result)
	ds2 = configapi.PluginDiscovery{OCI: &configapi.OCIDiscovery{Name: "gcp-test", Image: "image2"}}
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeOCI)
	assert.False(result)

	ds1 = configapi.PluginDiscovery{OCI: &configapi.OCIDiscovery{Name: "gcp-test", Image: "image1"}}
	ds2 = configapi.PluginDiscovery{Local: &configapi.LocalDiscovery{Name: "gcp-test", Path: "path1"}}
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeOCI)
	assert.False(result)
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeLocal)
	assert.False(result)
	result = CompareDiscoverySource(ds1, ds2, common.DiscoveryTypeREST)
	assert.False(result)
}

func TestSortVersion(t *testing.T) {
	tcs := []struct {
		name string
		act  []string
		exp  []string
		err  error
	}{
		{
			name: "Success",
			act:  []string{"v1.0.0", "v0.0.1", "0.0.1-dev"},
			exp:  []string{"0.0.1-dev", "v0.0.1", "v1.0.0"},
		},
		{
			name: "Success",
			act:  []string{"1.0.0", "0.0.a"},
			exp:  []string{"1.0.0", "0.0.a"},
			err:  semver.ErrInvalidSemVer,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SortVersions(tc.act)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.exp, tc.act)
		})
	}
}
