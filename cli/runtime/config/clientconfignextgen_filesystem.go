// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

const (
	// EnvConfigNextGenKey is the environment variable that points to a tanzu config.
	EnvConfigNextGenKey = "TANZU_CONFIG_NEXT_GEN"

	// CfgNextGenName is the name of the config metadata
	CfgNextGenName = "config-ng.yaml"
)

// clientConfigNextGenPath constructs the full config path, checking for environment overrides.
func clientConfigNextGenPath(localDirGetter func() (string, error)) (path string, err error) {
	localDir, err := localDirGetter()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigNextGenKey)
	if !ok {
		path = filepath.Join(localDir, CfgNextGenName)
		return
	}
	return
}

// ClientConfigNextGenPath retrieved config-alt file path
func ClientConfigNextGenPath() (path string, err error) {
	return clientConfigNextGenPath(LocalDir)
}
