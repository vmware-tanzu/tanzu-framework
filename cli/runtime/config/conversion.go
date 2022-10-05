// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/aunum/log"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

// PopulateContexts converts the known servers that are missing in contexts.
// This is needed when reading the config file persisted by an older core or plugin,
// so that it is forwards compatible with a new core plugin.
// Returns true if there was any delta.
func PopulateContexts(cfg *configapi.ClientConfig) bool {
	if cfg == nil || len(cfg.KnownServers) == 0 {
		return false
	}

	var delta bool
	if len(cfg.KnownContexts) == 0 {
		cfg.KnownContexts = make([]*configapi.Context, 0, len(cfg.KnownServers))
	}
	for _, s := range cfg.KnownServers {
		if cfg.HasContext(s.Name) {
			// server already present in known contexts; skip
			continue
		}

		delta = true
		// convert and append the server to the list of known contexts
		c := convertServerToContext(s)
		cfg.KnownContexts = append(cfg.KnownContexts, c)

		if s.Name == cfg.CurrentServer {
			err := cfg.SetCurrentContext(c.Type, c.Name)
			if err != nil {
				log.Warningf(err.Error())
			}
		}
	}

	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil {
		sources := cfg.ClientOptions.CLI.DiscoverySources
		for i := range cfg.ClientOptions.CLI.DiscoverySources {
			// This is a new field. So, using the K8s context since it is the only one available publicly.
			sources[i].ContextType = configapi.CtxTypeK8s
		}
	}

	return delta
}

func convertServerToContext(s *configapi.Server) *configapi.Context {
	if s == nil {
		return nil
	}

	return &configapi.Context{
		Name:             s.Name,
		Type:             convertServerTypeToContextType(s.Type),
		GlobalOpts:       s.GlobalOpts,
		ClusterOpts:      convertMgmtClusterOptsToClusterOpts(s.ManagementClusterOpts),
		DiscoverySources: s.DiscoverySources,
	}
}

func convertServerTypeToContextType(t configapi.ServerType) configapi.ContextType {
	switch t {
	case configapi.ManagementClusterServerType:
		return configapi.CtxTypeK8s
	case configapi.GlobalServerType:
		return configapi.CtxTypeTMC
	}
	// no other server type is supported in v0
	return configapi.ContextType(t)
}

func convertMgmtClusterOptsToClusterOpts(s *configapi.ManagementClusterServer) *configapi.ClusterServer {
	if s == nil {
		return nil
	}

	return &configapi.ClusterServer{
		Endpoint:            s.Endpoint,
		Path:                s.Path,
		Context:             s.Context,
		IsManagementCluster: true,
	}
}

// populateServers converts the known contexts that are missing in servers.
// This is needed when writing the config file from the newer core or plugin,
// so that it is backwards compatible with an older core or plugin.
func populateServers(cfg *configapi.ClientConfig) {
	if cfg == nil {
		return
	}

	if len(cfg.KnownServers) == 0 {
		cfg.KnownServers = make([]*configapi.Server, 0, len(cfg.KnownContexts))
	}
	for _, c := range cfg.KnownContexts {
		if cfg.HasServer(c.Name) {
			// context already present in known servers; skip
			continue
		}

		// convert and append the context to the list of known servers
		s := convertContextToServer(c)
		cfg.KnownServers = append(cfg.KnownServers, s)

		if cfg.CurrentServer == "" && (c.IsManagementCluster() || c.Type == configapi.CtxTypeTMC) && c.Name == cfg.CurrentContext[c.Type] {
			// This is lossy because only one server can be active at a time in the older CLI.
			// Using the K8s context for a management cluster or TMC, since these are the two
			// available publicly at the time of deprecation.
			cfg.CurrentServer = cfg.CurrentContext[configapi.CtxTypeK8s]
		}
	}
}

func convertContextToServer(c *configapi.Context) *configapi.Server {
	if c == nil {
		return nil
	}

	return &configapi.Server{
		Name:                  c.Name,
		Type:                  convertContextTypeToServerType(c.Type),
		GlobalOpts:            c.GlobalOpts,
		ManagementClusterOpts: convertClusterOptsToMgmtClusterOpts(c.ClusterOpts),
		DiscoverySources:      c.DiscoverySources,
	}
}

func convertContextTypeToServerType(t configapi.ContextType) configapi.ServerType {
	switch t {
	case configapi.CtxTypeK8s:
		// This is lossy because only management cluster servers are supported by the older CLI.
		return configapi.ManagementClusterServerType
	case configapi.CtxTypeTMC:
		return configapi.GlobalServerType
	}
	// no other context type is supported in v1 yet
	return configapi.ServerType(t)
}

func convertClusterOptsToMgmtClusterOpts(o *configapi.ClusterServer) *configapi.ManagementClusterServer {
	if o == nil || !o.IsManagementCluster {
		return nil
	}

	return &configapi.ManagementClusterServer{
		Endpoint: o.Endpoint,
		Path:     o.Path,
		Context:  o.Context,
	}
}
