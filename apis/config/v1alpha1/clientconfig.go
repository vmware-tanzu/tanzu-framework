// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fmt"

const (
	// AllUnstableVersions allows all plugin versions
	AllUnstableVersions VersionSelectorLevel = "all"
	// AlphaUnstableVersions allows all alpha tagged versions
	AlphaUnstableVersions VersionSelectorLevel = "alpha"
	// ExperimentalUnstableVersions includes all pre-releases, minus +build tags
	ExperimentalUnstableVersions VersionSelectorLevel = "experimental"
	// NoUnstableVersions allows no unstable plugin versions, format major.minor.patch only
	NoUnstableVersions VersionSelectorLevel = "none"
)

const (
	// FeatureCli allows a feature to be set at the CLI level (globally) rather than for a single plugin
	FeatureCli string = "cli"
)

// VersionSelectorLevel allows selecting plugin versions based on semver properties
type VersionSelectorLevel string

// IsGlobal tells if the server is global.
func (s *Server) IsGlobal() bool {
	return s.Type == GlobalServerType
}

// IsManagementCluster tells if the server is a management cluster.
func (s *Server) IsManagementCluster() bool {
	return s.Type == ManagementClusterServerType
}

// GetCurrentServer returns the current server/
func (c *ClientConfig) GetCurrentServer() (*Server, error) {
	for _, server := range c.KnownServers {
		if server.Name == c.CurrentServer {
			return server, nil
		}
	}
	return nil, fmt.Errorf("current server %q not found", c.CurrentServer)
}

// SetUnstableVersionSelector will help determine the unstable versions supported
// In order of restrictiveness:
// "all" -> "alpha" -> "experimental" -> "none"
// none: return stable versions only. the default for both the config and the old flag.
// alpha: only versions tagged with -alpha
// experimental: all pre-release versions without +build semver data
// all: return all unstable versions.
func (c *ClientConfig) SetUnstableVersionSelector(f VersionSelectorLevel) {
	if c.ClientOptions == nil {
		c.ClientOptions = &ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &CLIOptions{}
	}
	switch f {
	case AllUnstableVersions, AlphaUnstableVersions, ExperimentalUnstableVersions, NoUnstableVersions:
		c.ClientOptions.CLI.UnstableVersionSelector = f
		return
	}
	c.ClientOptions.CLI.UnstableVersionSelector = AllUnstableVersions
}
