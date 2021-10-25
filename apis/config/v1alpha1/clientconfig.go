// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

// IsConfigFeatureActivated return true if the feature is activated, false if not. An error if the featurePath is malformed
func (c *ClientConfig) IsConfigFeatureActivated(featurePath string) (bool, error) {
	plugin, flag, err := c.SplitFeaturePath(featurePath)
	if err != nil {
		return false, err
	}

	if c.ClientOptions == nil || c.ClientOptions.Features == nil ||
		c.ClientOptions.Features[plugin] == nil || c.ClientOptions.Features[plugin][flag] == "" {
		return false, nil
	}

	booleanValue, err := strconv.ParseBool(c.ClientOptions.Features[plugin][flag])
	if err != nil {
		errMsg := "error converting " + featurePath + " entry '" + c.ClientOptions.Features[plugin][flag] + "' to boolean value: " + err.Error()
		return false, errors.New(errMsg)
	}
	return booleanValue, nil
}

// SplitFeaturePath splits a features path into the pluginName and the featureName
// For example "features.management-cluster.dual-stack" returns "management-cluster", "dual-stack"
// An error results from a malformed path, including any path that does not start with "features."
func (c *ClientConfig) SplitFeaturePath(featurePath string) (string, string, error) {
	// parse the param
	paramArray := strings.Split(featurePath, ".")
	if len(paramArray) != 3 {
		return "", "", errors.New("unable to parse feature name config parameter into three parts [" + featurePath + "]  (was expecting features.<plugin>.<feature>)")
	}

	featuresLiteral := paramArray[0]
	plugin := paramArray[1]
	flag := paramArray[2]

	if featuresLiteral != "features" {
		return "", "", errors.New("unsupported feature config path parameter [" + featuresLiteral + "] (was expecting 'features.<plugin>.<feature>')")
	}
	return plugin, flag, nil
}
