// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package plugin ...
package plugin

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/distribution"
)

// Discovered defines discovered plugin resource
type Discovered struct {
	// Description is the plugin's description.
	Name string

	// Description is the plugin's description.
	Description string

	// RecommendedVersion is the version that Tanzu CLI should use if available.
	// The value should be a valid semantic version as defined in
	// https://semver.org/. E.g., 2.0.1
	RecommendedVersion string

	// InstalledVersion is the version that Tanzu CLI should use if available.
	// The value should be a valid semantic version as defined in
	// https://semver.org/. E.g., 2.0.1
	InstalledVersion string

	// SupportedVersions determines the list of supported CLI plugin versions.
	// The values are sorted in the semver prescribed order as defined in
	// https://github.com/Masterminds/semver#sorting-semantic-versions.
	SupportedVersions []string

	// Distribution is an interface to download a single plugin binary.
	Distribution distribution.Distribution

	// Optional specifies whether the plugin is mandatory or optional
	// If optional, the plugin will not get auto-downloaded as part of
	// `tanzu login` or `tanzu plugin sync` command
	// To view the list of plugin, user can use `tanzu plugin list` and
	// to download a specific plugin run, `tanzu plugin install <plugin-name>`
	Optional bool

	// Scope is the context association level of the plugin.
	Scope string

	// Source is the name of the discovery source from where the plugin was
	// discovered.
	Source string

	// DiscoveryType defines the type of the discovery. Possible values are
	// oci, local or kubernetes
	DiscoveryType string

	// Status is the installed/uninstalled status of the plugin.
	Status string
}
