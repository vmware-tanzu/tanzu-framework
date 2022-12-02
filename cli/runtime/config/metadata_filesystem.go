// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

const (
	// EnvConfigMetadataKey is the environment variable that points to a tanzu config.
	EnvConfigMetadataKey = "TANZU_CONFIG_METADATA"

	// CfgMetadataName is the name of the config metadata hidden file
	CfgMetadataName = ".config-metadata.yaml"
)

// metadataPath constructs the full config path, checking for environment overrides.
func metadataPath(localDirGetter func() (string, error)) (path string, err error) {
	localDir, err := localDirGetter()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigMetadataKey)
	if !ok {
		path = filepath.Join(localDir, CfgMetadataName)
		return
	}
	return
}

func CfgMetadataFilePath() (path string, err error) {
	return metadataPath(LocalDir)
}
