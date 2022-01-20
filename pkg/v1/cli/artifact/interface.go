// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package artifact implements interface to fetch the binary artifacts
// from different sources
package artifact

import (
	"net/url"
	"path/filepath"
)

// Artifact is an interface to download a single plugin binary.
type Artifact interface {
	// Fetch the binary for a plugin version.
	Fetch() ([]byte, error)
}

// NewURIArtifact creates new artifacts based on the URI
func NewURIArtifact(uri string) (Artifact, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case uriSchemeHTTP, uriSchemeHTTPS:
		return NewHTTPArtifact(uri), nil
	case uriSchemeLocal:
		return NewLocalArtifact(filepath.Join(u.Host, u.Path)), nil
	default:
		// The URI could point to a relative path without specifying any scheme
		// as prefix. Hence, defaulting to local artifact.
		return NewLocalArtifact(uri), nil
	}
}
