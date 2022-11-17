// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestPopulateContexts(t *testing.T) {
	tcs := []struct {
		name  string
		ip    *configapi.ClientConfig
		op    *configapi.ClientConfig
		delta bool
	}{
		{
			name:  "empty",
			ip:    &configapi.ClientConfig{},
			op:    &configapi.ClientConfig{},
			delta: false,
		},
		{
			name: "no delta",
			ip: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
			op: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
			delta: false,
		},
		{
			name: "w/ delta",
			ip: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
			op: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
			delta: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			delta := PopulateContexts(tc.ip)
			assert.Equal(t, tc.delta, delta)
			// ensure that the servers are not lost
			assert.Equal(t, len(tc.op.KnownServers), len(tc.ip.KnownServers))
			assert.Equal(t, tc.op.CurrentServer, tc.ip.CurrentServer)
			// ensure that the missing contexts are added
			assert.Equal(t, len(tc.op.KnownContexts), len(tc.ip.KnownContexts))
			assert.Equal(t, tc.op.CurrentContext, tc.ip.CurrentContext)
		})
	}
}

func TestPopulateServers(t *testing.T) {
	tcs := []struct {
		name string
		ip   *configapi.ClientConfig
		op   *configapi.ClientConfig
	}{
		{
			name: "empty",
			ip:   &configapi.ClientConfig{},
			op:   &configapi.ClientConfig{},
		},
		{
			name: "no delta",
			ip: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
			op: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
		},
		{
			name: "w/ delta",
			ip: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
			op: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.GlobalServerType,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: configapi.CtxTypeK8s,
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: configapi.CtxTypeTMC,
						GlobalOpts: &configapi.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
					configapi.CtxTypeTMC: "test-tmc",
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			populateServers(tc.ip)
			// ensure that the contexts are not lost
			assert.Equal(t, len(tc.op.KnownContexts), len(tc.ip.KnownContexts))
			assert.Equal(t, tc.op.CurrentContext, tc.ip.CurrentContext)
			// ensure that the missing servers are added
			assert.Equal(t, len(tc.op.KnownServers), len(tc.ip.KnownServers))
			assert.Equal(t, tc.op.CurrentServer, tc.ip.CurrentServer)
		})
	}
}
