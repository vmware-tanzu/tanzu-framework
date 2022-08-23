// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

const (
	uriSchemeLocal = "file"
)

// LocalArtifact defines local artifact path.
// Sample URI: file://home/user/workspace/tanzu-framework/artifacts/darwin/amd64/cli/login/v0.10.0-dev/tanzu-login-darwin_amd64
type LocalArtifact struct {
	// Path is path to local binary artifact
	// if path is not an absolute path search under
	// `xdg.ConfigHome/tanzu-plugin/localPath` directory
	Path string
}

// NewLocalArtifact creates Local Artifact object
// If path is not an absolute path
// search under `xdg.ConfigHome/tanzu-plugin/distribution` directory
func NewLocalArtifact(path string) Artifact {
	if !filepath.IsAbs(path) {
		path = filepath.Join(common.DefaultLocalPluginDistroDir, "distribution", path)
	}
	return &LocalArtifact{
		Path: path,
	}
}

// Fetch reads the local artifact from its path
func (l *LocalArtifact) Fetch() ([]byte, error) {
	b, err := os.ReadFile(l.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading artifact")
	}
	return b, nil
}
