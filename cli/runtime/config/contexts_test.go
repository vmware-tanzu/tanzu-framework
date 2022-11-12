// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestSetGetDeleteContext(t *testing.T) {
	// Setup config data
	f, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv("TANZU_CONFIG", f.Name())
	assert.NoError(t, err)

	//Setup metadata
	f2, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigMetadataKey, f2.Name())
	assert.NoError(t, err)

	// Cleanup
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f2.Name())

	ctx1 := &configapi.Context{
		Name: "test1",
		Type: "k8s",
		ClusterOpts: &configapi.ClusterServer{
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "updated-test-bucket",
					ManifestPath: "test-manifest-path",
				},
				ContextType: configapi.CtxTypeTMC,
			},
		},
	}

	ctx2 := &configapi.Context{
		Name: "test2",
		Type: configapi.CtxTypeK8s,
		ClusterOpts: &configapi.ClusterServer{
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "updated-test-bucket",
					ManifestPath: "test-manifest-path",
				},
				ContextType: configapi.CtxTypeTMC,
			},
		},
	}

	ctx, err := GetContext("test1")
	assert.Equal(t, "context test1 not found", err.Error())
	assert.Nil(t, ctx)

	err = SetContext(ctx1, true)
	assert.NoError(t, err)

	ctx, err = GetContext("test1")
	assert.Nil(t, err)
	assert.Equal(t, ctx1, ctx)

	ctx, err = GetCurrentContext(configapi.CtxTypeK8s)
	assert.Nil(t, err)
	assert.Equal(t, ctx1, ctx)

	err = SetContext(ctx2, false)
	assert.NoError(t, err)

	ctx, err = GetContext("test2")
	assert.Nil(t, err)
	assert.Equal(t, ctx2, ctx)

	ctx, err = GetCurrentContext(configapi.CtxTypeK8s)
	assert.Nil(t, err)
	assert.Equal(t, ctx1, ctx)

	err = DeleteContext("test")

	err = DeleteContext("test1")
	assert.Nil(t, err)

	ctx, err = GetContext("test1")
	assert.Nil(t, ctx)
	assert.Equal(t, "context test1 not found", err.Error())
}

func TestSetContextWithOldVersion(t *testing.T) {
	tanzuConfigBytes := `
currentContext:
    k8s: test-mc
contexts:
    - name: test-mc
      ctx-field: new-ctx-field
      optional: true
      type: k8s
      clusterOpts:
        isManagementCluster: true
        endpoint: old-test-endpoint
        annotation: one
        required: true
        annotationStruct:
            one: one
      discoverySources:
        - gcp:
            name: test
            bucket: test-ctx-bucket
            manifestPath: test-ctx-manifest-path
            annotation: one
            required: true
          contextType: tmc
        - gcp:
            name: test2
            bucket: test2-bucket
            manifestPath: test2-manifest-path
            annotation: one
            required: true
          contextType: tmc
`
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
	ctx := &configapi.Context{
		Name: "test-mc",
		Type: "k8s",
		ClusterOpts: &configapi.ClusterServer{
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "updated-test-bucket",
					ManifestPath: "test-manifest-path",
				},
				ContextType: configapi.CtxTypeTMC,
			},
		},
	}

	err = SetContext(ctx, false)
	assert.NoError(t, err)

	c, err := GetContext(ctx.Name)
	assert.NoError(t, err)
	assert.Equal(t, c.Name, ctx.Name)
	assert.Equal(t, c.ClusterOpts.Endpoint, "old-test-endpoint")
	assert.Equal(t, c.ClusterOpts.Path, ctx.ClusterOpts.Path)
	assert.Equal(t, c.ClusterOpts.Context, ctx.ClusterOpts.Context)
}

func TestSetContextWithDiscoverySourceWithNewFields(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name    string
		src     *configapi.ClientConfig
		ctx     *configapi.Context
		current bool
		errStr  string
	}{
		{
			name: "should add new context with new discovery sources to empty client config",
			src:  &configapi.ClientConfig{},
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
				DiscoverySources: []configapi.PluginDiscovery{
					{
						GCP: &configapi.GCPDiscovery{
							Name:         "test",
							Bucket:       "updated-test-bucket",
							ManifestPath: "test-manifest-path",
						},
						ContextType: configapi.CtxTypeTMC,
					},
				},
			},
			current: true,
		},
		{
			name: "should update existing context",
			src: &configapi.ClientConfig{
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: "k8s",
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
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
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
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
					},
				},
				CurrentServer: "test-mc",
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
				},
			},
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "updated-test-endpoint",
					Path:                "updated-test-path",
					Context:             "updated-test-context",
					IsManagementCluster: true,
				},
				DiscoverySources: []configapi.PluginDiscovery{
					{
						GCP: &configapi.GCPDiscovery{
							Name:         "test",
							Bucket:       "updated-test-bucket",
							ManifestPath: "updated-test-manifest-path",
						},
						ContextType: configapi.CtxTypeTMC,
					},
				},
			},
			current: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := SetContext(tc.ctx, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}

			ok, err := ContextExists(tc.ctx.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}

func TestSetContextWithDiscoverySource(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name    string
		src     *configapi.ClientConfig
		ctx     *configapi.Context
		current bool
		errStr  string
	}{
		{
			name: "should add new context with new discovery sources to empty client config",
			src:  &configapi.ClientConfig{},
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
				DiscoverySources: []configapi.PluginDiscovery{
					{
						GCP: &configapi.GCPDiscovery{
							Name:         "test",
							Bucket:       "updated-test-bucket",
							ManifestPath: "test-manifest-path",
						},
						ContextType: configapi.CtxTypeTMC,
					},
				},
			},
			current: true,
		},
		{
			name: "should update existing context",
			src: &configapi.ClientConfig{
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: "k8s",
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
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
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
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
					},
				},
				CurrentServer: "test-mc",
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
				},
			},
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "updated-test-endpoint",
					Path:                "updated-test-path",
					Context:             "updated-test-context",
					IsManagementCluster: true,
				},
				DiscoverySources: []configapi.PluginDiscovery{
					{
						GCP: &configapi.GCPDiscovery{
							Name:         "test",
							Bucket:       "updated-test-bucket",
							ManifestPath: "updated-test-manifest-path",
						},
						ContextType: configapi.CtxTypeTMC,
					},
				},
			},
			current: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := SetContext(tc.ctx, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}

			ok, err := ContextExists(tc.ctx.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}

func setupForGetContext(t *testing.T) {
	// setup
	cfg := &configapi.ClientConfig{
		KnownContexts: []*configapi.Context{
			{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
			{
				Name: "test-tmc",
				Type: "tmc",
				GlobalOpts: &configapi.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		CurrentContext: map[configapi.ContextType]string{
			"k8s": "test-mc",
			"tmc": "test-tmc",
		},
	}
	func() {
		LocalDirName = TestLocalDirName
		err := StoreClientConfig(cfg)
		assert.NoError(t, err)
	}()
}

func TestGetContext(t *testing.T) {
	setupForGetContext(t)

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tcs := []struct {
		name    string
		ctxName string
		errStr  string
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
		},
		{
			name:    "failure",
			ctxName: "test",
			errStr:  "context test not found",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := GetContext(tc.ctxName)
			if tc.errStr == "" {
				assert.Equal(t, tc.ctxName, c.Name)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
		})
	}
}

func TestContextExists(t *testing.T) {
	setupForGetContext(t)

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tcs := []struct {
		name    string
		ctxName string
		ok      bool
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
			ok:      true,
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
			ok:      true,
		},
		{
			name:    "failure",
			ctxName: "test",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := ContextExists(tc.ctxName)
			assert.Equal(t, tc.ok, ok)
			assert.NoError(t, err)
		})
	}
}

func TestSetContext(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
		// setup data
		node := &yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind:    yaml.MappingNode,
					Content: []*yaml.Node{},
				},
			},
		}
		err := persistConfig(node)
		assert.NoError(t, err)
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tcs := []struct {
		name    string
		ctx     *configapi.Context
		current bool
		errStr  string
	}{
		{
			name: "should add new context and set current on empty config",
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
			current: true,
		},

		{
			name: "should add new context but not current",
			ctx: &configapi.Context{
				Name: "test-mc2",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
		},
		{
			name: "success tmc current",
			ctx: &configapi.Context{
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
			ctx: &configapi.Context{
				Name: "test-tmc2",
				Type: "tmc",
				GlobalOpts: &configapi.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		{
			name: "success update test-mc",
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "good-test-endpoint",
					Path:                "updated-test-path",
					Context:             "updated-test-context",
					IsManagementCluster: true,
				},
			},
		},
		{
			name: "success update tmc",
			ctx: &configapi.Context{
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
			err := SetContext(tc.ctx, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			ctx, err := GetContext(tc.ctx.Name)
			assert.NoError(t, err)
			assert.Equal(t, tc.ctx.Name, ctx.Name)
			s, err := GetServer(tc.ctx.Name)
			assert.NoError(t, err)
			assert.Equal(t, tc.ctx.Name, s.Name)
		})
	}
}

func TestRemoveContext(t *testing.T) {
	// setup
	setupForGetContext(t)
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tcs := []struct {
		name    string
		ctxName string
		ctxType configapi.ContextType
		errStr  string
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
			ctxType: "k8s",
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
			ctxType: "tmc",
		},
		{
			name:    "failure",
			ctxName: "test",
			errStr:  "context test not found",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errStr == "" {
				ok, err := ContextExists(tc.ctxName)
				require.True(t, ok)
				require.NoError(t, err)
			}
			err := RemoveContext(tc.ctxName)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			ok, err := ContextExists(tc.ctxName)
			assert.False(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.ctxName)
			assert.Nil(t, err)
			assert.False(t, ok)
		})
	}
}

func TestSetCurrentContext(t *testing.T) {
	// setup
	setupForGetContext(t)
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tcs := []struct {
		name    string
		ctxType configapi.ContextType
		ctxName string
		errStr  string
	}{
		{
			name:    "success tmc",
			ctxName: "test-tmc",
			ctxType: "tmc",
		},
		{
			name:    "failure",
			ctxName: "test",
			errStr:  "context test not found",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SetCurrentContext(tc.ctxName)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			currCtx, err := GetCurrentContext(tc.ctxType)
			if tc.errStr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.ctxName, currCtx.Name)
			} else {
				assert.Error(t, err)
			}
			currSrv, err := GetCurrentServer()
			assert.NoError(t, err)
			if tc.errStr == "" {
				assert.Equal(t, tc.ctxName, currSrv.Name)
			}
		})
	}
}

func TestSetSingleContext(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tcs := []struct {
		name    string
		ctx     *configapi.Context
		current bool
		errStr  string
	}{
		{
			name: "success k8s current",
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SetContext(tc.ctx, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
			ok, err := ContextExists(tc.ctx.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.ctx.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}

func TestSetContextWithUniquePermissions(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	ctx := &configapi.Context{
		Name: "test-mc",
		Type: "tmc",
		GlobalOpts: &configapi.GlobalServer{
			Endpoint: "test-endpoint",
			Auth: configapi.GlobalServerAuth{
				IDToken: "",
				Issuer:  "https://console-stg.cloud.vmware.com/csp/gateway/am/api",
				Permissions: []string{
					"external/25834195-19aa-4ffd-8933-f5f20094ab24/service:member",
					"csp:org_owner",
					"external/f52d39b0-c298-4adf-9c6f-0a4a07351cd7/service:admin",
					"csp:org_member",
					"external/f52d39b0-c298-4adf-9c6f-0a4a07351cd7/service:member",
				},
				RefreshToken: "XXX",
				Type:         "api-token",
				UserName:     "tanzu-core",
			},
		},
	}

	ctx2 := &configapi.Context{
		Name: "test-mc",
		Type: "tmc",
		GlobalOpts: &configapi.GlobalServer{
			Endpoint: "test-endpoint-updated",
			Auth: configapi.GlobalServerAuth{
				IDToken: "",
				Issuer:  "https://console-stg.cloud.vmware.com/csp/gateway/am/api",
				Permissions: []string{
					"csp:org_member2",
					"external/f52d39b0-c298-4adf-9c6f-0a4a07351cd7/service:member",
				},
				RefreshToken: "XXX",
				Type:         "api-token",
				UserName:     "tanzu-core",
			},
		},
	}

	ctx3 := &configapi.Context{
		Name: "test-mc",
		Type: "tmc",
		GlobalOpts: &configapi.GlobalServer{
			Endpoint: "test-endpoint-updated3",
			Auth: configapi.GlobalServerAuth{
				IDToken: "",
				Issuer:  "https://console-stg.cloud.vmware.com/csp/gateway/am/api",
				Permissions: []string{
					"external/25834195-19aa-4ffd-8933-f5f20094ab24/service:member",
					"csp:org_owner3",
				},
				RefreshToken: "XXX",
				Type:         "api-token",
				UserName:     "tanzu-core",
			},
		},
	}

	for i := 1; i <= 100; i++ {
		err := SetContext(ctx, true)
		assert.NoError(t, err)
		err = SetContext(ctx2, true)
		assert.NoError(t, err)
		err = SetContext(ctx3, true)
		assert.NoError(t, err)
		err = SetContext(ctx, true)
		assert.NoError(t, err)
	}

	c, err := GetContext("test-mc")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(c.GlobalOpts.Auth.Permissions))

	s, err := GetServer("test-mc")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(s.GlobalOpts.Auth.Permissions))
}
