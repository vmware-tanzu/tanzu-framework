// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"

// Deprecated: This configuration variables are defined to support older plugins that are relying on
// this configuration to be set in the tanzu configuration file.
// This is pointing to the production registry to make sure existing plugins continue to work with
// newer version of the Tanzu CLI
const (
	tkgDefaultImageRepo              = "projects.registry.vmware.com/tkg"
	tkgDefaultCompatibilityImagePath = "tkg-compatibility"
)

func addCompatibilityFile(c *configapi.ClientConfig, compatibilityFilePath string) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configapi.CLIOptions{}
	}
	// CompatibilityFilePath has been deprecated and will be removed from future version
	c.ClientOptions.CLI.CompatibilityFilePath = compatibilityFilePath //nolint:staticcheck
}

func addBomRepo(c *configapi.ClientConfig, repo string) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configapi.CLIOptions{}
	}
	// BOMRepo has been deprecated and will be removed from future version
	c.ClientOptions.CLI.BOMRepo = repo //nolint:staticcheck
}

// AddCompatibilityFileIfMissing adds the compatibility file to the client configuration to ensure it can be downloaded
func AddCompatibilityFileIfMissing(config *configapi.ClientConfig) bool {
	// CompatibilityFilePath has been deprecated and will be removed from future version
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.CompatibilityFilePath == "" { //nolint:staticcheck
		addCompatibilityFile(config, tkgDefaultCompatibilityImagePath)
		return true
	}
	return false
}

// AddBomRepoIfMissing adds the bomRepository to the client configuration if it is not already present
func AddBomRepoIfMissing(config *configapi.ClientConfig) bool {
	// BOMRepo has been deprecated and will be removed from future version
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.BOMRepo == "" { //nolint:staticcheck
		addBomRepo(config, tkgDefaultImageRepo)
		return true
	}
	return false
}
