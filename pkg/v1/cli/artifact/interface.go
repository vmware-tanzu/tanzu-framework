// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package artifact implements interface to fetch the binary artifacts
// from different sources
package artifact

// Artifact is an interface to download a single plugin binary.
type Artifact interface {
	// Fetch the binary for a plugin version.
	Fetch() ([]byte, error)
}

// NewURIArtifact creates new artifacts based on the URI
func NewURIArtifact(uri string) Artifact {
	// TODO: Support other artifact types
	return NewLocalArtifact(uri)
}
