// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfigpaths provides utilities to get info related to TKG configuration paths.
package tkgconfigpaths

type client struct {
	configDir string
}

// New creates new tkg configuration paths client
func New(configDir string) Client {
	tkgconfigpaths := &client{
		configDir: configDir,
	}
	return tkgconfigpaths
}

// Client implements TKG configuration paths functions
type Client interface {
	// GetTKGDirectory returns path to tkg config directory "$HOME/.tkg"
	GetTKGDirectory() (string, error)

	// GetTKGProvidersDirectory returns path to tkg config directory "$HOME/.tkg/providers"
	GetTKGProvidersDirectory() (string, error)

	// GetTKGProvidersCheckSumPath returns path to the providers checksum file
	GetTKGProvidersCheckSumPath() (string, error)

	// GetTKGCompatibilityDirectory returns path to tkg compatibility directory "<TKGConfigDirectory>/compatibility"
	GetTKGCompatibilityDirectory() (string, error)

	// GetTKGBoMDirectory returns path to tkg config directory "$HOME/.tkg/bom"
	GetTKGBoMDirectory() (string, error)

	// GetTKGConfigDirectories returns tkg config directories in below order
	// (tkgDir, bomDir, providersDir, error)
	GetTKGConfigDirectories() (string, string, string, error)

	// GetProvidersConfigFilePath returns config file path from providers dir
	// "$HOME/.tkg/providers/config.yaml"
	GetProvidersConfigFilePath() (string, error)

	// GetTKGConfigPath returns tkg configfile path
	GetTKGConfigPath() (string, error)

	// GetDefaultClusterConfigPath returns default cluster config file path
	GetDefaultClusterConfigPath() (string, error)

	// GetTKGCompatibilityConfigPath returns TKG compatibility file path
	GetTKGCompatibilityConfigPath() (string, error)

	// GetConfigDefaultsFilePath returns config_default.yaml file path under TKG directory
	GetConfigDefaultsFilePath() (string, error)

	// GetLogDirectory returns the directory path where log files should be stored by default.
	GetLogDirectory() (string, error)

	// GetClusterConfigurationDirectory returns the directory path where cluster configuration files will be stored
	GetClusterConfigurationDirectory() (string, error)
}
