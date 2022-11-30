// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

const (
	// EnvConfigKey is the environment variable that points to a tanzu config.
	EnvConfigKey = "TANZU_CONFIG"

	// ConfigName is the name of the config
	ConfigName = "config.yaml"
)

// ClientConfigPath returns the tanzu config path, checking for environment overrides.
func ClientConfigPath() (path string, err error) {
	return configPath(LocalDir)
}

// configPath constructs the full config path, checking for environment overrides.
func configPath(localDirGetter func() (string, error)) (path string, err error) {
	localDir, err := localDirGetter()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigKey)
	if !ok {
		path = filepath.Join(localDir, ConfigName)
		return
	}
	return
}
