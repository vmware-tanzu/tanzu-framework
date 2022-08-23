// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// A file map getter takes an OCI image name and returns a map representing the
// file contents.
type fileMapGetterFn func(string) (map[string][]byte, error)

// OCIArtifact defines OCI artifact image endpoint
type OCIArtifact struct {
	Image                string
	getFilesMapFromImage fileMapGetterFn
}

// NewOCIArtifact creates OCI Artifact object
func NewOCIArtifact(image string) Artifact {
	return &OCIArtifact{
		Image:                image,
		getFilesMapFromImage: carvelhelpers.GetFilesMapFromImage,
	}
}

// Fetch an artifact.
func (g *OCIArtifact) Fetch() ([]byte, error) {
	filesMap, err := g.getFilesMapFromImage(g.Image)
	if err != nil {
		return nil, errors.Wrap(err, "unable fetch plugin binary")
	}

	var bytesData []byte
	fileCount := 0

	for path, fileData := range filesMap {
		// Skip any testing related directory paths if bundled
		if utils.ContainsString(strings.Split(path, "/"), "test") {
			continue
		}

		bytesData = fileData
		fileCount++
	}

	if fileCount != 1 {
		return nil, fmt.Errorf("oci artifact image for plugin is required to have only 1 file, but found %v", fileCount)
	}

	return bytesData, nil
}
