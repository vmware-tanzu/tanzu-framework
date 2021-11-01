// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

// This block is for global feature constants, to allow them to be used more broadly
const (
	// FeatureContextAwareDiscovery determines whether to use legacy way of discovering plugins or
	// to use the new context-aware Plugin API based plugin discovery mechanism
	// Users can set this featureflag so that we can have context-aware plugin discovery be opt-in for now.
	FeatureContextAwareDiscovery = "features.global.context-aware-discovery"
)

// DefaultCliFeatureFlags is used to populate an initially empty config file with default values for feature flags.
// If a developer expects that their feature will be ready to release, they should create an entry here with a true
// value. If a developer has a beta feature they want to expose, but leave turned off by default, they should create
// an entry here with a false value. The keys MUST be in the format "features.<plugin>.<feature>" or initialization
// will fail. Note that "global" is a special value for <plugin> to be used for CLI-wide features.
var (
	DefaultCliFeatureFlags = map[string]bool{
		FeatureContextAwareDiscovery:                          false,
		"features.management-cluster.import":                  false,
		"features.management-cluster.export-from-confirm":     true,
		"features.management-cluster.standalone-cluster-mode": false,
		"features.global.use-context-aware-discovery":         common.IsContextAwareDiscoveryEnabled,
	}
)

const (
	// EnvConfigKey is the environment variable that points to a tanzu config.
	EnvConfigKey = "TANZU_CONFIG"

	// EnvEndpointKey is the environment variable that overrides the tanzu endpoint.
	EnvEndpointKey = "TANZU_ENDPOINT"

	//nolint:gosec // Avoid "hardcoded credentials" false positive.
	// EnvAPITokenKey is the environment variable that overrides the tanzu API token for global auth.
	EnvAPITokenKey = "TANZU_API_TOKEN"

	// ConfigName is the name of the config
	ConfigName = "config.yaml"
)

var (
	// LocalDirName is the name of the local directory in which tanzu state is stored.
	LocalDirName = ".config/tanzu"

	// legacyLocalDirName is the name of the old local directory in which to look for tanzu state. This will be
	// removed in the future in favor of LocalDirName.
	legacyLocalDirName = ".tanzu"
)

// LocalDir returns the local directory in which tanzu state is stored.
func LocalDir() (path string, err error) {
	return localDirPath(LocalDirName)
}

func legacyLocalDir() (path string, err error) {
	return localDirPath(legacyLocalDirName)
}

// localDirPath returns the full path of the directory name in which tanzu state is stored.
func localDirPath(dirname string) (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate local tanzu dir")
	}
	path = filepath.Join(home, dirname)
	return
}

// ClientConfigPath returns the tanzu config path, checking for environment overrides.
func ClientConfigPath() (path string, err error) {
	return configPath(LocalDir)
}

// legacyConfigPath returns the legacy tanzu config path, checking for environment overrides.
func legacyConfigPath() (path string, err error) {
	return configPath(legacyLocalDir)
}

// configPath constructs the full config path, checking for environment overrides.
func configPath(localDirGetter func() (string, error)) (path string, err error) {
	localDir, err := localDirGetter()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigKey)
	if !ok {
		path = filepath.Join(localDir, ConfigName)
		return
	}
	return
}

// NewClientConfig returns a new config.
func NewClientConfig() (*configv1alpha1.ClientConfig, error) {
	c := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}
	err := StoreClientConfig(c)
	if err != nil {
		return nil, err
	}

	err = populateDefaultCliFeatureValues(c, DefaultCliFeatureFlags)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func populateDefaultCliFeatureValues(c *configv1alpha1.ClientConfig, defaultCliFeatureFlags map[string]bool) error {
	for featureName, flagValue := range defaultCliFeatureFlags {
		plugin, flag, err := c.SplitFeaturePath(featureName)
		if err != nil {
			return err
		}
		addFeatureFlag(c, plugin, flag, flagValue)
	}
	return nil
}

func addFeatureFlag(c *configv1alpha1.ClientConfig, plugin, flag string, flagValue bool) {
	if c.ClientOptions == nil {
		c.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if c.ClientOptions.Features == nil {
		c.ClientOptions.Features = make(map[string]configv1alpha1.FeatureMap)
	}
	if c.ClientOptions.Features[plugin] == nil {
		c.ClientOptions.Features[plugin] = make(map[string]string)
	}
	c.ClientOptions.Features[plugin][flag] = strconv.FormatBool(flagValue)
}

// ClientConfigNotExistError is thrown when a tanzu config cannot be found.
type ClientConfigNotExistError struct {
	s string
}

// Error is the error message.
func (c *ClientConfigNotExistError) Error() string {
	return c.s
}

// NewConfigNotExistError returns a new ClientConfigNotExistError.
func NewConfigNotExistError(err error) *ClientConfigNotExistError {
	return &ClientConfigNotExistError{errors.Wrap(err, "failed to read config file").Error()}
}

// CopyLegacyConfigDir copies configuration files from legacy config dir to the new location. This is a no-op if the legacy dir
// does not exist or if the new config dir already exists.
func CopyLegacyConfigDir() error {
	legacyPath, err := legacyLocalDir()
	if err != nil {
		return err
	}
	legacyPathExists, err := fileExists(legacyPath)
	if err != nil {
		return err
	}
	newPath, err := LocalDir()
	if err != nil {
		return err
	}
	newPathExists, err := fileExists(newPath)
	if err != nil {
		return err
	}
	if legacyPathExists && !newPathExists {
		if err := copyDir(legacyPath, newPath); err != nil {
			return nil
		}
		log.Warningf("Configuration is now stored in %s. Legacy configuration directory %s is deprecated and will be removed in a future release.", newPath, legacyPath)
		log.Warningf("To complete migration, please remove legacy configuration directory %s and adjust your script(s), if any, to point to the new location.", legacyPath)
	}
	return nil
}

// GetClientConfig retrieves the config from the local directory.
func GetClientConfig() (cfg *configv1alpha1.ClientConfig, err error) {
	cfgPath, err := ClientConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		cfg, err = NewClientConfig()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	scheme, err := configv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	var c configv1alpha1.ClientConfig
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode config file")
	}

	added := addMissingDefaultFeatureFlags(&c, DefaultCliFeatureFlags)
	if added {
		_ = StoreClientConfig(&c)
	}

	return &c, nil
}

// addMissingDefaultFeatureFlags augments the given configuration object with any default feature flags that do not already have a value
// and returns TRUE if any were added (so the config can be written out to disk, if the caller wants to)
func addMissingDefaultFeatureFlags(config *configv1alpha1.ClientConfig, defaultFeatureFlags map[string]bool) bool {
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
func containsFeatureFlag(config *configv1alpha1.ClientConfig, plugin, feature string) bool {
	return config.ClientOptions != nil && config.ClientOptions.Features != nil && config.ClientOptions.Features[plugin] != nil &&
		config.ClientOptions.Features[plugin][feature] != ""
}

// storeConfigToLegacyDir stores configuration to legacy dir and logs warning in case of errors.
func storeConfigToLegacyDir(data []byte) {
	var (
		err                      error
		legacyDir, legacyCfgPath string
		legacyDirExists          bool
	)

	defer func() {
		if err != nil {
			log.Warningf("Failed to write config to legacy location for backward compatibility: %v", err)
			log.Warningf("To stop writing config to legacy location, please point your script(s), "+
				"if any, to the new config directory and remove legacy config directory %s", legacyDir)
		}
	}()

	legacyDir, err = legacyLocalDir()
	if err != nil {
		return
	}
	legacyDirExists, err = fileExists(legacyDir)
	if err != nil || !legacyDirExists {
		// Assume user has migrated and ignore writing to legacy location if that dir does not exist.
		return
	}
	legacyCfgPath, err = legacyConfigPath()
	if err != nil {
		return
	}
	err = os.WriteFile(legacyCfgPath, data, 0644)
}

// StoreClientConfig stores the config in the local directory.
func StoreClientConfig(cfg *configv1alpha1.ClientConfig) error {
	cfgPath, err := ClientConfigPath()
	if err != nil {
		return errors.Wrap(err, "could not find config path")
	}
	cfgPathExists, err := fileExists(cfgPath)
	if err != nil {
		return errors.Wrap(err, "failed to check config path existence")
	}
	if !cfgPathExists {
		localDir, err := LocalDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	}

	scheme, err := configv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	// Set GVK explicitly as encoder does not do it.
	cfg.GetObjectKind().SetGroupVersionKind(configv1alpha1.GroupVersionKind)
	buf := new(bytes.Buffer)
	if err := s.Encode(cfg, buf); err != nil {
		return errors.Wrap(err, "failed to encode config file")
	}
	// TODO (pbarker): need to consider races.
	if err = os.WriteFile(cfgPath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}
	storeConfigToLegacyDir(buf.Bytes())
	return nil
}

// DeleteClientConfig deletes the config from the local directory.
func DeleteClientConfig() error {
	cfgPath, err := ClientConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(cfgPath)
	if err != nil {
		return errors.Wrap(err, "could not remove config")
	}
	return nil
}

// GetServer by name.
func GetServer(name string) (s *configv1alpha1.Server, err error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return s, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == name {
			return server, nil
		}
	}
	return s, fmt.Errorf("could not find server %q", name)
}

// ServerExists tells whether the server by the given name exists.
func ServerExists(name string) (bool, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return false, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// AddServer adds a server to the config.
func AddServer(s *configv1alpha1.Server, setCurrent bool) error {
	cfg, err := GetClientConfig()
	if err != nil {
		return err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == s.Name {
			return fmt.Errorf("server %q already exists", s.Name)
		}
	}
	cfg.KnownServers = append(cfg.KnownServers, s)
	if setCurrent {
		cfg.CurrentServer = s.Name
	}
	return StoreClientConfig(cfg)
}

// PutServer adds or updates the server.
func PutServer(s *configv1alpha1.Server, setCurrent bool) error {
	cfg, err := GetClientConfig()
	if err != nil {
		return err
	}
	newServers := []*configv1alpha1.Server{s}
	for _, server := range cfg.KnownServers {
		if server.Name == s.Name {
			continue
		}
		newServers = append(newServers, server)
	}
	cfg.KnownServers = newServers
	if setCurrent {
		cfg.CurrentServer = s.Name
	}
	return StoreClientConfig(cfg)
}

// RemoveServer adds a server to the config.
func RemoveServer(name string) error {
	cfg, err := GetClientConfig()
	if err != nil {
		return err
	}

	newServers := []*configv1alpha1.Server{}
	for _, server := range cfg.KnownServers {
		if server.Name != name {
			newServers = append(newServers, server)
		}
	}
	cfg.KnownServers = newServers

	if cfg.CurrentServer == name {
		cfg.CurrentServer = ""
	}

	err = StoreClientConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// SetCurrentServer sets the current server.
func SetCurrentServer(name string) error {
	cfg, err := GetClientConfig()
	if err != nil {
		return err
	}
	var exists bool
	for _, server := range cfg.KnownServers {
		if server.Name == name {
			exists = true
		}
	}
	if !exists {
		return fmt.Errorf("could not set current server; %q is not a known server", name)
	}
	cfg.CurrentServer = name
	err = StoreClientConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentServer sets the current server.
func GetCurrentServer() (s *configv1alpha1.Server, err error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return s, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == cfg.CurrentServer {
			return server, nil
		}
	}
	return s, fmt.Errorf("current server %q not found in tanzu config", cfg.CurrentServer)
}

// EndpointFromServer returns the endpoint from server.
func EndpointFromServer(s *configv1alpha1.Server) (endpoint string, err error) {
	switch s.Type {
	case configv1alpha1.ManagementClusterServerType:
		return s.ManagementClusterOpts.Endpoint, nil
	case configv1alpha1.GlobalServerType:
		return s.GlobalOpts.Endpoint, nil
	default:
		return endpoint, fmt.Errorf("unknown server type %q", s.Type)
	}
}

// IsFeatureActivated returns true if the given feature is activated
// User can set this CLI feature flag using `tanzu config set features.global.<feature> true`
func IsFeatureActivated(feature string) bool {
	cfg, err := GetClientConfig()
	if err != nil {
		return false
	}
	status, err := cfg.IsConfigFeatureActivated(feature)
	if err != nil {
		return false
	}
	return status
}
