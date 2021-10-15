// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

// LocalArtifact defines local artifact path
type LocalArtifact struct {
	Path string
}

// NewLocalArtifact creates Local Artifact object
func NewLocalArtifact(path string) Artifact {
	// If path is not an absolute path
	// search under `xdg.ConfigHome/tanzu-plugin/localPath` directory
	if !filepath.IsAbs(path) {
		path = filepath.Join(common.DefaultLocalPluginDistroDir, "distribution", path)
	}
	return &LocalArtifact{
		Path: path,
	}
}

// Fetch an artifact.
func (l *LocalArtifact) Fetch() ([]byte, error) {
	b, err := os.ReadFile(l.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading manifest file")
	}
	return b, nil
}
