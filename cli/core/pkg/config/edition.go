// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func addEdition(c *configv1alpha1.ClientConfig, edition configv1alpha1.EditionSelector) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configv1alpha1.CLIOptions{}
	}
	c.ClientOptions.CLI.Edition = edition
}

// addDefaultEditionIfMissing returns true if the default edition was added to the configuration (because there was no edition)
func addDefaultEditionIfMissing(config *configv1alpha1.ClientConfig) bool {
	if config.ClientOptions == nil || config.ClientOptions.CLI == nil || config.ClientOptions.CLI.Edition == "" {
		addEdition(config, DefaultEdition)
		return true
	}
	return false
}
