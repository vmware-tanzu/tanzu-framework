// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupData() (string, string) {
	tanzuConfigBytes := `clientOptions:
  cli:
    discoverySources:
      - oci:
          name: default
          image: "/:"
          unknown: cli-unknown
        contextType: k8s
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
        contextType: tmc
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
      endpoint: updated-test-endpoint
      path: updated-test-path
      context: updated-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: updated-test-bucket
          manifestPath: updated-test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test-two
          bucket: updated-test-bucket
          manifestPath: updated-test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  k8s: test-mc
`
	expectedConfig := `clientOptions:
    cli:
        discoverySources:
            - oci:
                name: default
                image: "update-default-image"
                unknown: cli-unknown
              contextType: k8s
            - local:
                name: admin-local
                path: admin
            - oci:
                name: new-default
                image: new-default-image
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
          contextType: tmc
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
        endpoint: updated-test-endpoint
        path: updated-test-path
        context: updated-test-context
      discoverySources:
        - gcp:
            name: test
            bucket: updated-test-bucket
            manifestPath: updated-test-manifest-path
            annotation: one
            required: true
          contextType: tmc
        - gcp:
            name: test-two
            bucket: updated-test-bucket
            manifestPath: updated-test-manifest-path
            annotation: two
            required: true
          contextType: tmc
currentContext:
    k8s: test-mc
`
	return tanzuConfigBytes, expectedConfig
}

func TestCLIDiscoverySourceIntegration(t *testing.T) {
	// Setup Data and Test config file
	tanzuConfigBytes, expectedConfig := setupData()
	f, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f.Name())
	err = os.Setenv("TANZU_CONFIG", f.Name())
	assert.NoError(t, err)

	// Get CLI DiscoverySources
	sources, err := GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sources))

	// Add new OCI CLI DiscoverySource
	ds := &configapi.PluginDiscovery{
		OCI: &configapi.OCIDiscovery{
			Name:  "new-default",
			Image: "new-default-image",
		},
	}
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	sources, err = GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(sources))
	// Should not persist on Adding same OCI CLI DiscoverySource
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	sources, err = GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(sources))

	// Update existing OCI CLI DiscoverySource
	ds = &configapi.PluginDiscovery{
		OCI: &configapi.OCIDiscovery{
			Name:  "default",
			Image: "update-default-image",
		},
	}
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	source, err := GetCLIDiscoverySource("default")
	assert.Nil(t, err)
	assert.NotNil(t, source)
	assert.Equal(t, ds.OCI, source.OCI)
	file, err := os.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte(expectedConfig), file)

	// Delete existing DiscoverySource
	err = DeleteCLIDiscoverySource("new-default")
	assert.NoError(t, err)
	sources, err = GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sources))
}

func setupDataWithPatchStrategy() (string, string) {
	tanzuConfigBytes := `clientOptions:
  cli:
    discoverySources:
      - oci:
          name: default
          image: "/:"
          unknown: cli-unknown
          annotation: new-annotation
        contextType: k8s
      - local:
          name: admin-local
          path: admin
`
	expectedConfig := `clientOptions:
    cli:
        discoverySources:
            - oci:
                name: default
                image: "update-default-image"
                unknown: cli-unknown
              contextType: k8s
            - local:
                name: admin-local
                path: admin
            - oci:
                name: new-default
                image: new-default-image
`
	return tanzuConfigBytes, expectedConfig
}

func TestCLIDiscoverySourceIntegrationWithPatchStrategy(t *testing.T) {
	// Setup Data and Test config file
	tanzuConfigBytes, expectedConfig := setupDataWithPatchStrategy()
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

	//Setup metadata
	metadata := `configMetadata:
  patchStrategy:
    clientOptions.cli.discoverySources.oci.annotation: replace`
	f2, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(metadata), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f2.Name())
	err = os.Setenv(EnvConfigMetadataKey, f2.Name())
	assert.NoError(t, err)

	// Get CLI DiscoverySources
	sources, err := GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sources))
	// Add new OCI CLI DiscoverySource
	ds := &configapi.PluginDiscovery{
		OCI: &configapi.OCIDiscovery{
			Name:  "new-default",
			Image: "new-default-image",
		},
	}
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	sources, err = GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(sources))

	// Should not persist on Adding same OCI CLI DiscoverySource
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	sources, err = GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(sources))

	// Update existing OCI CLI DiscoverySource
	ds = &configapi.PluginDiscovery{
		OCI: &configapi.OCIDiscovery{
			Name:  "default",
			Image: "update-default-image",
		},
	}
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	source, err := GetCLIDiscoverySource("default")
	assert.Nil(t, err)
	assert.NotNil(t, source)
	assert.Equal(t, ds.OCI, source.OCI)
	file, err := os.ReadFile(f1.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte(expectedConfig), file)
}
