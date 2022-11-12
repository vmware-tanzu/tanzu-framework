// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestSetGetDeleteServer(t *testing.T) {
	// Setup config data
	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, f1.Name())
	assert.NoError(t, err)

	// Setup config-ng data
	f2, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(""), 0644)
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

	server1 := &configapi.Server{
		Name: "test1",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint",
			Path:     "test-path",
			Context:  "test-server",
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

	server2 := &configapi.Server{
		Name: "test2",
		Type: "managementcluster",
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Endpoint: "test-endpoint",
			Path:     "test-path",
			Context:  "test-server",
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

	s, err := GetServer("test1")
	assert.Equal(t, "could not find server \"test1\"", err.Error())
	assert.Nil(t, s)

	err = SetServer(server1, true)
	assert.NoError(t, err)

	s, err = GetServer("test1")
	assert.Nil(t, err)
	assert.Equal(t, server1, s)

	s, err = GetCurrentServer()
	assert.Nil(t, err)
	assert.Equal(t, server1, s)

	err = SetServer(server2, false)
	assert.NoError(t, err)

	s, err = GetServer("test2")
	assert.Nil(t, err)
	assert.Equal(t, server2, s)

	server, err := GetCurrentServer()
	assert.Nil(t, err)
	assert.Equal(t, server1, server)

	err = DeleteServer("test1")
	assert.Nil(t, err)

	err = DeleteServer("test2")
	assert.Nil(t, err)

	server, err = GetServer("test2")
	assert.Nil(t, server)
	assert.Equal(t, "could not find server \"test2\"", err.Error())
}

func TestGetServer(t *testing.T) {
	// Setup config data
	cfg, cfg2 := setupMultiCfgData()
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

	tcs := []struct {
		name       string
		serverName string
		errStr     string
	}{
		{
			name:       "success test-mc",
			serverName: "test-mc",
		},
		{
			name:       "success test-mc2",
			serverName: "test-mc2",
		},
		{
			name:       "failure",
			serverName: "test",
			errStr:     "could not find server \"test\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := GetServer(tc.serverName)
			if tc.errStr == "" {
				assert.Equal(t, tc.serverName, c.Name)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
		})
	}
}

func TestServerExists(t *testing.T) {
	// Setup config data
	cfg, cfg2 := setupMultiCfgData()
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

	tcs := []struct {
		name       string
		serverName string
		ok         bool
	}{
		{
			name:       "success test-mc",
			serverName: "test-mc",
			ok:         true,
		},
		{
			name:       "success test-mc2",
			serverName: "test-mc2",
			ok:         true,
		},
		{
			name:       "failure",
			serverName: "test",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := ServerExists(tc.serverName)
			assert.Equal(t, tc.ok, ok)
			assert.NoError(t, err)
		})
	}
}

func TestSetServer(t *testing.T) {
	// Setup config data
	cfg, cfg2 := setupMultiCfgData()
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

	tcs := []struct {
		name    string
		server  *configapi.Server
		current bool
		errStr  string
	}{
		{
			name: "should add new server and set current on empty config",
			server: &configapi.Server{
				Name: "test",
				Type: "managementcluster",
				ManagementClusterOpts: &configapi.ManagementClusterServer{
					Endpoint: "test-endpoint",
					Path:     "test-path",
					Context:  "test-server",
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
			current: true,
		},
		{
			name: "should add new server but not current",
			server: &configapi.Server{
				Name: "test2",
				Type: "managementcluster",
				ManagementClusterOpts: &configapi.ManagementClusterServer{
					Endpoint: "test-endpoint",
					Path:     "test-path",
					Context:  "test-server",
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
		{
			name: "success tmc current",
			server: &configapi.Server{
				Name: "test-tmc1",
				Type: "tmc",
				GlobalOpts: &configapi.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
			current: true,
		},
		{
			name: "success tmc not_current",
			server: &configapi.Server{
				Name: "test-tmc2",
				Type: "tmc",
				GlobalOpts: &configapi.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		{
			name: "success update test-mc",
			server: &configapi.Server{
				Name: "test",
				Type: "managementcluster",
				ManagementClusterOpts: &configapi.ManagementClusterServer{
					Endpoint: "test-endpoint",
					Path:     "test-path",
					Context:  "test-server",
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
		{
			name: "success update tmc",
			server: &configapi.Server{
				Name: "test-tmc",
				Type: "tmc",
				GlobalOpts: &configapi.GlobalServer{
					Endpoint: "updated-test-endpoint",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// perform test
			err := SetServer(tc.server, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			server, err := GetServer(tc.server.Name)
			assert.NoError(t, err)
			assert.Equal(t, tc.server.Name, server.Name)
			s, err := GetServer(tc.server.Name)
			assert.NoError(t, err)
			assert.Equal(t, tc.server.Name, s.Name)
		})
	}
}

func TestRemoveServer(t *testing.T) {
	// Setup config data
	cfg, cfg2 := setupMultiCfgData()
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

	tcs := []struct {
		name       string
		serverName string
		errStr     string
	}{
		{
			name:       "success test-mc",
			serverName: "test-mc",
		},
		{
			name:       "success test-mc2",
			serverName: "test-mc2",
		},
		{
			name:       "failure",
			serverName: "test",
			errStr:     "could not find server \"test\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errStr == "" {
				ok, err := ServerExists(tc.serverName)
				require.True(t, ok)
				require.NoError(t, err)
			}
			err := RemoveServer(tc.serverName)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			ok, err := ServerExists(tc.serverName)
			assert.False(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.serverName)
			assert.Nil(t, err)
			assert.False(t, ok)
		})
	}
}

func TestSetCurrentServer(t *testing.T) {
	// Setup config data
	cfg, cfg2 := setupMultiCfgData()
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

	tcs := []struct {
		name       string
		serverName string
		errStr     string
	}{
		{
			name:       "success mc",
			serverName: "test-mc",
		},
		{
			name:       "failure",
			serverName: "test",
			errStr:     "could not find server \"test\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SetCurrentServer(tc.serverName)
			if tc.errStr != "" {
				assert.Equal(t, "could not find server \"test\"", err.Error())
			} else {
				assert.NoError(t, err)
				currCtx, err := GetCurrentServer()
				if tc.errStr != "" {
					assert.Equal(t, "", err.Error())
				} else {
					assert.Equal(t, tc.serverName, currCtx.Name)
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestSetSingleServer(t *testing.T) {
	// Setup config data
	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, f1.Name())
	assert.NoError(t, err)

	f2, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(""), 0644)
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

	tcs := []struct {
		name    string
		server  *configapi.Server
		current bool
		errStr  string
	}{
		{
			name: "success k8s current",
			server: &configapi.Server{
				Name: "test",
				Type: "managementcluster",
				ManagementClusterOpts: &configapi.ManagementClusterServer{
					Endpoint: "test-endpoint",
					Path:     "test-path",
					Context:  "test-server",
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
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SetServer(tc.server, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			ok, err := ServerExists(tc.server.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.server.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}

func TestSetSingleServerWithMigrateToNewConfig(t *testing.T) {
	// Setup config data
	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, f1.Name())
	assert.NoError(t, err)

	f2, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigNextGenKey, f2.Name())
	assert.NoError(t, err)

	//Setup metadata
	fMeta, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(fMeta.Name(), []byte(setupConfigMetadataWithMigrateToNewConfig()), 0644)
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

	tcs := []struct {
		name    string
		server  *configapi.Server
		current bool
		errStr  string
	}{
		{
			name: "success k8s current",
			server: &configapi.Server{
				Name: "test",
				Type: "managementcluster",
				ManagementClusterOpts: &configapi.ManagementClusterServer{
					Endpoint: "test-endpoint",
					Path:     "test-path",
					Context:  "test-server",
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
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SetServer(tc.server, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			ok, err := ServerExists(tc.server.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.server.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}
