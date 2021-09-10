// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

// Distribution is an interface to download a single plugin binary.
type Distribution interface {
	// Fetch the binary for a plugin version.
	Fetch(version, os, arch string) ([]byte, error)
}

// VersionInfo provides the set of supported and recommended versions of a
// plugin.
type VersionInfo interface {
	// SupportedVersions determines the list of supported CLI plugin versions.
	// The values are sorted in the semver prescribed order as defined in
	// https://github.com/Masterminds/semver#sorting-semantic-versions.
	GetSupportedVersions() []string
	// RecommendedVersion version that Tanzu CLI should use if available.
	// The value should be a valid semantic version as defined in
	// https://semver.org/.
	GetRecommendedVersion() string
}

// Plugin is an interface that provides necessary metadata about a CLI plugin.
type Plugin interface {
	// Name is the name of the plugin.
	GetName() string
	// Description is the plugin's description.
	GetDescription() string
	// Required denotes if this plugin is needed for all or at least most use cases.
	IsRequired() bool
	// Distribution mechanism for the plugin.
	Distribution
	// VersionInfo for the plugin describes constraints
	// around using a version of the plugin.
	VersionInfo

	// GetDiscovery specificies the name of the discovery from where
	// this plugin is discovered.
	GetDiscovery() string
	// Scope specificies the scope of the plugin. Stand-Alone or Context
	GetScope() string
	// Status specificies the current plugin installation status
	GetStatus() string

	// SetDiscovery specificies the name of the discovery from where
	// this plugin is discovered.
	SetDiscovery(discovery string)
	// SetScope specificies the scope of the plugin. Stand-Alone or Context
	SetScope(scope string)
	// SetStatus specificies the current plugin installation status
	SetStatus(status string)
}
