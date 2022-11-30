// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupServersTestData() (string, string) {
	tanzuConfigBytes := `clientOptions:
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
	expectedConfig := `clientOptions:
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
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket
            manifestPath: test-manifest-path
            annotation: one
            required: true
          contextType: tmc
    - name: test-mc2
      type: k8s
      clusterOpts:
        endpoint: test-endpoint-updated
        path: test-path
        isManagementCluster: true
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path
          contextType: tmc
currentContext:
    k8s: test-mc2
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
      type: k8s
      managementClusterOpts:
        endpoint: test-endpoint-updated
        path: test-path
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket-updated
            manifestPath: test-manifest-path
          contextType: tmc
current: test-mc2
`
	return tanzuConfigBytes, expectedConfig
}
func TestServersIntegration(t *testing.T) {
	//Setup data and test config file
	tanzuConfigBytes, expectedConfig := setupServersTestData()
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
				ContextType: configapi.CtxTypeTMC,
			},
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, server)
	// Add new Server
	newServer := &configapi.Server{
		Name: "test-mc2",
		Type: "k8s",
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
				ContextType: configapi.CtxTypeTMC,
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
		Type: "k8s",
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
				ContextType: configapi.CtxTypeTMC,
			},
		},
	}
	err = SetServer(updatedServer, true)
	assert.NoError(t, err)
	s, err = GetServer("test-mc2")
	assert.Nil(t, err)
	assert.Equal(t, updatedServer, s)
	file, err := os.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte(expectedConfig), file)
	// Delete server
	err = DeleteServer("test-mc2")
	assert.NoError(t, err)
	s, err = GetServer("test-mc2")
	assert.Equal(t, err.Error(), "could not find server \"test-mc2\"")
	assert.Nil(t, s)
}
