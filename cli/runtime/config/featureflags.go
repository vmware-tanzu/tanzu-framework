// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strconv"

	"github.com/pkg/errors"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func ConfigureDefaultFeatureFlagsIfMissing(defaultFeatureFlags map[string]bool) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	c, err := GetClientConfigNoLock()
	if err != nil {
		return errors.Wrap(err, "error while getting client config")
	}

	configUpdated := UpdateDefaultFeatureFlagsIfMissing(c, defaultFeatureFlags)
	if !configUpdated {
		return nil
	}

	err = StoreClientConfig(c)
	if err != nil {
		return errors.Wrap(err, "error while storing client config")
	}
	return nil
}

// UpdateDefaultFeatureFlagsIfMissing updates clientConfig if the featureflags are missing
func UpdateDefaultFeatureFlagsIfMissing(c *configapi.ClientConfig, defaultFeatureFlags map[string]bool) bool {
	added := false
	for featurePath, activated := range defaultFeatureFlags {
		plugin, feature, err := c.SplitFeaturePath(featurePath)
		if err == nil && !containsFeatureFlag(c, plugin, feature) {
			addFeatureFlag(c, plugin, feature, activated)
			added = true
		}
	}
	return added
}

// containsFeatureFlag returns true if the features section in the configuration object contains any value for the plugin.feature combination
func containsFeatureFlag(c *configapi.ClientConfig, plugin, feature string) bool {
	return c.ClientOptions != nil && c.ClientOptions.Features != nil && c.ClientOptions.Features[plugin] != nil &&
		c.ClientOptions.Features[plugin][feature] != ""
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
