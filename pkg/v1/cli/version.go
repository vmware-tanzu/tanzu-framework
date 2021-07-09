// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"strings"

	"golang.org/x/mod/semver"
)

// VersionLatest is the latest version.
var VersionLatest = "latest"

// VersionSelector returns a version from a set of versions.
type VersionSelector func(versions []string) string

// DefaultVersionSelector is the default version selector.
var DefaultVersionSelector = SelectVersionStable

// FilterVersions returns the list of valid versions depending on whether
// unstable versions are requested or not.
func FilterVersions(versions []string) (results []string) {
	for _, version := range versions {
		if !semver.IsValid(version) {
			continue
		}
		results = append(results, version)
	}
	return
}

// SelectVersionStable returns the latest stable version from a list of
// versions. If there are no stable versions it will return an empty string.
func SelectVersionStable(versions []string) (v string) {
	for _, version := range FilterVersions(versions) {
		// Both build and pre should be blank for stable versions
		build := semver.Build(version)
		if build != "" {
			continue
		}

		pre := semver.Prerelease(version)
		if pre != "" {
			continue
		}

		c := semver.Compare(v, version)
		if c == -1 {
			v = version
		}
	}
	return
}

// SelectVersionAny returns the latest version from a list of versions including prereleases.
func SelectVersionAny(versions []string) (v string) {
	for _, version := range FilterVersions(versions) {
		c := semver.Compare(v, version)
		if c == -1 {
			v = version
		}
	}
	return
}

// SelectVersionAlpha specifically returns only -alpha tagged releases
func SelectVersionAlpha(versions []string) (v string) {
	for _, version := range FilterVersions(versions) {
		build := semver.Build(version)
		if build != "" {
			continue
		}

		pre := semver.Prerelease(version)
		if pre != "" && !strings.Contains(pre, "alpha") {
			continue
		}

		c := semver.Compare(v, version)
		if c == -1 {
			v = version
		}
	}
	return
}

// SelectVersionExperimental includes all prerelease tagged plugin versions, minus +build versions
func SelectVersionExperimental(versions []string) (v string) {
	for _, version := range FilterVersions(versions) {
		// All build releases are excluded from experimental
		build := semver.Build(version)
		if build != "" {
			continue
		}

		c := semver.Compare(v, version)
		if c == -1 {
			v = version
		}
	}
	return
}
