// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func TestPopulateContexts(t *testing.T) {
	tcs := []struct {
		name  string
		ip    *v1alpha1.ClientConfig
		op    *v1alpha1.ClientConfig
		delta bool
	}{
		{
			name:  "empty",
			ip:    &v1alpha1.ClientConfig{},
			op:    &v1alpha1.ClientConfig{},
			delta: false,
		},
		{
			name: "no delta",
			ip: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
				},
			},
			op: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
				},
			},
			delta: false,
		},
		{
			name: "w/ delta",
			ip: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeTMC: "test-tmc",
				},
			},
			op: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
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
		ip   *v1alpha1.ClientConfig
		op   *v1alpha1.ClientConfig
	}{
		{
			name: "empty",
			ip:   &v1alpha1.ClientConfig{},
			op:   &v1alpha1.ClientConfig{},
		},
		{
			name: "no delta",
			ip: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
				},
			},
			op: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
				},
			},
		},
		{
			name: "w/ delta",
			ip: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
				},
			},
			op: &v1alpha1.ClientConfig{
				KnownServers: []*v1alpha1.Server{
					{
						Name: "test-mc",
						Type: v1alpha1.ManagementClusterServerType,
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Path:     "test-path",
							Context:  "test-context",
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.GlobalServerType,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*v1alpha1.Context{
					{
						Name: "test-mc",
						Type: v1alpha1.CtxTypeK8s,
						ClusterOpts: &v1alpha1.ClusterServer{
							Endpoint:            "test-endpoint",
							Path:                "test-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
					},
					{
						Name: "test-tmc",
						Type: v1alpha1.CtxTypeTMC,
						GlobalOpts: &v1alpha1.GlobalServer{
							Endpoint: "test-endpoint",
						},
					},
				},
				CurrentContext: map[v1alpha1.ContextType]string{
					v1alpha1.CtxTypeK8s: "test-mc",
					v1alpha1.CtxTypeTMC: "test-tmc",
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
