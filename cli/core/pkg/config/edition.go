// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func addEdition(c *configapi.ClientConfig, edition configapi.EditionSelector) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configapi.CLIOptions{}
	}
	c.ClientOptions.CLI.Edition = edition //nolint:staticcheck
}

// addDefaultEditionIfMissing returns true if the default edition was added to the configuration (because there was no edition)
func addDefaultEditionIfMissing(config *configapi.ClientConfig) bool {
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.Edition == "" { //nolint:staticcheck
		addEdition(config, DefaultEdition)
		return true
	}
	return false
}
