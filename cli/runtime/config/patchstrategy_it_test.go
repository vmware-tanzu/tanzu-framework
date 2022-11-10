// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupConfigData() (string, string) {
	tanzuConfigBytes := `clientOptions:
  cli:
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
      - gcp:
          name: test2
          bucket: test-bucket2
          manifestPath: test-manifest-path2
          annotation: one
          required: true
      - local:
          name: test-local
          bucket: test-bucket2
          manifestPath: test-manifest-path2
          annotation: one
          required: true
servers:
  - name: test-mc
    type: managementcluster
    managementClusterOpts:
      endpoint: test-endpoint
      path: test-path
      context: test-context
      annotation: one
      required: true
current: test-mc
contexts:
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
currentContext:
  k8s: test-mc
`
	expectedConfig := `clientOptions:
    cli:
        discoverySources:
            - gcp:
                name: test
                bucket: updated-test-bucket
                manifestPath: updated-test-manifest-path
                annotation: one
            - gcp:
                name: test2
                bucket: test-bucket2
                manifestPath: test-manifest-path2
                annotation: one
                required: true
            - oci:
                name: test-local
                image: test-local-image-path
servers:
    - name: test-mc
      type: managementcluster
      managementClusterOpts:
        endpoint: test-endpoint
        path: test-path
        context: test-context
        annotation: one
        required: true
current: test-mc
contexts:
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
currentContext:
    k8s: test-mc
`
	return tanzuConfigBytes, expectedConfig
}

func setupConfigMetadata() string {
	metadata := `configMetadata:
  patchStrategy:
    contexts.group: replace
    contexts.clusterOpts.endpoint: replace
    contexts.clusterOpts.annotation: replace
    clientOptions.cli.discoverySources.gcp.required: replace
`
	return metadata
}

func TestIntegrationWithReplacePatchStrategy(t *testing.T) {
	//Setup data
	tanzuConfigBytes, expectedConfig := setupConfigData()
	metadata := setupConfigMetadata()

	// create temp config file
	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f1.Name())
	err = os.Setenv("TANZU_CONFIG", f1.Name())
	assert.NoError(t, err)

	//create temp config metadata file
	f2, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(metadata), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f2.Name())
	err = os.Setenv("TANZU_CONFIG_METADATA", f2.Name())
	assert.NoError(t, err)

	// Actions

	// Get CLI discovery sources
	expectedSources := []configapi.PluginDiscovery{
		{
			GCP: &configapi.GCPDiscovery{
				Name:         "test",
				Bucket:       "test-bucket",
				ManifestPath: "test-manifest-path",
			},
		},
		{
			GCP: &configapi.GCPDiscovery{
				Name:         "test2",
				Bucket:       "test-bucket2",
				ManifestPath: "test-manifest-path2",
			},
		},
		{
			Local: &configapi.LocalDiscovery{
				Name: "test-local",
			},
		},
	}

	sources, err := GetCLIDiscoverySources()
	assert.NoError(t, err)
	assert.Equal(t, expectedSources, sources)

	// Get CLI Discovery Source
	expectedSource := &configapi.PluginDiscovery{
		GCP: &configapi.GCPDiscovery{
			Name:         "test",
			Bucket:       "test-bucket",
			ManifestPath: "test-manifest-path",
		},
	}

	source, err := GetCLIDiscoverySource("test")
	assert.NoError(t, err)
	assert.Equal(t, expectedSource, source)

	// Update CLI discovery sources
	updatedSources := []configapi.PluginDiscovery{
		{
			GCP: &configapi.GCPDiscovery{
				Name:         "test",
				Bucket:       "updated-test-bucket",
				ManifestPath: "updated-test-manifest-path",
			},
		},
		{
			OCI: &configapi.OCIDiscovery{
				Name:  "test-local",
				Image: "test-local-image-path",
			},
		},
	}

	err = SetCLIDiscoverySources(updatedSources)
	assert.NoError(t, err)

	//Expectations on file content
	file, err := os.ReadFile(f1.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, string(file))
}
