// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupData() (string, string, string, string) {
	CFG := `clientOptions:
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
      endpoint: test-endpoint
      path: test-path
      context: test-context
      annotation: one
      required: true
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
current: test-mc
contexts:
  - name: test-mc
    target: kubernetes
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
      - gcp:
          name: test-two
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
`

	CFG2 := `contexts:
  - name: test-mc
    target: kubernetes
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
      - gcp:
          name: test-two
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
`

	expectedCFG := `clientOptions:
    cli:
        discoverySources:
            - oci:
                name: default
                image: "default-image"
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
        endpoint: test-endpoint
        path: test-path
        context: test-context
        annotation: one
        required: true
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket
            manifestPath: test-manifest-path
            annotation: one
            required: true
          contextType: tmc
current: test-mc
contexts:
    - name: test-mc
      target: kubernetes
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
        - gcp:
            name: test-two
            bucket: test-bucket
            manifestPath: test-manifest-path
            annotation: two
            required: true
          contextType: tmc
currentContext:
    kubernetes: test-mc
`
	//nolint:goconst
	expectedCFG2 := `contexts:
    - name: test-mc
      target: kubernetes
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
        - gcp:
            name: test-two
            bucket: test-bucket
            manifestPath: test-manifest-path
            annotation: two
            required: true
          contextType: tmc
currentContext:
    kubernetes: test-mc
`

	return CFG, expectedCFG, CFG2, expectedCFG2
}

func TestCLIDiscoverySourceIntegration(t *testing.T) {
	// Setup config data
	cfg, expectedCfg, cfg2, expectedCfg2 := setupData()
	cfgFiles, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

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
			Image: "default-image",
		},
	}
	err = SetCLIDiscoverySource(*ds)
	assert.NoError(t, err)
	source, err := GetCLIDiscoverySource("default")
	assert.Nil(t, err)
	assert.NotNil(t, source)
	assert.Equal(t, ds.OCI, source.OCI)

	file, err := os.ReadFile(cfgFiles[0].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg, string(file))

	file, err = os.ReadFile(cfgFiles[1].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg2, string(file))

	// Delete existing DiscoverySource
	err = DeleteCLIDiscoverySource("new-default")
	assert.NoError(t, err)
	sources, err = GetCLIDiscoverySources()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sources))
}

func setupDataWithPatchStrategy() (string, string, string, string) {
	cfg := `clientOptions:
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
	expectedCfg := `clientOptions:
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

	expectedCfg2 := `{}
`

	return cfg, expectedCfg, "", expectedCfg2
}

func TestCLIDiscoverySourceIntegrationWithPatchStrategy(t *testing.T) {
	// Setup config data
	cfg, expectedCfg, cfg2, expectedCfg2 := setupDataWithPatchStrategy()
	metadata := `configMetadata:
  patchStrategy:
    clientOptions.cli.discoverySources.oci.annotation: replace`
	cfgFiles, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2, cfgMetadata: metadata})

	defer func() {
		cleanUp()
	}()

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

	file, err := os.ReadFile(cfgFiles[0].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg, string(file))

	file, err = os.ReadFile(cfgFiles[1].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg2, string(file))
}
