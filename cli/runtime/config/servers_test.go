// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestSetGetDeleteServer(t *testing.T) {
	// Setup config data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfg2})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

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
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfgMetadata: setupConfigMetadataWithMigrateToNewConfig()})

	defer func() {
		cleanUp()
	}()

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
