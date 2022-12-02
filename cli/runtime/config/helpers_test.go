// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CfgTestData struct {
	cfg         string
	cfgNextGen  string
	cfgMetadata string
}

func setupTestConfig(t *testing.T, data *CfgTestData) (files []*os.File, cleanup func()) {
	// Setup config data
	cfgFile, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(cfgFile.Name(), []byte(data.cfg), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, cfgFile.Name())
	assert.NoError(t, err)

	cfgNextGenFile, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(cfgNextGenFile.Name(), []byte(data.cfgNextGen), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigNextGenKey, cfgNextGenFile.Name())
	assert.NoError(t, err)

	cfgMetadataFile, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(cfgMetadataFile.Name(), []byte(data.cfgMetadata), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigMetadataKey, cfgMetadataFile.Name())
	assert.NoError(t, err)

	cleanup = func() {
		err = os.Remove(cfgFile.Name())
		assert.NoError(t, err)

		err = os.Remove(cfgNextGenFile.Name())
		assert.NoError(t, err)

		err = os.Remove(cfgMetadataFile.Name())
		assert.NoError(t, err)
	}

	return []*os.File{cfgFile, cfgNextGenFile, cfgMetadataFile}, cleanup
}

func setupConfigMetadataWithMigrateToNewConfig() string {
	metadata := `configMetadata:
  settings:
    useUnifiedConfig: true`

	return metadata
}

func setupMultiCfgData() (string, string) {
	cfg := `servers:
  - name: test-mc
    type: managementcluster
    managementClusterOpts:
      endpoint: test-ctx-endpoint
      path: test-ctx-path
      context: test-ctx-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
  - name: test-mc2
    type: managementcluster
    managementClusterOpts:
      endpoint: test-ctx-endpoint
      path: test-ctx-path
      context: test-ctx-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
  - name: test-mc3
    type: managementcluster
    managementClusterOpts:
      endpoint: test-ctx-endpoint
      path: test-ctx-path
      context: test-ctx-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
current: test-mc
`
	cfg2 := `currentContext:
  kubernetes: test-mc
contexts:
  - name: test-mc
    ctx-field: new-ctx-field
    optional: true
    target: kubernetes
    clusterOpts:
      isManagementCluster: true
      endpoint: test-endpoint
      annotation: one
      required: true
      annotationStruct:
        one: one
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
  - name: test-mc2
    ctx-field: new-ctx-field
    optional: true
    target: kubernetes
    clusterOpts:
      isManagementCluster: true
      endpoint: test-endpoint
      annotation: one
      required: true
      annotationStruct:
        one: one
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
  - name: test-mc3
    ctx-field: new-ctx-field
    optional: true
    target: kubernetes
    clusterOpts:
      isManagementCluster: true
      endpoint: test-endpoint
      annotation: one
      required: true
      annotationStruct:
        one: one
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
`
	return cfg, cfg2
}
