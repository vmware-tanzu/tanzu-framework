// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

// Deprecated: This configuration variables are defined to support older plugins that are relying on
// this configuration to be set in the tanzu configuration file.
// This is pointing to the production registry to make sure existing plugins continue to work with
// newer version of the Tanzu CLI
const (
	tkgDefaultImageRepo              = "projects.registry.vmware.com/tkg"
	tkgDefaultCompatibilityImagePath = "tkg-compatibility"
)

func addCompatabilityFile(c *configapi.ClientConfig, compatibilityFilePath string) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configapi.CLIOptions{}
	}
	c.ClientOptions.CLI.CompatibilityFilePath = compatibilityFilePath
}

func addBomRepo(c *configapi.ClientConfig, repo string) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configapi.CLIOptions{}
	}
	c.ClientOptions.CLI.BOMRepo = repo
}

// addCompatibilityFileIfMissing adds the compatibility file to the client configuration to ensure it can be downloaded
func addCompatibilityFileIfMissing(config *configapi.ClientConfig) bool {
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.CompatibilityFilePath == "" {
		addCompatabilityFile(config, tkgDefaultCompatibilityImagePath)
		return true
	}
	return false
}

// addBomRepoIfMissing adds the bomRepository to the client configuration if it is not already present
func addBomRepoIfMissing(config *configapi.ClientConfig) bool {
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.BOMRepo == "" {
		addBomRepo(config, tkgDefaultImageRepo)
		return true
	}
	return false
}
