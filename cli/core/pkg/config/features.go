// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strconv"

	"github.com/pkg/errors"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	configlib "github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

// ConfigureDefaultFeatureFlagsIfMissing configures default feature-flags to ClientConfig if missing
func ConfigureDefaultFeatureFlagsIfMissing(defaultFeatureFlags map[string]bool) error {
	// Acquire tanzu config lock
	configlib.AcquireTanzuConfigLock()
	defer configlib.ReleaseTanzuConfigLock()

	c, err := configlib.GetClientConfigNoLock()
	if err != nil {
		return errors.Wrap(err, "error while getting client config")
	}

	configUpdated := AddDefaultFeatureFlagsIfMissing(c, defaultFeatureFlags)
	if !configUpdated {
		return nil
	}

	err = configlib.StoreClientConfig(c)
	if err != nil {
		return errors.Wrap(err, "error while storing client config")
	}
	return nil
}

func populateDefaultCliFeatureValues(c *configapi.ClientConfig, defaultCliFeatureFlags map[string]bool) error {
	for featureName, flagValue := range defaultCliFeatureFlags {
		plugin, flag, err := c.SplitFeaturePath(featureName)
		if err != nil {
			return err
		}
		addFeatureFlag(c, plugin, flag, flagValue)
	}
	return nil
}

func addFeatureFlag(c *configapi.ClientConfig, plugin, flag string, flagValue bool) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.Features == nil {
		c.ClientOptions.Features = make(map[string]configapi.FeatureMap)
	}
	if c.ClientOptions.Features[plugin] == nil {
		c.ClientOptions.Features[plugin] = make(map[string]string)
	}
	c.ClientOptions.Features[plugin][flag] = strconv.FormatBool(flagValue)
}

// AddDefaultFeatureFlagsIfMissing augments the given configuration object with any default feature flags that do not already have a value
// and returns TRUE if any were added (so the config can be written out to disk, if the caller wants to)
func AddDefaultFeatureFlagsIfMissing(config *configapi.ClientConfig, defaultFeatureFlags map[string]bool) bool {
	added := false

	for featurePath, activated := range defaultFeatureFlags {
		plugin, feature, err := config.SplitFeaturePath(featurePath)
		if err == nil && !containsFeatureFlag(config, plugin, feature) {
			addFeatureFlag(config, plugin, feature, activated)
			added = true
		}
	}

	return added
}

// containsFeatureFlag returns true if the features section in the configuration object contains any value for the plugin.feature combination
func containsFeatureFlag(config *configapi.ClientConfig, plugin, feature string) bool {
	return config.ClientOptions != nil && config.ClientOptions.Features != nil && config.ClientOptions.Features[plugin] != nil &&
		config.ClientOptions.Features[plugin][feature] != ""
}
