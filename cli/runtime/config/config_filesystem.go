// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (

	// EnvEndpointKey is the environment variable that overrides the tanzu endpoint.
	EnvEndpointKey = "TANZU_ENDPOINT"

	//nolint:gosec // Avoid "hardcoded credentials" false positive.
	// EnvAPITokenKey is the environment variable that overrides the tanzu API token for global auth.
	EnvAPITokenKey = "TANZU_API_TOKEN"
)

var (
	// LocalDirName is the name of the local directory in which tanzu state is stored.
	LocalDirName = ".config/tanzu"
	// TestLocalDirName is the name of the local directory in which tanzu state is stored for testing.
	TestLocalDirName = ".tanzu-test"
)

// LocalDir returns the local directory in which tanzu state is stored.
func LocalDir() (path string, err error) {
	return localDirPath(LocalDirName)
}

// localDirPath returns the full path of the directory name in which tanzu state is stored.
func localDirPath(dirname string) (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate local tanzu dir")
	}
	path = filepath.Join(home, dirname)
	return
}
