// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func cleanupDir(dir string) {
	p, _ := localDirPath(dir)
	_ = os.RemoveAll(p)
}

func randString() string {
	return uuid.NewV4().String()[:5]
}

func TestClientConfig(t *testing.T) {
	LocalDirName = fmt.Sprintf(".tanzu-test-%s", randString())
	server0 := &configv1alpha1.Server{
		Name: "test",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &configv1alpha1.ClientConfig{
		KnownServers: []*configv1alpha1.Server{
			server0,
		},
		CurrentServer: "test",
	}

	err := StoreClientConfig(testCtx)
	require.NoError(t, err)

	defer cleanupDir(LocalDirName)

	_, err = GetClientConfig()
	require.NoError(t, err)

	s, err := GetServer("test")
	require.NoError(t, err)

	require.Equal(t, s, server0)

	e, err := ServerExists("test")
	require.NoError(t, err)
	require.True(t, e)

	server1 := &configv1alpha1.Server{
		Name: "test1",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test1",
		},
	}

	err = AddServer(server1, true)
	require.NoError(t, err)

	c, err := GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 2)
	require.Equal(t, c.CurrentServer, "test1")

	err = RemoveServer("test")
	require.NoError(t, err)

	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 1)

	err = RemoveServer("test1")
	require.NoError(t, err)

	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 0)
	require.Equal(t, c.CurrentServer, "")

	err = DeleteClientConfig()
	require.NoError(t, err)
}

func TestConfigLegacyDir(t *testing.T) {
	r := randString()
	LocalDirName = fmt.Sprintf(".tanzu-test-%s", r)

	// Setup legacy config dir.
	legacyLocalDirName = fmt.Sprintf(".tanzu-test-legacy-%s", r)
	legacyLocalDir, err := legacyLocalDir()
	require.NoError(t, err)
	err = os.MkdirAll(legacyLocalDir, 0755)
	require.NoError(t, err)
	legacyCfgPath, err := legacyConfigPath()
	require.NoError(t, err)

	server0 := &configv1alpha1.Server{
		Name: "test",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &configv1alpha1.ClientConfig{
		KnownServers: []*configv1alpha1.Server{
			server0,
		},
		CurrentServer: "test",
	}

	err = StoreClientConfig(testCtx)
	require.NoError(t, err)
	require.FileExists(t, legacyCfgPath)

	defer cleanupDir(LocalDirName)
	defer cleanupDir(legacyLocalDirName)

	_, err = GetClientConfig()
	require.NoError(t, err)

	server1 := &configv1alpha1.Server{
		Name: "test1",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test1",
		},
	}

	err = AddServer(server1, true)
	require.NoError(t, err)

	c, err := GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 2)
	require.Equal(t, c.CurrentServer, "test1")

	err = RemoveServer("test")
	require.NoError(t, err)

	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 1)

	tmp := LocalDirName
	LocalDirName = legacyLocalDirName
	configCopy, err := GetClientConfig()
	require.NoError(t, err)
	if diff := cmp.Diff(c, configCopy); diff != "" {
		t.Errorf("ClientConfig object mismatch between legacy and new config location (-want +got): \n%s", diff)
	}
	LocalDirName = tmp

	err = DeleteClientConfig()
	require.NoError(t, err)
}

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
