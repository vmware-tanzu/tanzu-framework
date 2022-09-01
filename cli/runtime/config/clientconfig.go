// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	c, err := GetClientConfigNoLock()
	if err != nil {
		log.Warningf("unable to get client config: %v", err)
	}

	addedDefaultDiscovery := populateDefaultStandaloneDiscovery(c)
	addedFeatureFlags := addDefaultFeatureFlagsIfMissing(c, DefaultCliFeatureFlags)
	addedEdition := addDefaultEditionIfMissing(c)
	addedBomRepo := addBomRepoIfMissing(c)
	addedCompatabilityFile := addCompatibilityFileIfMissing(c)
	// contexts could be lost when older plugins edit the config, so populate them from servers
	addedContexts := populateContexts(c)

	if addedFeatureFlags || addedDefaultDiscovery || addedEdition || addedCompatabilityFile || addedBomRepo || addedContexts {
		_ = StoreClientConfig(c)
	}
}

// This block is for global feature constants, to allow them to be used more broadly
const (
	// FeatureContextAwareCLIForPlugins determines whether to use legacy way of discovering plugins or
	// to use the new context-aware Plugin API based plugin discovery mechanism
	// Users can set this featureflag so that we can have context-aware plugin discovery be opt-in for now.
	FeatureContextAwareCLIForPlugins = "features.global.context-aware-cli-for-plugins"
	// FeatureContextCommand determines whether to surface the context command. This is disabled by default.
	FeatureContextCommand = "features.global.context-target"
	// DualStack feature flags determine whether it is permitted to create
	// clusters with a dualstack TKG_IP_FAMILY.  There are separate flags for
	// each primary, "ipv4,ipv6" vs "ipv6,ipv4", and flags for management vs
	// workload cluster plugins.
	FeatureFlagManagementClusterDualStackIPv4Primary = "features.management-cluster.dual-stack-ipv4-primary"
	FeatureFlagManagementClusterDualStackIPv6Primary = "features.management-cluster.dual-stack-ipv6-primary"
	FeatureFlagClusterDualStackIPv4Primary           = "features.cluster.dual-stack-ipv4-primary"
	FeatureFlagClusterDualStackIPv6Primary           = "features.cluster.dual-stack-ipv6-primary"
	// Custom Nameserver feature flags determine whether it is permitted to
	// provide the CONTROL_PLANE_NODE_NAMESERVERS and WORKER_NODE_NAMESERVERS
	// when creating a cluster.
	FeatureFlagManagementClusterCustomNameservers = "features.management-cluster.custom-nameservers"
	FeatureFlagClusterCustomNameservers           = "features.cluster.custom-nameservers"
	// AWS Instance Types Exclude ARM feature flags determine whether instance types with processor architecture
	// support of ARM should be included when discovering available AWS instance types. Setting feature flag to true
	// filters out ARM supporting instance types; false allows ARM instance types to be included in results.
	FeatureFlagAwsInstanceTypesExcludeArm = "features.management-cluster.aws-instance-types-exclude-arm"
	// PackageBasedLCM feature flag determines whether to use package based lifecycle management of management component
	// or legacy way of managing management components. This is also used for clusterclass based management and workload
	// cluster provisioning
	FeatureFlagPackageBasedLCM = "features.global.package-based-lcm-beta"
	// TKR version v1alpha3 feature flag determines whether to use Tanzu Kubernetes Release API version v1alpha3. Setting
	// feature flag to true will allow to use the TKR version v1alpha3; false allows to use legacy TKR version v1alpha1
	FeatureFlagTKRVersionV1Alpha3 = "features.global.tkr-version-v1alpha3-beta"
	// Package Plugin Kctrl Command Tree determines whether to use the command tree from kctrl. Setting feature flag to
	// true will allow to use the package command tree from kctrl for package plugin
	FeatureFlagPackagePluginKctrlCommandTree = "features.package.kctrl-package-command-tree"
	// FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration feature flag determines whether to auto-apply the generated ClusterClass
	// based configuration after converting legacy configration to ClusterClass based config or not
	// Note: This is a hidden feature-flag that doesn't get persisted to config.yaml by default
	FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration = "features.cluster.auto-apply-generated-clusterclass-based-configuration"
	// FeatureFlagForceDeployClusterWithClusterClass if this feature flag is set CLI will try to deploy ClusterClass
	// based cluster even if user has done any customization to the provider templates
	// Note: This is a hidden feature-flag that doesn't get persisted to config.yaml by default
	FeatureFlagForceDeployClusterWithClusterClass = "features.cluster.force-deploy-cluster-with-clusterclass"
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
		FeatureContextAwareCLIForPlugins:                      common.ContextAwareDiscoveryEnabled(),
		FeatureContextCommand:                                 false,
		"features.management-cluster.import":                  false,
		"features.management-cluster.export-from-confirm":     true,
		"features.management-cluster.standalone-cluster-mode": false,
		FeatureFlagManagementClusterDualStackIPv4Primary:      false,
		FeatureFlagManagementClusterDualStackIPv6Primary:      false,
		FeatureFlagClusterDualStackIPv4Primary:                false,
		FeatureFlagClusterDualStackIPv6Primary:                false,
		FeatureFlagManagementClusterCustomNameservers:         false,
		FeatureFlagClusterCustomNameservers:                   false,
		FeatureFlagAwsInstanceTypesExcludeArm:                 true,
		FeatureFlagTKRVersionV1Alpha3:                         false,
		FeatureFlagPackagePluginKctrlCommandTree:              false,
	}
>>>>>>> e60ec549 (Moving pkg/v1/tkg as a top level module):pkg/v1/config/clientconfig.go
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
			CLI: &configv1alpha1.CLIOptions{},
		},
	}

	// Check if the lock is acquired by the current process or not
	// If not try to acquire the lock before Storing the client config
	// and release the lock after updating the config
	if !IsTanzuConfigLockAcquired() {
		AcquireTanzuConfigLock()
		defer ReleaseTanzuConfigLock()
	}

	err := StoreClientConfig(c)
	if err != nil {
		return nil, err
	}

	return c, nil
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

// GetClientConfig retrieves the config from the local directory with file lock
func GetClientConfig() (cfg *configv1alpha1.ClientConfig, err error) {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	return GetClientConfigNoLock()
}

// GetClientConfigNoLock retrieves the config from the local directory without acquiring the lock
func GetClientConfigNoLock() (cfg *configv1alpha1.ClientConfig, err error) {
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

	return &c, nil
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
// Make sure to Acquire and Release tanzu lock when reading/writing to the
// tanzu client configuration
func StoreClientConfig(cfg *configv1alpha1.ClientConfig) error {
	// new plugins would be setting only contexts, so populate servers for backwards compatibility
	populateServers(cfg)
	// old plugins would be setting only servers, so populate contexts for forwards compatibility
	PopulateContexts(cfg)

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

	if !IsTanzuConfigLockAcquired() {
		return errors.New("error while updating the tanzu config file, lock is not acquired for updating tanzu config file")
	}

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
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
	if err != nil {
		return err
	}

	for _, server := range cfg.KnownServers {
		if server.Name == s.Name {
			return fmt.Errorf("server %q already exists", s.Name)
		}
	}

	cfg.KnownServers = append(cfg.KnownServers, s)
	c := convertServerToContext(s)
	cfg.KnownContexts = append(cfg.KnownContexts, c)

	if setCurrent {
		cfg.CurrentServer = s.Name
		err = cfg.SetCurrentContext(c.Type, c.Name)
		if err != nil {
			return err
		}
	}
	return StoreClientConfig(cfg)
}

// PutServer adds or updates the server.
func PutServer(s *configv1alpha1.Server, setCurrent bool) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
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

	c := convertServerToContext(s)
	newContexts := []*configv1alpha1.Context{c}
	for _, ctx := range cfg.KnownContexts {
		if ctx.Name == c.Name {
			continue
		}
		newContexts = append(newContexts, ctx)
	}
	cfg.KnownContexts = newContexts

	if setCurrent {
		cfg.CurrentServer = s.Name
		err = cfg.SetCurrentContext(c.Type, c.Name)
		if err != nil {
			return err
		}
	}
	return StoreClientConfig(cfg)
}

// RemoveServer adds a server to the config.
func RemoveServer(name string) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
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

	var c *configv1alpha1.Context
	newContexts := []*configv1alpha1.Context{}
	for _, ctx := range cfg.KnownContexts {
		if ctx.Name != name {
			newContexts = append(newContexts, ctx)
		} else {
			c = ctx
		}
	}
	cfg.KnownContexts = newContexts

	if cfg.CurrentServer == name {
		cfg.CurrentServer = ""
	}

	if cfg.CurrentContext[c.Type] == name {
		delete(cfg.CurrentContext, c.Type)
	}

	err = StoreClientConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// SetCurrentServer sets the current server.
func SetCurrentServer(name string) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
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

	c, err := cfg.GetContext(name)
	if err != nil {
		return err
	}
	err = cfg.SetCurrentContext(c.Type, c.Name)
	if err != nil {
		return err
	}

	err = StoreClientConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentServer gets the current server.
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

// EndpointFromContext returns the endpoint from context.
func EndpointFromContext(s *configv1alpha1.Context) (endpoint string, err error) {
	switch s.Type {
	case configv1alpha1.CtxTypeK8s:
		return s.ClusterOpts.Endpoint, nil
	case configv1alpha1.CtxTypeTMC:
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

// GetDiscoverySources returns all discovery sources
// Includes standalone discovery sources and if server is available
// it also includes context based discovery sources as well
func GetDiscoverySources(serverName string) []configv1alpha1.PluginDiscovery {
	server, err := GetServer(serverName)
	if err != nil {
		log.Warningf("unknown server '%s', Unable to get server based discovery sources: %s", serverName, err.Error())
		return []configv1alpha1.PluginDiscovery{}
	}

	discoverySources := server.DiscoverySources
	// If current server type is management-cluster, then add
	// the default kubernetes discovery endpoint pointing to the
	// management-cluster kubeconfig
	if server.Type == configv1alpha1.ManagementClusterServerType {
		defaultClusterK8sDiscovery := configv1alpha1.PluginDiscovery{
			Kubernetes: &configv1alpha1.KubernetesDiscovery{
				Name:    fmt.Sprintf("default-%s", serverName),
				Path:    server.ManagementClusterOpts.Path,
				Context: server.ManagementClusterOpts.Context,
			},
		}
		discoverySources = append(discoverySources, defaultClusterK8sDiscovery)
	}
	return discoverySources
}

// GetEnvConfigurations returns a map of configured environment variables
// to values as part of tanzu configuration file
// it returns nil if configuration is not yet defined
func GetEnvConfigurations() map[string]string {
	cfg, err := GetClientConfig()
	if err != nil {
		return nil
	}
	return cfg.GetEnvConfigurations()
}

// GetEdition returns the edition from the local configuration file
func GetEdition() (string, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return "", err
	}
	if cfg != nil && cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil {
		return string(cfg.ClientOptions.CLI.Edition), nil
	}
	return "", nil
}

// GetDefaultRepo returns the bomRepo set in the client configuration. If it
// cannot be resolved, the default repo set at build time is returned along
// with an error describing why the bomRepo could not be resolved from the
// client configuration.
func GetDefaultRepo() (string, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return "", err
	}
	if cfg == nil {
		return "", fmt.Errorf("client configuration is empty")
	}
	if cfg.ClientOptions == nil {
		return "", fmt.Errorf("client options missing from client configuration")
	}
	if cfg.ClientOptions.CLI == nil {
		return "", fmt.Errorf("CLI settings are missing from client options in client configuration")
	}
	if cfg.ClientOptions.CLI.BOMRepo == "" {
		return "", fmt.Errorf("bom repo is missing from CLI settings in the client configuration")
	}
	return cfg.ClientOptions.CLI.BOMRepo, nil
}

// GetCompatibilityFilePath returns the compatibilityPath set in the client
// configuration. If it cannot be resolved, the default path set at build time
// is returned along with an error describing why the path could not be
// resolved from the client configuration.
func GetCompatibilityFilePath() (string, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return "", err
	}
	if cfg == nil {
		return "", fmt.Errorf("client configuration is empty")
	}
	if cfg.ClientOptions == nil {
		return "", fmt.Errorf("client options missing from client configuration")
	}
	if cfg.ClientOptions.CLI == nil {
		return "", fmt.Errorf("CLI settings are missing from client options in client configuration")
	}
	if cfg.ClientOptions.CLI.CompatibilityFilePath == "" {
		return "", fmt.Errorf("compatibility file is missing from CLI settings in the client configuration")
	}
	return cfg.ClientOptions.CLI.CompatibilityFilePath, nil
}
