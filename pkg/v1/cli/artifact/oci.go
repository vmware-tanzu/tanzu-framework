// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

// OCIArtifact defines OCI artifact image endpoint
type OCIArtifact struct {
	Image string
}

// NewOCIArtifact creates OCI Artifact object
func NewOCIArtifact(image string) Artifact {
	return &OCIArtifact{
		Image: image,
	}
}

// Fetch an artifact.
func (g *OCIArtifact) Fetch() ([]byte, error) {
	// TODO(anujc25): implement OCI artifact fetch
	return nil, nil
}
