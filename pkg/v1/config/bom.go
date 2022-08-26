// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
)

func addCompatabilityFile(c *configv1alpha1.ClientConfig, compatibilityFilePath string) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configv1alpha1.CLIOptions{}
	}
	c.ClientOptions.CLI.CompatibilityFilePath = compatibilityFilePath
}

func addBomRepo(c *configv1alpha1.ClientConfig, repo string) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configv1alpha1.CLIOptions{}
	}
	c.ClientOptions.CLI.BOMRepo = repo
}

// addCompatibilityFileIfMissing adds the compatibility file to the client configuration to ensure it can be downloaded
func addCompatibilityFileIfMissing(config *configv1alpha1.ClientConfig) bool {
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.CompatibilityFilePath == "" {
		addCompatabilityFile(config, tkgconfigpaths.TKGDefaultCompatibilityImagePath)
		return true
	}
	return false
}

// addBomRepoIfMissing adds the bomRepository to the client configuration if it is not already present
func addBomRepoIfMissing(config *configv1alpha1.ClientConfig) bool {
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.BOMRepo == "" {
		addBomRepo(config, tkgconfigpaths.TKGDefaultImageRepo)
		return true
	}
	return false
}
