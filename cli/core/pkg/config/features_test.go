// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func TestConfigFeatures(t *testing.T) {
	const pluginName = "management-cluster"
	const featureName = "foo"
	const featurePath = "features." + pluginName + "." + featureName
	cliFeatureFlags := configv1alpha1.FeatureMap{
		featureName: "true",
	}
	cliFeatureMap := make(map[string]configv1alpha1.FeatureMap)
	cliFeatureMap[pluginName] = cliFeatureFlags
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
			Features: cliFeatureMap,
		},
	}
	activated, err := cfg.IsConfigFeatureActivated(featurePath)
	require.True(t, activated, "IsConfigFeatureActivated should report true for feature "+featurePath)
	require.NoError(t, err)
}

func TestConfigFeaturesDefault(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}
	const featureFoo = "features.management-cluster.foo"
	activated, err := cfg.IsConfigFeatureActivated(featureFoo)
	require.False(t, activated, "feature reported true before defaults are  set")
	require.NoError(t, err)

	cliFeatureFlags := map[string]bool{
		featureFoo: true,
	}
	err = populateDefaultCliFeatureValues(cfg, cliFeatureFlags)

	require.NoError(t, err)
	activated, err = cfg.IsConfigFeatureActivated(featureFoo)
	require.True(t, activated, "feature "+featureFoo+" should report true after defaults are set")
	require.NoError(t, err)

	const featureBar = "features.management-cluster.bar"
	activated, err = cfg.IsConfigFeatureActivated(featureBar)
	require.False(t, activated, "feature "+featureBar+" should report false after defaults are set")
	require.NoError(t, err)
}

func TestConfigFeaturesDefaultInvalid(t *testing.T) {
	const featureFoo = "invalid.foo"
	cliFeatureFlags := map[string]bool{
		featureFoo: true,
	}
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}
	err := populateDefaultCliFeatureValues(cfg, cliFeatureFlags)
	require.Error(t, err, "invalid default feature should generate error")
}

func TestConfigFeaturesInvalidName(t *testing.T) {
	const featureFoo = "invalid.foo"
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}
	result, err := cfg.IsConfigFeatureActivated(featureFoo)
	require.False(t, result, "invalid feature name '"+featureFoo+"' should generate false return value")
	require.Error(t, err, "invalid feature name '"+featureFoo+"' should generate error")
}

func TestConfigFeaturesInvalidValue(t *testing.T) {
	const featureFoo = "features.management-cluster.foo"
	cliFeatureFlags := configv1alpha1.FeatureMap{
		"foo": "INVALID",
	}
	cliFeatureMap := make(map[string]configv1alpha1.FeatureMap)
	cliFeatureMap["management-cluster"] = cliFeatureFlags
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
			Features: cliFeatureMap,
		},
	}
	activated, err := cfg.IsConfigFeatureActivated(featureFoo)
	require.False(t, activated, "IsConfigFeatureActivated should report false given invalid value")
	require.Error(t, err, "IsConfigFeatureActivated should return error given invalid value")
}

func TestConfigFeaturesSplitName(t *testing.T) {
	const featureValid = "features.valid-plugin.foo"
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}
	pluginName, featureName, err := cfg.SplitFeaturePath(featureValid)
	require.Equal(t, pluginName, "valid-plugin", "failed to parse '"+featureValid+"' correctly")
	require.Equal(t, featureName, "foo", "failed to parse '"+featureValid+"' correctly")
	require.NoError(t, err, "valid feature name '"+featureValid+"' should not generate error")
}

func TestConfigFeaturesSplitNameInvalid(t *testing.T) {
	const featureInvalid = "invalid.foo"
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}
	_, _, err := cfg.SplitFeaturePath(featureInvalid)
	require.Error(t, err, "invalid feature name '"+featureInvalid+"' should generate error")
}

func TestConfigFeaturesDefaultsAdded(t *testing.T) {
	defaultFeatureFlags := map[string]bool{
		"features.global.truthy":   true,
		"features.global.falsey":   false,
		"features.existing.truthy": true,
		"features.existing.falsey": false,
	}
	// NOTE: the existing values are OPPOSITE of the default and should stay that way:
	cliFeatureFlags := configv1alpha1.FeatureMap{
		"truthy": "false",
		"falsey": "true",
	}
	cliFeatureMap := make(map[string]configv1alpha1.FeatureMap)
	cliFeatureMap["existing"] = cliFeatureFlags
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
			Features: cliFeatureMap,
		},
	}

	added := addDefaultFeatureFlagsIfMissing(cfg, defaultFeatureFlags)
	require.True(t, added, "addDefaultFeatureFlagsIfMissing should have added missing default values")
	require.Equal(t, cfg.ClientOptions.Features["existing"]["truthy"], "false", "addDefaultFeatureFlagsIfMissing should have left existing FALSE value for truthy")
	require.Equal(t, cfg.ClientOptions.Features["existing"]["falsey"], "true", "addDefaultFeatureFlagsIfMissing should have left existing TRUE value for falsey")
	require.Equal(t, cfg.ClientOptions.Features["global"]["truthy"], "true", "addDefaultFeatureFlagsIfMissing should have added global TRUE value for truthy")
	require.Equal(t, cfg.ClientOptions.Features["global"]["falsey"], "false", "addDefaultFeatureFlagsIfMissing should have added global FALSE value for falsey")
}

func TestConfigFeaturesDefaultsNoneAdded(t *testing.T) {
	defaultFeatureFlags := map[string]bool{
		"features.existing.truthy": true,
		"features.existing.falsey": false,
	}
	// NOTE: the existing values are OPPOSITE of the default and should stay that way:
	cliFeatureFlags := configv1alpha1.FeatureMap{
		"truthy": "false",
		"falsey": "true",
	}
	cliFeatureMap := make(map[string]configv1alpha1.FeatureMap)
	cliFeatureMap["existing"] = cliFeatureFlags
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
			Features: cliFeatureMap,
		},
	}

	added := addDefaultFeatureFlagsIfMissing(cfg, defaultFeatureFlags)
	require.False(t, added, "addDefaultFeatureFlagsIfMissing should NOT have added any default values")
	require.Equal(t, cfg.ClientOptions.Features["existing"]["truthy"], "false", "addDefaultFeatureFlagsIfMissing should have left existing FALSE value for truthy")
	require.Equal(t, cfg.ClientOptions.Features["existing"]["falsey"], "true", "addDefaultFeatureFlagsIfMissing should have left existing TRUE value for falsey")
}
