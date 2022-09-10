// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// populateContexts converts the known servers that are missing in contexts.
// This is needed when reading the config file persisted by an older core or plugin,
// so that it is forwards compatible with a new core plugin.
// Returns true if there was any delta.
func populateContexts(cfg *configv1alpha1.ClientConfig) bool {
	if cfg == nil || len(cfg.KnownServers) == 0 {
		return false
	}

	var delta bool
	if len(cfg.KnownContexts) == 0 {
		cfg.KnownContexts = make([]*configv1alpha1.Context, 0, len(cfg.KnownServers))
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
			cfg.SetCurrentContext(c.Type, c.Name)
		}
	}

	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil {
		sources := cfg.ClientOptions.CLI.DiscoverySources
		for i := range cfg.ClientOptions.CLI.DiscoverySources {
			// This is a new field. So, using the K8s context since it is the only one available publicly.
			sources[i].ContextType = configv1alpha1.CtxTypeK8s
		}
	}

	return delta
}

func convertServerToContext(s *configv1alpha1.Server) *configv1alpha1.Context {
	if s == nil {
		return nil
	}

	return &configv1alpha1.Context{
		Name:             s.Name,
		Type:             convertServerTypeToContextType(s.Type),
		GlobalOpts:       s.GlobalOpts,
		ClusterOpts:      convertMgmtClusterOptsToClusterOpts(s.ManagementClusterOpts),
		DiscoverySources: s.DiscoverySources,
	}
}

func convertServerTypeToContextType(t configv1alpha1.ServerType) configv1alpha1.ContextType {
	switch t {
	case configv1alpha1.ManagementClusterServerType:
		return configv1alpha1.CtxTypeK8s
	case configv1alpha1.GlobalServerType:
		return configv1alpha1.CtxTypeTMC
	}
	// no other server type is supported in v0
	return configv1alpha1.ContextType(t)
}

func convertMgmtClusterOptsToClusterOpts(s *configv1alpha1.ManagementClusterServer) *configv1alpha1.ClusterServer {
	if s == nil {
		return nil
	}

	return &configv1alpha1.ClusterServer{
		Endpoint:            s.Endpoint,
		Path:                s.Path,
		Context:             s.Context,
		IsManagementCluster: true,
	}
}

// populateServers converts the known contexts that are missing in servers.
// This is needed when writing the config file from the newer core or plugin,
// so that it is backwards compatible with an older core or plugin.
func populateServers(cfg *configv1alpha1.ClientConfig) {
	if cfg == nil {
		return
	}

	if len(cfg.KnownServers) == 0 {
		cfg.KnownServers = make([]*configv1alpha1.Server, 0, len(cfg.KnownContexts))
	}
	for _, c := range cfg.KnownContexts {
		if cfg.HasServer(c.Name) {
			// context already present in known servers; skip
			continue
		}

		// convert and append the context to the list of known servers
		s := convertContextToServer(c)
		cfg.KnownServers = append(cfg.KnownServers, s)

		if cfg.CurrentServer == "" && (c.IsManagementCluster() || c.Type == configv1alpha1.CtxTypeTMC) && c.Name == cfg.CurrentContext[c.Type] {
			// This is lossy because only one server can be active at a time in the older CLI.
			// Using the K8s context for a management cluster or TMC, since these are the two
			// available publicly at the time of deprecation.
			cfg.CurrentServer = cfg.CurrentContext[configv1alpha1.CtxTypeK8s]
		}
	}
}

func convertContextToServer(c *configv1alpha1.Context) *configv1alpha1.Server {
	if c == nil {
		return nil
	}

	return &configv1alpha1.Server{
		Name:                  c.Name,
		Type:                  convertContextTypeToServerType(c.Type),
		GlobalOpts:            c.GlobalOpts,
		ManagementClusterOpts: convertClusterOptsToMgmtClusterOpts(c.ClusterOpts),
		DiscoverySources:      c.DiscoverySources,
	}
}

func convertContextTypeToServerType(t configv1alpha1.ContextType) configv1alpha1.ServerType {
	switch t {
	case configv1alpha1.CtxTypeK8s:
		// This is lossy because only management cluster servers are supported by the older CLI.
		return configv1alpha1.ManagementClusterServerType
	case configv1alpha1.CtxTypeTMC:
		return configv1alpha1.GlobalServerType
	}
	// no other context type is supported in v1 yet
	return configv1alpha1.ServerType(t)
}

func convertClusterOptsToMgmtClusterOpts(o *configv1alpha1.ClusterServer) *configv1alpha1.ManagementClusterServer {
	if o == nil || !o.IsManagementCluster {
		return nil
	}

	return &configv1alpha1.ManagementClusterServer{
		Endpoint: o.Endpoint,
		Path:     o.Path,
		Context:  o.Context,
	}
}
