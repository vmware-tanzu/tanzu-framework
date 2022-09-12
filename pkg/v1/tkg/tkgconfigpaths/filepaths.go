// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigpaths

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// GetTKGDirectory returns path to tkg config directory "$HOME/.tkg"
func (c *client) GetTKGDirectory() (string, error) {
	if c.configDir == "" {
		return "", errors.New("tkg config directory is empty")
	}
	return c.configDir, nil
}

// GetTKGProvidersDirectory returns path to tkg config directory "$HOME/.tkg/providers"
func (c *client) GetTKGProvidersDirectory() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.LocalProvidersFolderName), nil
}

// GetTKGProvidersCheckSumPath returns path to the providers checksum file
func (c *client) GetTKGProvidersCheckSumPath() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.LocalProvidersFolderName, constants.LocalProvidersChecksumFileName), nil
}

// GetTKGBoMDirectory returns path to tkg config directory "$HOME/.tkg/bom"
func (c *client) GetTKGBoMDirectory() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.LocalBOMsFolderName), nil
}

// GetTKGCompatibilityDirectory returns path to tkg compatibility directory "<TKGConfigDirectory>/compatibility"
func (c *client) GetTKGCompatibilityDirectory() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.LocalCompatibilityFolderName), nil
}

// GetTKGConfigDirectories returns tkg config directories in below order
// (tkgDir, bomDir, providersDir, error)
func (c *client) GetTKGConfigDirectories() (string, string, string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", "", "", err
	}
	bomDir := filepath.Join(tkgDir, constants.LocalBOMsFolderName)
	providerDir := filepath.Join(tkgDir, constants.LocalProvidersFolderName)
	return tkgDir, bomDir, providerDir, nil
}

// GetProvidersConfigFilePath returns config file path from providers dir
// "$HOME/.tkg/providers/config.yaml"
func (c *client) GetProvidersConfigFilePath() (string, error) {
	providersDir, err := c.GetTKGProvidersDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(providersDir, constants.LocalProvidersConfigFileName), nil
}

// GetTKGCompatibilityConfigPath returns TKG compatibility file path
func (c *client) GetTKGCompatibilityConfigPath() (string, error) {
	compatibilityDir, err := c.GetTKGCompatibilityDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(compatibilityDir, constants.TKGCompatibilityFileName), nil
}

// GetTKGConfigPath returns tkg configfile path
func (c *client) GetTKGConfigPath() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.TKGConfigFileName), nil
}

// GetDefaultClusterConfigPath returns default cluster config file path
func (c *client) GetDefaultClusterConfigPath() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.TKGDefaultClusterConfigFileName), nil
}

// GetOverridesDirectory returns path of overrides directory
func GetOverridesDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return constants.OverrideFolder
	}
	return filepath.Join(homeDir, constants.OverrideFolder)
}

// GetRegistryCertFile returns the registry cert file path
func GetRegistryCertFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate user home dir")
	}
	return path.Join(home, constants.TKGRegistryCertFile), nil
}

// GetRegistryTrustedCACertFileForWindows returns the registry trusted root ca cert filepath for windows
func GetRegistryTrustedCACertFileForWindows() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate user home dir")
	}
	return path.Join(home, constants.TKGRegistryTrustedRootCAFileForWindows), nil
}

// GetConfigDefaultsFilePath returns config_default.yaml file path under TKG directory
func (c *client) GetConfigDefaultsFilePath() (string, error) {
	tkgDir, err := c.GetTKGProvidersDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.TKGConfigDefaultFileName), nil
}

// GetLogDirectory returns the directory path where log files should be stored by default.
func (c *client) GetLogDirectory() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.LogFolderName), nil
}

// GetClusterConfigurationDirectory returns the directory path where cluster configuration files will be stored
func (c *client) GetClusterConfigurationDirectory() (string, error) {
	tkgDir, err := c.GetTKGDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(tkgDir, constants.TKGClusterConfigFileDirForUI), nil
}
