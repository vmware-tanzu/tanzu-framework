// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupServersTestData() (string, string, string, string) {
	cfg := `clientOptions:
  cli:
    discoverySources:
      - oci:
          name: default
          image: "/:"
          unknown: cli-unknown
        contextType: k8s
      - local:
          name: default-local
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
`
	expectedCfg := `clientOptions:
    cli:
        discoverySources:
            - oci:
                name: default
                image: "/:"
                unknown: cli-unknown
              contextType: k8s
            - local:
                name: default-local
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
    - name: test-mc2
      type: managementcluster
      managementClusterOpts:
        endpoint: test-endpoint-updated
        path: test-path
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path
current: test-mc2
`

	cfg2 := `contexts:
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
currentContext:
    kubernetes: test-mc
`
	expectedCfg2 := `contexts:
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
    - name: test-mc2
      target: kubernetes
      clusterOpts:
        endpoint: test-endpoint-updated
        path: test-path
        isManagementCluster: true
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path
currentContext:
    kubernetes: test-mc2
`

	return cfg, expectedCfg, cfg2, expectedCfg2
}

func TestServersIntegration(t *testing.T) {
	// Setup config data
	cfg, expectedCfg, cfg2, expectedCfg2 := setupServersTestData()
	cfgTestFiles, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

	// Get Server
	server, err := GetServer("test-mc")
	expected := &configapi.Server{
		Name: "test-mc",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint",
			Path:     "test-path",
			Context:  "test-context",
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
	assert.Equal(t, expected, server)
	// Add new Server
	newServer := &configapi.Server{
		Name: "test-mc2",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint",
			Path:     "test-path",
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
	err = SetServer(newServer, true)
	assert.NoError(t, err)
	s, err := GetServer("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, newServer, s)
	// Update existing Server
	updatedServer := &configapi.Server{
		Name: "test-mc2",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint-updated",
			Path:     "test-path",
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket-updated",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}
	err = SetServer(updatedServer, true)
	assert.NoError(t, err)
	s, err = GetServer("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, updatedServer, s)

	file, err := os.ReadFile(cfgTestFiles[0].Name())
	assert.NoError(t, err)
	content := string(file)
	assert.Equal(t, expectedCfg, content)

	file, err = os.ReadFile(cfgTestFiles[1].Name())
	assert.NoError(t, err)
	content = string(file)
	assert.Equal(t, expectedCfg2, content)

	// Delete server
	err = DeleteServer("test-mc2")
	assert.NoError(t, err)
	s, err = GetServer("test-mc2")
	assert.Equal(t, err.Error(), "could not find server \"test-mc2\"")
	assert.Nil(t, s)
}

func TestServersIntegrationAndMigratedToNewConfig(t *testing.T) {
	// Setup config data
	cfg, _, cfg2, _ := setupServersTestData()
	expectedCfg2 := `contexts:
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
    - name: test-mc2
      target: kubernetes
      clusterOpts:
        endpoint: test-endpoint-updated
        path: test-path
        isManagementCluster: true
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path
currentContext:
    kubernetes: test-mc2
servers:
    - name: test-mc2
      type: managementcluster
      managementClusterOpts:
        endpoint: test-endpoint-updated
        path: test-path
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path
current: test-mc2
`
	// Setup config data
	cfgTestFiles, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2, cfgMetadata: setupConfigMetadataWithMigrateToNewConfig()})

	defer func() {
		cleanUp()
	}()

	_, err := GetServer("test-mc")
	assert.Equal(t, "could not find server \"test-mc\"", err.Error())

	// Add new Server
	newServer := &configapi.Server{
		Name: "test-mc2",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint",
			Path:     "test-path",
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
	err = SetServer(newServer, true)
	assert.NoError(t, err)

	s, err := GetServer("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, newServer, s)

	// Update existing Server
	updatedServer := &configapi.Server{
		Name: "test-mc2",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint-updated",
			Path:     "test-path",
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket-updated",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}
	err = SetServer(updatedServer, true)
	assert.NoError(t, err)
	s, err = GetServer("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, updatedServer, s)

	file, err := os.ReadFile(cfgTestFiles[0].Name())
	assert.NoError(t, err)
	assert.Equal(t, cfg, string(file))

	file, err = os.ReadFile(cfgTestFiles[1].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg2, string(file))

	// Delete server
	err = DeleteServer("test-mc2")
	assert.NoError(t, err)
	s, err = GetServer("test-mc2")
	assert.Equal(t, err.Error(), "could not find server \"test-mc2\"")
	assert.Nil(t, s)
}
