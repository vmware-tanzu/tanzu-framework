// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

// Distribution is an interface to download a single plugin binary.
type Distribution interface {
	// Fetch the binary for a plugin version.
	Fetch(version, os, arch string) ([]byte, error)

	// GetDigest returns the SHA256 hash of the binary for a plugin version.
	GetDigest(version, os, arch string) (string, error)
}
