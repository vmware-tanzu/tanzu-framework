// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupStoreClientConfigData() (string, string, *configapi.ClientConfig) {
	tanzuConfig := `clientOptions:
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
            - gcp:
                name: test
                bucket: ctx-test-bucket
                manifestPath: ctx-test-manifest-path
              contextType: k8s
        repositories:
            - gcpPluginRepository:
                name: test
                bucketName: bucket
                rootPath: root-path
        unstableVersionSelector: unstable-version
        edition: test=tkg
        bomRepo: test-bomrepo
        compatibilityFilePath: test-compatibility-file-path
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
      type: k8s
      group: one
      clusterOpts:
        isManagementCluster: true
        annotation: one
        required: true
        annotationStruct:
            one: one
        endpoint: test-context-endpoint
        path: test-context-path
        context: test-context
      discoverySources:
        - gcp:
            name: test
            bucket: ctx-test-bucket
            manifestPath: ctx-test-manifest-path
            annotation: one
            required: true
          contextType: tmc
currentContext:
    k8s: test-mc
`
	cfg := &configapi.ClientConfig{
		KnownServers: []*configapi.Server{
			{
				Name: "test-mc",
				Type: configapi.ManagementClusterServerType,
				ManagementClusterOpts: &configapi.ManagementClusterServer{
					Endpoint: "test-endpoint",
					Context:  "test-context",
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
			},
		},
		CurrentServer: "test-mc",
		KnownContexts: []*configapi.Context{
			{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-context-endpoint",
					Path:                "test-context-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
				DiscoverySources: []configapi.PluginDiscovery{
					{
						GCP: &configapi.GCPDiscovery{
							Name:         "test",
							Bucket:       "ctx-test-bucket",
							ManifestPath: "ctx-test-manifest-path",
						},
						ContextType: configapi.CtxTypeTMC,
					},
				},
			},
		},
		CurrentContext: map[configapi.ContextType]string{
			configapi.CtxTypeK8s: "test-mc",
		},
		ClientOptions: &configapi.ClientOptions{
			CLI: &configapi.CLIOptions{
				Repositories: []configapi.PluginRepository{
					{
						GCPPluginRepository: &configapi.GCPPluginRepository{
							Name:       "test",
							BucketName: "bucket",
							RootPath:   "root-path",
						},
					},
				},
				DiscoverySources: []configapi.PluginDiscovery{
					{
						GCP: &configapi.GCPDiscovery{
							Name:         "test",
							Bucket:       "ctx-test-bucket",
							ManifestPath: "ctx-test-manifest-path",
						},
						ContextType: configapi.CtxTypeTMC,
					},
				},
				UnstableVersionSelector: configapi.VersionSelectorLevel("unstable-version"),
				Edition:                 configapi.EditionSelector("test=tkg"),
				BOMRepo:                 "test-bomrepo",
				CompatibilityFilePath:   "test-compatibility-file-path",
			},
		},
	}
	return tanzuConfig, expectedConfig, cfg
}

func TestStoreClientConfig(t *testing.T) {
	tanzuConfig, expectedConfig, cfg := setupStoreClientConfigData()
	f, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f.Name(), []byte(tanzuConfig), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f.Name())
	err = os.Setenv("TANZU_CONFIG", f.Name())
	assert.NoError(t, err)
	err = StoreClientConfig(cfg)
	assert.NoError(t, err)
	file, err := os.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte(expectedConfig), file)
}
