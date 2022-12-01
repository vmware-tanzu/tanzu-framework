// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetCLIDiscoverySources(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	tests := []struct {
		name   string
		in     *configapi.ClientConfig
		out    []configapi.PluginDiscovery
		errStr string
	}{
		{
			name: "success get all",
			in: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "updated-test-bucket",
									ManifestPath: "test-manifest-path",
								},
							},
						},
					},
				},
			},
			out: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "updated-test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := StoreClientConfig(spec.in)
			assert.NoError(t, err)
			c, err := GetCLIDiscoverySources()
			assert.Equal(t, spec.out, c)
			assert.NoError(t, err)
		})
	}
}

func TestGetCLIDiscoverySource(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	tests := []struct {
		name string
		in   *configapi.ClientConfig
		out  *configapi.PluginDiscovery
	}{
		{
			name: "success get",
			in: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "updated-test-bucket",
									ManifestPath: "test-manifest-path",
								},
							},
						},
					},
				},
			},
			out: &configapi.PluginDiscovery{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "updated-test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := StoreClientConfig(spec.in)
			assert.NoError(t, err)
			c, err := GetCLIDiscoverySource("test")
			assert.Equal(t, spec.out, c)
			assert.NoError(t, err)
		})
	}
}

func TestSetCLIDiscoverySources(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	tests := []struct {
		name  string
		input []configapi.PluginDiscovery
		total int
	}{
		{
			name: "success add test",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "default",
						Path: "standalone",
					},
					ContextType: configapi.CtxTypeK8s,
				},
			},
			total: 2,
		},
		{
			name: "success add test",
			input: []configapi.PluginDiscovery{
				{
					Local: &configapi.LocalDiscovery{
						Name: "admin-local",
						Path: "admin",
					},
				},
			},
			total: 3,
		},
		{
			name: "success add test",
			input: []configapi.PluginDiscovery{
				{
					OCI: &configapi.OCIDiscovery{
						Name:  "default",
						Image: "test-image",
					},
				},
			},
			total: 3,
		},
		{
			name: "success update test",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "test-updated-bucket",
						ManifestPath: "test-updated-manifest-path",
					},
				},
			},
			total: 3,
		},
		{
			name: "should not persist same test",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "test-updated-bucket",
						ManifestPath: "test-updated-manifest-path",
					},
				},
			},
			total: 3,
		},
		{
			name: "success add default gcp",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},
			total: 3,
		},
		{
			name: "success add default-local gcp",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default-local",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},
			total: 4,
		},
		{
			name: "success add default-local local",
			input: []configapi.PluginDiscovery{
				{
					Local: &configapi.LocalDiscovery{
						Name: "default-local",
						Path: "test-path",
					},
				},
			},
			total: 4,
		},
		{
			name: "success add default-local local",
			input: []configapi.PluginDiscovery{
				{
					Local: &configapi.LocalDiscovery{
						Name: "default-local",
						Path: "test-path",
					},
					ContextType: configapi.CtxTypeTMC,
				},
			},
			total: 4,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := SetCLIDiscoverySources(spec.input)
			assert.NoError(t, err)
			sources, err := GetCLIDiscoverySources()
			assert.NoError(t, err)
			assert.Equal(t, spec.total, len(sources))
		})
	}
}

func TestDeleteCLIDiscoverySource(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	tests := []struct {
		name    string
		src     *configapi.ClientConfig
		input   string
		deleted bool
		errStr  string
	}{
		{
			name: "should return true on deleting non existing item",
			src: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "test-bucket",
									ManifestPath: "test-manifest-path",
								},
							},
						},
					},
				},
			},
			input:   "test-notfound",
			deleted: true,
			errStr:  "cli discovery source not found",
		},
		{
			name: "should return true on deleting existing item",
			src: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "test-bucket",
									ManifestPath: "test-manifest-path",
								},
							},
						},
					},
				},
			},
			input:   "test",
			deleted: true,
		},
		{
			name: "should return true on deleting existing item2",
			src: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						DiscoverySources: []configapi.PluginDiscovery{
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
						},
					},
				},
			},
			input:   "test",
			deleted: true,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := StoreClientConfig(spec.src)
			assert.NoError(t, err)
			err = DeleteCLIDiscoverySource(spec.input)
			if spec.errStr != "" {
				assert.Equal(t, err.Error(), spec.errStr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntegrationSetGetDeleteCLIDiscoverySource(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	sources := []configapi.PluginDiscovery{
		{
			GCP: &configapi.GCPDiscovery{
				Name:         "default",
				Bucket:       "test-bucket",
				ManifestPath: "test-manifest-path",
			},
		},
	}

	// Get from the empty config
	ds, err := GetCLIDiscoverySource("test")
	assert.Equal(t, "cli discovery source not found", err.Error())
	assert.Nil(t, ds)

	// Add source to empty config
	err = SetCLIDiscoverySources(sources)
	assert.NoError(t, err)

	ds, err = GetCLIDiscoverySource("default")
	assert.Nil(t, err)
	assert.Equal(t, sources[0], *ds)

	// Delete existing source
	err = DeleteCLIDiscoverySource("default")
	assert.NoError(t, err)

	ds, err = GetCLIDiscoverySource("default")
	assert.Equal(t, "cli discovery source not found", err.Error())
	assert.Nil(t, ds)

	err = DeleteCLIDiscoverySource("default-local")
	assert.Equal(t, "cli discovery source not found", err.Error())

	ds, err = GetCLIDiscoverySource("default-local")
	assert.Equal(t, "cli discovery source not found", err.Error())
	assert.Nil(t, ds)

	ds, err = GetCLIDiscoverySource("default")
	assert.Equal(t, "cli discovery source not found", err.Error())
	assert.Nil(t, ds)
}

func TestSetCLIDiscoverySourceLocalMulti(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	src := &configapi.ClientConfig{
		ClientOptions: &configapi.ClientOptions{
			CLI: &configapi.CLIOptions{},
		},
	}
	input := configapi.PluginDiscovery{
		Local: &configapi.LocalDiscovery{
			Name: "admin-local",
			Path: "admin",
		},
	}
	input2 := configapi.PluginDiscovery{
		Local: &configapi.LocalDiscovery{
			Name: "default-local",
			Path: "standalone",
		},
	}
	updateInput2 := configapi.PluginDiscovery{
		Local: &configapi.LocalDiscovery{
			Name: "default-local",
			Path: "standalone-updated",
		},
	}

	// Actions
	err := StoreClientConfig(src)
	assert.NoError(t, err)
	err = SetCLIDiscoverySource(input)
	assert.NoError(t, err)
	c, err := GetCLIDiscoverySource("admin-local")
	assert.Equal(t, input.Local, c.Local)
	assert.NoError(t, err)
	err = SetCLIDiscoverySource(input2)
	assert.NoError(t, err)
	c, err = GetCLIDiscoverySource("default-local")
	assert.Equal(t, input2.Local, c.Local)
	assert.NoError(t, err)
	// Update Input2
	err = SetCLIDiscoverySource(updateInput2)
	assert.NoError(t, err)
	c, err = GetCLIDiscoverySource("default-local")
	assert.Equal(t, updateInput2.Local, c.Local)
	assert.NoError(t, err)
}

func TestSetCLIDiscoverySourceWithDefaultAndDefaultLocal(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	tests := []struct {
		name         string
		input        []configapi.PluginDiscovery
		totalSources int
	}{
		{
			name: "success add default-test source",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default-test",
						Bucket:       "default-test-bucket",
						ManifestPath: "default-test-manifest-path",
					},
				},
			},
			totalSources: 1,
		},
		{
			name: "success add default source",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "default-test-bucket",
						ManifestPath: "default-test-manifest-path",
					},
				},
			},
			totalSources: 2,
		},
		{
			name: "success add default-local source",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default-local",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},

			totalSources: 3,
		},
		{
			name: "success update default",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "default-test-bucket-updated",
						ManifestPath: "default-test-manifest-path-updated",
					},
				},
			},

			totalSources: 3,
		},
		{
			name: "success update default-local",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default-local",
						Bucket:       "default-test-bucket-updated",
						ManifestPath: "default-test-manifest-path-updated",
					},
				},
			},

			totalSources: 3,
		},
		{
			name: "success add default",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "default-test-bucket-updated",
						ManifestPath: "default-test-manifest-path-updated",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "test-local",
						Path: "test-local-path",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "default",
						Path: "default-local-path",
					},
				},
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "default-test-bucket-updated2",
						ManifestPath: "default-test-manifest-path-updated2",
					},
				},
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test-gcp1",
						Bucket:       "test-bucket-updated",
						ManifestPath: "test-manifest-path-updated",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "default-local",
						Path: "default-local-path",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "test-gcp1",
						Path: "default-local-path",
					},
				},
			},
			totalSources: 5,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			for _, ds := range spec.input {
				err := SetCLIDiscoverySource(ds)
				assert.NoError(t, err)
			}

			if spec.totalSources != 0 {
				sources, err := GetCLIDiscoverySources()
				assert.NoError(t, err)
				assert.Equal(t, spec.totalSources, len(sources))
			}
		})
	}
}

func TestSetCLIDiscoverySourceMultiTypes(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	tests := []struct {
		name         string
		input        []configapi.PluginDiscovery
		totalSources int
	}{

		{
			name: "success add multiple discovery source types",
			input: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "default-test-bucket-updated",
						ManifestPath: "default-test-manifest-path-updated",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "test-local",
						Path: "test-local-path",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "default",
						Path: "default-local-path",
					},
				},
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "default",
						Bucket:       "default-test-bucket-updated2",
						ManifestPath: "default-test-manifest-path-updated2",
					},
				},
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test-gcp1",
						Bucket:       "test-bucket-updated",
						ManifestPath: "test-manifest-path-updated",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "default-local",
						Path: "default-local-path",
					},
				},
				{
					Local: &configapi.LocalDiscovery{
						Name: "test-gcp1",
						Path: "default-local-path",
					},
				},
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test-gcp2",
						Bucket:       "test-bucket-updated",
						ManifestPath: "test-manifest-path-updated",
					},
				},
			},
			totalSources: 5,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			for _, ds := range spec.input {
				err := SetCLIDiscoverySource(ds)
				assert.NoError(t, err)
			}

			if spec.totalSources != 0 {
				sources, err := GetCLIDiscoverySources()
				assert.NoError(t, err)
				assert.Equal(t, spec.totalSources, len(sources))
			}
		})
	}
}
