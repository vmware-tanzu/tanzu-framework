// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

var (
	// IsContextAwareDiscoveryEnabled defines default to use when the user has not configured a value
	// This variable is configured at the build time of the CLI
	IsContextAwareDiscoveryEnabled = ""
)

// This block is for global feature constants, to allow them to be used more broadly
const (
	// FeatureContextAwareCLIForPlugins determines whether to use legacy way of discovering plugins or
	// to use the new context-aware Plugin API based plugin discovery mechanism
	// Users can set this featureflag so that we can have context-aware plugin discovery be opt-in for now.
	FeatureContextAwareCLIForPlugins = "features.global.context-aware-cli-for-plugins"
	// FeatureContextCommand determines whether to surface the context command. This is disabled by default.
	FeatureContextCommand = "features.global.context-target"
)

// DefaultCliFeatureFlags is used to populate an initially empty config file with default values for feature flags.
// The keys MUST be in the format "features.<plugin>.<feature>" or initialization
// will fail. Note that "global" is a special value for <plugin> to be used for CLI-wide features.
//
// If a developer expects that their feature will be ready to release, they should create an entry here with a true
// value.
// If a developer has a beta feature they want to expose, but leave turned off by default, they should create
// an entry here with a false value. WE HIGHLY RECOMMEND the use of a SEPARATE flag for beta use; one that ends in "-beta".
// Thus, if you plan to eventually release a feature with a flag named "features.cluster.foo-bar", you should consider
// releasing the beta version with "features.cluster.foo-bar-beta". This will make it much easier when it comes time for
// mainstreaming the feature (with a default true value) under the flag name "features.cluster.foo-bar", as there will be
// no conflict with previous installs (that have a false value for the entry "features.cluster.foo-bar-beta").
var (
	DefaultCliFeatureFlags = map[string]bool{
		FeatureContextAwareCLIForPlugins: ContextAwareDiscoveryEnabled(),
		FeatureContextCommand:            false,
	}
)

// ContextAwareDiscoveryEnabled returns true if the IsContextAwareDiscoveryEnabled
// is set to true during build time
func ContextAwareDiscoveryEnabled() bool {
	return strings.EqualFold(IsContextAwareDiscoveryEnabled, "true")
}

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
