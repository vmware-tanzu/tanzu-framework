// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupContextsData() (string, string, string, string) {
	//nolint:goconst
	cfg := `clientOptions:
  cli:
    discoverySources:
      - oci:
          name: default
          image: "/:"
          unknown: cli-unknown
      - local:
          name: default-local
      - local:
          name: admin-local
          path: admin
servers:
  - name: test-mc
    type: managementcluster
    managementClusterOpts:
      endpoint: updated-test-endpoint
      path: updated-test-path
      context: updated-test-context
      annotation: one
      required: true
    discoverySources:
      - gcp:
          name: test
          bucket: updated-test-bucket
          manifestPath: updated-test-manifest-path
          annotation: one
          required: true
current: test-mc
`
	expectedCfg := `clientOptions:
    cli:
        discoverySources:
            - oci:
                name: default
                image: "/:"
                unknown: cli-unknown
            - local:
                name: default-local
            - local:
                name: admin-local
                path: admin
servers:
    - name: test-mc
      type: managementcluster
      managementClusterOpts:
        endpoint: updated-test-endpoint
        path: updated-test-path
        context: updated-test-context
        annotation: one
        required: true
      discoverySources:
        - gcp:
            name: test
            bucket: updated-test-bucket
            manifestPath: updated-test-manifest-path
            annotation: one
            required: true
    - name: test-mc2
      type: managementcluster
      managementClusterOpts:
        path: test-path-updated
        context: test-context-updated
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path-updated
current: test-mc2
contexts: []
currentContext: {}
`

	//nolint:goconst
	cfg2 := `contexts:
  - name: test-mc
    type: k8s
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: test-endpoint
      path: test-path
      context: test-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
currentContext:
  k8s: test-mc
`
	expectedCfg2 := `contexts:
    - name: test-mc
      type: k8s
      group: one
      clusterOpts:
        isManagementCluster: true
        annotation: one
        required: true
        annotationStruct:
            one: one
        endpoint: test-endpoint
        path: test-path
        context: test-context
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket
            manifestPath: test-manifest-path
            annotation: one
            required: true
    - name: test-mc2
      type: k8s
      clusterOpts:
        path: test-path-updated
        context: test-context-updated
        isManagementCluster: true
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path-updated
currentContext:
    k8s: test-mc2
`

	return cfg, expectedCfg, cfg2, expectedCfg2
}
func TestContextsIntegration(t *testing.T) {
	// Setup config data
	cfg, expectedCfg, cfg2, expectedCfg2 := setupContextsData()

	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(cfg), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, f1.Name())
	assert.NoError(t, err)

	f2, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(cfg2), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigNextGenKey, f2.Name())
	assert.NoError(t, err)

	//Setup metadata
	fMeta, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(fMeta.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigMetadataKey, fMeta.Name())
	assert.NoError(t, err)

	// Cleanup
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f1.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f2.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(fMeta.Name())

	// Get Context
	context, err := GetContext("test-mc")
	expected := &configapi.Context{
		Name: "test-mc",
		Type: "k8s",
		ClusterOpts: &configapi.ClusterServer{
			Endpoint:            "test-endpoint",
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, context)
	// Add new Context
	newCtx := &configapi.Context{
		Name: "test-mc2",
		Type: "k8s",
		ClusterOpts: &configapi.ClusterServer{
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}
	err = SetContext(newCtx, true)
	assert.NoError(t, err)
	ctx, err := GetContext("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, newCtx, ctx)
	// Update existing Context
	updatedCtx := &configapi.Context{
		Name: "test-mc2",
		Type: "k8s",
		ClusterOpts: &configapi.ClusterServer{
			Path:                "test-path-updated",
			Context:             "test-context-updated",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket-updated",
					ManifestPath: "test-manifest-path-updated",
				},
			},
		},
	}
	err = SetContext(updatedCtx, true)
	assert.NoError(t, err)
	ctx, err = GetContext("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, updatedCtx, ctx)

	file, err := os.ReadFile(f1.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg, string(file))

	file, err = os.ReadFile(f2.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg2, string(file))

	// Delete context
	err = DeleteContext("test-mc2")
	assert.NoError(t, err)
	ctx, err = GetContext("test-mc2")
	assert.Equal(t, "context test-mc2 not found", err.Error())
	assert.Nil(t, ctx)
}
