// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

// Artifact is an interface to download a single plugin binary.
type Artifact interface {
	// Fetch the binary for a plugin version.
	Fetch() ([]byte, error)
}

// NewURIArtifact creates new artifacs based on the URI
func NewURIArtifact(path string) Artifact {
	return nil
}
