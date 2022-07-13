// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"golang.org/x/sync/errgroup"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

func cleanupDir(dir string) {
	p, _ := localDirPath(dir)
	_ = os.RemoveAll(p)
}

func randString() string {
	return uuid.NewString()[:5]
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
	AcquireTanzuConfigLock()
	err := StoreClientConfig(testCtx)
	require.NoError(t, err)
	ReleaseTanzuConfigLock()

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

	AcquireTanzuConfigLock()
	err = StoreClientConfig(testCtx)
	ReleaseTanzuConfigLock()
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

func TestConfigFeaturesDefaultEditionAdded(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}

	added := addDefaultEditionIfMissing(cfg)
	require.True(t, added, "addDefaultEditionIfMissing should have returned true (having added missing default edition value)")
	errMsg := "addDefaultEditionIfMissing should have added default edition (" + configv1alpha1.EditionStandard + ") instead of " + cfg.ClientOptions.CLI.Edition
	require.Equal(t, cfg.ClientOptions.CLI.Edition, configv1alpha1.EditionSelector(configv1alpha1.EditionStandard), errMsg)
}

func TestConfigFeaturesDefaultEditionNotAdded(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
				Edition:                 "tce",
			},
		},
	}

	added := addDefaultEditionIfMissing(cfg)
	require.False(t, added, "addDefaultEditionIfMissing should have returned false (without adding default edition value)")
	errMsg := "addDefaultEditionIfMissing should have left existing edition value intact instead of replacing with [" + cfg.ClientOptions.CLI.Edition + "]"
	require.Equal(t, cfg.ClientOptions.CLI.Edition, configv1alpha1.EditionSelector(configv1alpha1.EditionCommunity), errMsg)
}

func TestConfigPopulateDefaultStandaloneDiscovery(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				DiscoverySources: []configv1alpha1.PluginDiscovery{},
			},
		},
	}
	configureTestDefaultStandaloneDiscoveryOCI()

	assert := assert.New(t)

	added := populateDefaultStandaloneDiscovery(cfg)
	assert.Equal(true, added)
	assert.Equal(len(cfg.ClientOptions.CLI.DiscoverySources), 1)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Name, DefaultStandaloneDiscoveryName)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Image, "fake.image.repo/package/standalone-plugins:v1.0.0")
}

func TestConfigPopulateDefaultStandaloneDiscoveryWhenDefaultDiscoveryExistsAndIsSame(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				DiscoverySources: []configv1alpha1.PluginDiscovery{
					configv1alpha1.PluginDiscovery{
						OCI: &configv1alpha1.OCIDiscovery{
							Name:  DefaultStandaloneDiscoveryName,
							Image: "fake.image.repo/package/standalone-plugins:v1.0.0",
						},
					},
				},
			},
		},
	}
	configureTestDefaultStandaloneDiscoveryOCI()

	assert := assert.New(t)

	added := populateDefaultStandaloneDiscovery(cfg)
	assert.Equal(false, added)
	assert.Equal(len(cfg.ClientOptions.CLI.DiscoverySources), 1)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Name, DefaultStandaloneDiscoveryName)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Image, "fake.image.repo/package/standalone-plugins:v1.0.0")
}

func TestConfigPopulateDefaultStandaloneDiscoveryWhenDefaultDiscoveryExistsAndIsNotSame(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				DiscoverySources: []configv1alpha1.PluginDiscovery{
					configv1alpha1.PluginDiscovery{
						OCI: &configv1alpha1.OCIDiscovery{
							Name:  DefaultStandaloneDiscoveryName,
							Image: "fake.image/path:v2.0.0",
						},
					},
					configv1alpha1.PluginDiscovery{
						OCI: &configv1alpha1.OCIDiscovery{
							Name:  "additional-discovery",
							Image: "additional-discovery/path:v1.0.0",
						},
					},
				},
			},
		},
	}
	configureTestDefaultStandaloneDiscoveryOCI()

	assert := assert.New(t)

	added := populateDefaultStandaloneDiscovery(cfg)
	assert.Equal(true, added)
	assert.Equal(len(cfg.ClientOptions.CLI.DiscoverySources), 2)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Name, DefaultStandaloneDiscoveryName)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Image, "fake.image.repo/package/standalone-plugins:v1.0.0")
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[1].OCI.Name, "additional-discovery")
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[1].OCI.Image, "additional-discovery/path:v1.0.0")
}

func TestConfigPopulateDefaultStandaloneDiscoveryLocal(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				DiscoverySources: []configv1alpha1.PluginDiscovery{},
			},
		},
	}

	configureTestDefaultStandaloneDiscoveryLocal()

	assert := assert.New(t)

	added := populateDefaultStandaloneDiscovery(cfg)
	assert.Equal(true, added)
	assert.Equal(len(cfg.ClientOptions.CLI.DiscoverySources), 1)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].Local.Name, DefaultStandaloneDiscoveryNameLocal)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].Local.Path, "local/path")
}

func TestConfigPopulateDefaultStandaloneDiscoveryEnvVariables(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				DiscoverySources: []configv1alpha1.PluginDiscovery{},
			},
		},
	}

	configureTestDefaultStandaloneDiscoveryOCI()

	os.Setenv(constants.ConfigVariableCustomImageRepository, "env.fake.image.repo")
	os.Setenv(constants.ConfigVariableDefaultStandaloneDiscoveryImagePath, "package/env/standalone-plugins")
	os.Setenv(constants.ConfigVariableDefaultStandaloneDiscoveryImageTag, "v2.0.0")

	assert := assert.New(t)

	added := populateDefaultStandaloneDiscovery(cfg)
	assert.Equal(true, added)
	assert.Equal(len(cfg.ClientOptions.CLI.DiscoverySources), 1)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Name, DefaultStandaloneDiscoveryName)
	assert.Equal(cfg.ClientOptions.CLI.DiscoverySources[0].OCI.Image, "env.fake.image.repo/package/env/standalone-plugins:v2.0.0")
}

func TestGetDiscoverySources(t *testing.T) {
	assert := assert.New(t)

	tanzuConfigBytes := `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    useContextAwareDiscovery: true
current: mgmt
kind: ClientConfig
metadata:
  creationTimestamp: null
servers:
- managementClusterOpts:
    context: mgmt-admin@mgmt
    path: config
  name: mgmt
  type: managementcluster
`
	f, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(err)
	err = os.WriteFile(f.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(err)
	defer os.Remove(f.Name())
	os.Setenv("TANZU_CONFIG", f.Name())

	pds := GetDiscoverySources("mgmt")
	assert.Equal(1, len(pds))
	assert.Equal(pds[0].Kubernetes.Name, "default-mgmt")
	assert.Equal(pds[0].Kubernetes.Path, "config")
	assert.Equal(pds[0].Kubernetes.Context, "mgmt-admin@mgmt")
}

func configureTestDefaultStandaloneDiscoveryOCI() {
	DefaultStandaloneDiscoveryType = "oci"
	DefaultStandaloneDiscoveryRepository = "fake.image.repo"
	DefaultStandaloneDiscoveryImagePath = "package/standalone-plugins"
	DefaultStandaloneDiscoveryImageTag = "v1.0.0"
}

func configureTestDefaultStandaloneDiscoveryLocal() {
	DefaultStandaloneDiscoveryType = "local"
	DefaultStandaloneDiscoveryLocalPath = "local/path"
}

func TestClientConfigUpdateInParallel(t *testing.T) {
	assert := assert.New(t)
	addServer := func(mcName string) error {
		_, err := GetClientConfig()
		if err != nil {
			return err
		}

		s := &configv1alpha1.Server{
			Name: mcName,
			Type: configv1alpha1.ManagementClusterServerType,
			ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
				Context: "fake-context",
				Path:    "fake-path",
			},
		}
		err = AddServer(s, true)
		if err != nil {
			return err
		}

		_, err = GetClientConfig()
		return err
	}

	// Creates temp configuration file and runs addServer in parallel
	runTestInParallel := func() {
		// Get the temp tanzu config file
		f, err := os.CreateTemp("", "tanzu_config*")
		assert.Nil(err)
		os.Setenv("TANZU_CONFIG", f.Name())

		// run addServer in parallel
		parallelExecutionCounter := 100
		group, _ := errgroup.WithContext(context.Background())
		for i := 1; i <= parallelExecutionCounter; i++ {
			id := i
			group.Go(func() error {
				return addServer(fmt.Sprintf("mc-%v", id))
			})
		}
		err = group.Wait()
		assert.Nil(err)

		// Make sure that the configuration file is not corrupted
		clientconfig, err := GetClientConfig()
		assert.Nil(err)

		// Make sure all expected servers are added to the knownServers list
		assert.Equal(parallelExecutionCounter, len(clientconfig.KnownServers))
	}

	// Run the parallel tests of reading and updating the configuration file
	// multiple times to make sure all the attempts are successful
	for testCounter := 1; testCounter <= 5; testCounter++ {
		log.Infof("Running parallel test #%v", testCounter)
		runTestInParallel()
	}
}
