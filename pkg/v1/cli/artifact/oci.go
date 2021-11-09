// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

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
	filesMap, err := carvelhelpers.GetFilesMapFromImage(g.Image)
	if err != nil {
		return nil, errors.Wrap(err, "unable fetch plugin binary")
	}

	var bytesData []byte
	fileCount := 0

	for path, fileData := range filesMap {
		// Skip any testing related directory paths if bundled
		if utils.ContainsString(filepath.SplitList(path), "test") {
			continue
		}

		bytesData = fileData
		fileCount++
	}

	if fileCount != 1 {
		return nil, errors.Wrapf(err, "oci artifact image for plugin require to have only 1 file but found %v", fileCount)
	}

	return bytesData, nil
}
