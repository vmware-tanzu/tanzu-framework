// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package distribution implements plugin distribution interface
// Distribution is the interface to download a plugin version binary
// for a given OS and architecture combination.
// E.g., a GCP object location, an OCI compliant image repository, etc.
package distribution

// Distribution is an interface to download a single plugin binary.
type Distribution interface {
	// Fetch the binary for a plugin version.
	Fetch(version, os, arch string) ([]byte, error)

	// GetDigest returns the SHA256 hash of the binary for a plugin version.
	GetDigest(version, os, arch string) (string, error)

	// DescribeArtifact returns the artifact resource based plugin metadata
	DescribeArtifact(version, os, arch string) (Artifact, error)
}
