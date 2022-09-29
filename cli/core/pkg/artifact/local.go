// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
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

// FetchTest reads test plugin artifact based on the local plugin artifact path
// To fetch the test binary from the plugin we are using plugin binary path and creating test plugin path from it
// If the plugin binary path is `artifacts/darwin/amd64/cli/cluster/v0.27.0-dev/tanzu-cluster-darwin_amd64`
// then test plugin binary will be under `artifacts/darwin/amd64/cli/cluster/v0.27.0-dev/test/` directory
func (l *LocalArtifact) FetchTest() ([]byte, error) {
	// test plugin directory based on the plugin binary location
	testPluginDir := filepath.Join(filepath.Dir(l.Path), "test")
	dirEntries, err := os.ReadDir(testPluginDir)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read test plugin directory")
	}
	if len(dirEntries) != 1 {
		return nil, errors.Errorf("expected only 1 file under the '%v' directory, but found %v files", testPluginDir, len(dirEntries))
	}

	// Assuming the file under the test directory is the test plugin binary
	testPluginFile := dirEntries[0]
	if testPluginFile.IsDir() {
		return nil, errors.Errorf("expected to find test plugin binary but found directory %q", testPluginFile.Name())
	}

	testPluginPath := filepath.Join(testPluginDir, testPluginFile.Name())
	b, err := os.ReadFile(testPluginPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading test artifact")
	}
	return b, nil
}
