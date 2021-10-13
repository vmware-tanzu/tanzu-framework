// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

type OCIArtifact struct {
	Image string
}

func NewOCIArtifact(image string) Artifact {
	return &OCIArtifact{
		Image: image,
	}
}

// Fetch an artifact.
func (g *OCIArtifact) Fetch() ([]byte, error) {
	return nil, nil
}
