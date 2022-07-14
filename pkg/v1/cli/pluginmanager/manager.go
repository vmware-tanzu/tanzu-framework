// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pluginmanager is resposible for plugin discovery and installation
package pluginmanager

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/artifact"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/catalog"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/discovery"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

const (
	// exe is an executable file extension
	exe = ".exe"
	// ManifestFileName is the file name for the manifest.
	ManifestFileName = "manifest.yaml"
	// PluginFileName is the file name for the plugin descriptor.
	PluginFileName = "plugin.yaml"
)

var execCommand = exec.Command

// ValidatePlugin validates the plugin descriptor.
func ValidatePlugin(p *cliv1alpha1.PluginDescriptor) (err error) {
	// skip builder plugin for bootstrapping
	if p.Name == "builder" {
		return nil
	}
	if p.Name == "" {
		err = multierr.Append(err, errors.New("plugin name cannot be empty"))
	}
	if p.Version == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q version cannot be empty", p.Name))
	}
	if !semver.IsValid(p.Version) && p.Version != "dev" {
		err = multierr.Append(err, fmt.Errorf("version %q %q is not a valid semantic version", p.Name, p.Version))
	}
	if p.Description == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q description cannot be empty", p.Name))
	}
	if p.Group == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q group cannot be empty", p.Name))
	}
	return
}

func discoverPlugins(pd []v1alpha1.PluginDiscovery) ([]plugin.Discovered, error) {
	allPlugins := make([]plugin.Discovered, 0)
	for _, d := range pd {
		discObject, err := discovery.CreateDiscoveryFromV1alpha1(d)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create discovery")
		}

		plugins, err := discObject.List()
		if err != nil {
			log.Warningf("unable to list plugin from discovery '%v': %v", discObject.Name(), err.Error())
			continue
		}
		allPlugins = append(allPlugins, plugins...)
	}
	return allPlugins, nil
}

// DiscoverStandalonePlugins returns the available standalone plugins
func DiscoverStandalonePlugins() (plugins []plugin.Discovered, err error) {
	cfg, e := config.GetClientConfig()
	if e != nil {
		err = errors.Wrapf(e, "unable to get client configuration")
		return
	}

	if cfg == nil || cfg.ClientOptions == nil || cfg.ClientOptions.CLI == nil {
		plugins = []plugin.Discovered{}
		return
	}

	plugins, err = discoverPlugins(cfg.ClientOptions.CLI.DiscoverySources)
	if err != nil {
		return
	}

	for i := range plugins {
		plugins[i].Scope = common.PluginScopeStandalone
		plugins[i].Status = common.PluginStatusNotInstalled
	}
	return
}

// DiscoverServerPlugins returns the available plugins associated with the given server
func DiscoverServerPlugins(serverName string) ([]plugin.Discovered, error) {
	plugins := []plugin.Discovered{}
	if serverName == "" {
		// If servername is not specified than returning empty list
		// as there are no server plugins that can be discovered
		return plugins, nil
	}

	discoverySources := config.GetDiscoverySources(serverName)
	plugins, err := discoverPlugins(discoverySources)
	if err != nil {
		return plugins, err
	}
	for i := range plugins {
		plugins[i].Scope = common.PluginScopeContext
		plugins[i].Status = common.PluginStatusNotInstalled
	}
	return plugins, nil
}

// DiscoverPlugins returns the available plugins that can be used with the given server
// If serverName is empty(""), return only standalone plugins
func DiscoverPlugins(serverName string) ([]plugin.Discovered, []plugin.Discovered) {
	serverPlugins, err := DiscoverServerPlugins(serverName)
	if err != nil {
		log.Warningf("unable to discover server plugins, %v", err.Error())
	}

	standalonePlugins, err := DiscoverStandalonePlugins()
	if err != nil {
		log.Warningf("unable to discover standalone plugins, %v", err.Error())
	}

	// TODO(anuj): Remove duplicate plugins with server plugins getting higher priority
	return serverPlugins, standalonePlugins
}

// AvailablePlugins returns the list of available plugins including discovered and installed plugins
// If serverName is empty(""), return only available standalone plugins
func AvailablePlugins(serverName string) ([]plugin.Discovered, error) {
	discoveredServerPlugins, discoveredStandalonePlugins := DiscoverPlugins(serverName)
	return availablePlugins(serverName, discoveredServerPlugins, discoveredStandalonePlugins)
}

// AvailablePluginsFromLocalSource returns the list of available plugins from local source
func AvailablePluginsFromLocalSource(localPath string) ([]plugin.Discovered, error) {
	localStandalonePlugins, err := DiscoverPluginsFromLocalSource(localPath)
	if err != nil {
		log.Warningf("Unable to discover standalone plugins from local source, %v", err.Error())
	}
	return availablePlugins("", []plugin.Discovered{}, localStandalonePlugins)
}

func availablePlugins(serverName string, discoveredServerPlugins, discoveredStandalonePlugins []plugin.Discovered) ([]plugin.Discovered, error) {
	installedSeverPluginDesc, installedStandalonePluginDesc, err := InstalledPlugins(serverName)
	if err != nil {
		return nil, err
	}

	availablePlugins := availablePluginsFromStandaloneAndServerPlugins(discoveredServerPlugins, discoveredStandalonePlugins)

	setAvailablePluginsStatus(availablePlugins, installedSeverPluginDesc)
	setAvailablePluginsStatus(availablePlugins, installedStandalonePluginDesc)

	installedButNotDiscoveredPlugins := getInstalledButNotDiscoveredStandalonePlugins(availablePlugins, installedStandalonePluginDesc)
	availablePlugins = append(availablePlugins, installedButNotDiscoveredPlugins...)

	return availablePlugins, nil
}

func getInstalledButNotDiscoveredStandalonePlugins(availablePlugins []plugin.Discovered, installedPluginDesc []cliv1alpha1.PluginDescriptor) []plugin.Discovered {
	var newPlugins []plugin.Discovered
	for i := range installedPluginDesc {
		found := false
		for j := range availablePlugins {
			if installedPluginDesc[i].Name == availablePlugins[j].Name {
				found = true
				// If plugin is installed but marked as not installed as part of availablePlugins list
				// mark the plugin as installed
				// This is possible if user has used --local mode to install the plugin which is also
				// getting discovered from the configured discovery sources
				if availablePlugins[j].Status == common.PluginStatusNotInstalled {
					availablePlugins[j].Status = common.PluginStatusInstalled
				}
			}
		}
		if !found {
			p := DiscoveredFromPluginDescriptor(&installedPluginDesc[i])
			p.Scope = common.PluginScopeStandalone
			p.Status = common.PluginStatusInstalled
			newPlugins = append(newPlugins, p)
		}
	}
	return newPlugins
}

// DiscoveredFromPluginDescriptor returns discovered plugin object from k8sV1alpha1
func DiscoveredFromPluginDescriptor(p *cliv1alpha1.PluginDescriptor) plugin.Discovered {
	dp := plugin.Discovered{
		Name:               p.Name,
		Description:        p.Description,
		RecommendedVersion: p.Version,
		Source:             p.Discovery,
		SupportedVersions:  []string{p.Version},
	}
	return dp
}

func setAvailablePluginsStatus(availablePlugins []plugin.Discovered, installedPluginDesc []cliv1alpha1.PluginDescriptor) {
	for i := range installedPluginDesc {
		for j := range availablePlugins {
			if installedPluginDesc[i].Name == availablePlugins[j].Name {
				// Match found, Check for update available and update status
				if installedPluginDesc[i].DiscoveredRecommendedVersion == availablePlugins[j].RecommendedVersion {
					availablePlugins[j].Status = common.PluginStatusInstalled
				} else {
					availablePlugins[j].Status = common.PluginStatusUpdateAvailable
				}
				availablePlugins[j].InstalledVersion = installedPluginDesc[i].Version
			}
		}
	}
}

func availablePluginsFromStandaloneAndServerPlugins(discoveredServerPlugins, discoveredStandalonePlugins []plugin.Discovered) []plugin.Discovered {
	availablePlugins := discoveredServerPlugins

	// Check whether the default standalone discovery type is local or not
	isLocalStandaloneDiscovery := config.GetDefaultStandaloneDiscoveryType() == common.DiscoveryTypeLocal

	for i := range discoveredStandalonePlugins {
		matchIndex := pluginIndexForName(availablePlugins, discoveredStandalonePlugins[i].Name)

		// Add the standalone plugin to available plugins if it doesn't exist in the serverPlugins list
		// OR
		// Current standalone discovery or plugin discovered is of type 'local'
		// We are overriding the discovered plugins that we got from server in case of 'local' discovery type
		// to allow developers to use the plugins that are built locally and not returned from the server
		// This local discovery is only used for development purpose and should not be used for production
		if matchIndex < 0 {
			availablePlugins = append(availablePlugins, discoveredStandalonePlugins[i])
			continue
		}
		if isLocalStandaloneDiscovery || discoveredStandalonePlugins[i].DiscoveryType == common.DiscoveryTypeLocal { // matchIndex >= 0 is guaranteed here
			availablePlugins[matchIndex] = discoveredStandalonePlugins[i]
		}
	}
	return availablePlugins
}

func pluginIndexForName(availablePlugins []plugin.Discovered, pluginName string) int {
	for j := range availablePlugins {
		if pluginName == availablePlugins[j].Name {
			return j
		}
	}
	return -1 // haven't found a match
}

// InstalledPlugins returns the installed plugins.
// If serverName is empty(""), return only installed standalone plugins
func InstalledPlugins(serverName string, exclude ...string) (serverPlugins, standalonePlugins []cliv1alpha1.PluginDescriptor, err error) {
	var serverCatalog, standAloneCatalog *catalog.ContextCatalog

	if serverName != "" {
		serverCatalog, err = catalog.NewContextCatalog(serverName)
		if err != nil {
			return nil, nil, err
		}
		serverPlugins = serverCatalog.List()
	}

	standAloneCatalog, err = catalog.NewContextCatalog("")
	if err != nil {
		return nil, nil, err
	}
	standalonePlugins = standAloneCatalog.List()
	return
}

// DescribePlugin describes a plugin.
// If serverName is empty(""), only consider standalone plugins
func DescribePlugin(serverName, pluginName string) (desc *cliv1alpha1.PluginDescriptor, err error) {
	c, err := catalog.NewContextCatalog(serverName)
	if err != nil {
		return nil, err
	}
	descriptor, ok := c.Get(pluginName)
	if !ok {
		err = fmt.Errorf("could not get plugin path for plugin %q", pluginName)
	}

	return &descriptor, err
}

// InitializePlugin initializes the plugin configuration
// If serverName is empty(""), only consider standalone plugins
func InitializePlugin(serverName, pluginName string) error {
	c, err := catalog.NewContextCatalog(serverName)
	if err != nil {
		return err
	}
	descriptor, ok := c.Get(pluginName)
	if !ok {
		return errors.Wrapf(err, "could not get plugin path for plugin %q", pluginName)
	}

	b, err := execCommand(descriptor.InstallationPath, "post-install").CombinedOutput()

	// Note: If user is installing old version of plugin than it is possible that
	// the plugin does not implement post-install command. Ignoring the
	// errors if the command does not exist for a particular plugin.
	if err != nil && !strings.Contains(string(b), "unknown command") {
		log.Warningf("Warning: Failed to initialize plugin '%q' after installation. %v", pluginName, string(b))
	}

	return nil
}

// InstallPlugin installs a plugin from the given repository.
// If serverName is empty(""), only consider standalone plugins
func InstallPlugin(serverName, pluginName, version string) error {
	availablePlugins, err := AvailablePlugins(serverName)
	if err != nil {
		return err
	}
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			if availablePlugins[i].Scope == common.PluginScopeStandalone {
				serverName = ""
			}
			return installOrUpgradePlugin(serverName, &availablePlugins[i], version)
		}
	}

	return errors.Errorf("unable to find plugin '%v'", pluginName)
}

// UpgradePlugin upgrades a plugin from the given repository.
// If serverName is empty(""), only consider standalone plugins
func UpgradePlugin(serverName, pluginName, version string) error {
	availablePlugins, err := AvailablePlugins(serverName)
	if err != nil {
		return err
	}
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			if availablePlugins[i].Scope == common.PluginScopeStandalone {
				serverName = ""
			}
			return installOrUpgradePlugin(serverName, &availablePlugins[i], version)
		}
	}

	return errors.Errorf("unable to find plugin '%v'", pluginName)
}

// GetRecommendedVersionOfPlugin returns recommended version of the plugin
// If serverName is empty(""), only consider standalone plugins
func GetRecommendedVersionOfPlugin(serverName, pluginName string) (string, error) {
	availablePlugins, err := AvailablePlugins(serverName)
	if err != nil {
		return "", err
	}
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			return availablePlugins[i].RecommendedVersion, nil
		}
	}
	return "", errors.Errorf("unable to find plugin '%v'", pluginName)
}

func installOrUpgradePlugin(serverName string, p *plugin.Discovered, version string) error {
	log.Infof("Installing plugin '%v:%v'", p.Name, version)

	// verify plugin before download
	err := verifyPluginPreDownload(p)
	if err != nil {
		return errors.Wrapf(err, "%q plugin pre-download verification failed", p.Name)
	}

	b, err := p.Distribution.Fetch(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	// verify plugin after download but before installation
	d, err := p.Distribution.GetDigest(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	err = verifyPluginPostDownload(p, d, b)
	if err != nil {
		return errors.Wrapf(err, "%q plugin post-download verification failed", p.Name)
	}

	pluginName := p.Name
	pluginPath := filepath.Join(common.DefaultPluginRoot, pluginName, version)

	err = os.MkdirAll(filepath.Dir(pluginPath), os.ModePerm)
	if err != nil {
		return err
	}

	if common.BuildArch().IsWindows() {
		pluginPath += exe
	}

	err = os.WriteFile(pluginPath, b, 0755)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	b, err = execCommand(pluginPath, "info").Output()
	if err != nil {
		return errors.Wrapf(err, "could not describe plugin %q", pluginName)
	}
	var descriptor cliv1alpha1.PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		return errors.Wrapf(err, "could not unmarshal plugin %q description", pluginName)
	}
	descriptor.InstallationPath = pluginPath
	descriptor.Discovery = p.Source
	descriptor.DiscoveredRecommendedVersion = p.RecommendedVersion

	c, err := catalog.NewContextCatalog(serverName)
	if err != nil {
		return err
	}
	err = c.Upsert(&descriptor)
	if err != nil {
		log.Info("Plugin descriptor could not be updated in cache")
	}
	err = InitializePlugin(serverName, pluginName)
	if err != nil {
		log.Infof("could not initialize plugin after installing: %v", err.Error())
	}
	return nil
}

// DeletePlugin deletes a plugin.
// If serverName is empty(""), only consider standalone plugins
func DeletePlugin(serverName, pluginName string) error {
	c, err := catalog.NewContextCatalog(serverName)
	if err != nil {
		return err
	}
	_, ok := c.Get(pluginName)
	if !ok {
		return fmt.Errorf("could not get plugin path for plugin %q", pluginName)
	}

	err = c.Delete(pluginName)
	if err != nil {
		return fmt.Errorf("plugin %q could not be deleted from cache", pluginName)
	}

	// TODO: delete the plugin binary if it is not used by any server

	return nil
}

// SyncPlugins automatically downloads all available plugins to users machine
// If serverName is empty(""), only sync standalone plugins
func SyncPlugins(serverName string) error {
	log.Info("Checking for required plugins...")
	plugins, err := AvailablePlugins(serverName)
	if err != nil {
		return err
	}

	installed := false

	errList := make([]error, 0)
	for idx := range plugins {
		if plugins[idx].Status != common.PluginStatusInstalled {
			installed = true
			err = InstallPlugin(serverName, plugins[idx].Name, plugins[idx].RecommendedVersion)
			if err != nil {
				errList = append(errList, err)
			}
		}
	}
	err = kerrors.NewAggregate(errList)
	if err != nil {
		return err
	}

	if !installed {
		log.Info("All required plugins are already installed and up-to-date")
	} else {
		log.Info("Successfully installed all required plugins")
	}
	return nil
}

// InstallPluginsFromLocalSource installs plugin from local source directory
func InstallPluginsFromLocalSource(pluginName, version, localPath string) error {
	// Set default local plugin distro to localpath as while installing the plugin
	// from local source we should take t
	common.DefaultLocalPluginDistroDir = localPath

	plugins, err := DiscoverPluginsFromLocalSource(localPath)
	if err != nil {
		return errors.Wrap(err, "unable to discover plugins")
	}

	found := false

	errList := make([]error, 0)
	for idx := range plugins {
		if pluginName == cli.AllPlugins || pluginName == plugins[idx].Name {
			found = true
			err := installOrUpgradePlugin("", &plugins[idx], plugins[idx].RecommendedVersion)
			if err != nil {
				errList = append(errList, err)
			}
		}
	}
	err = kerrors.NewAggregate(errList)
	if err != nil {
		return err
	}
	if !found {
		return errors.Errorf("unable to find plugin '%v'", pluginName)
	}
	return nil
}

// DiscoverPluginsFromLocalSource returns the available plugins that are discovered from the provided local path
func DiscoverPluginsFromLocalSource(localPath string) ([]plugin.Discovered, error) {
	if localPath == "" {
		return nil, nil
	}

	plugins, err := discoverPluginsFromLocalSource(localPath)
	// If no error then return the discovered plugins
	if err == nil {
		return plugins, nil
	}

	// Check if the manifest.yaml file exists to see if the directory is legacy structure or not
	if _, err2 := os.Stat(filepath.Join(localPath, ManifestFileName)); errors.Is(err2, os.ErrNotExist) {
		return nil, err
	}

	// As manifest.yaml file exists it assumes in this case the directory is in
	// the legacy structure, and attempt to process it as such
	return discoverPluginsFromLocalSourceWithLegacyDirectoryStructure(localPath)
}

func discoverPluginsFromLocalSource(localPath string) ([]plugin.Discovered, error) {
	// Set default local plugin distro to localpath while installing the plugin
	// from local source. This is done to allow CLI to know the basepath incase the
	// relative path is provided as part of CLIPlugin definition for local discovery
	common.DefaultLocalPluginDistroDir = localPath

	var pds []v1alpha1.PluginDiscovery

	items, err := os.ReadDir(filepath.Join(localPath, "discovery"))
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading local plugin manifest directory")
	}
	for _, item := range items {
		if item.IsDir() {
			pd := v1alpha1.PluginDiscovery{
				Local: &v1alpha1.LocalDiscovery{
					Name: "",
					Path: filepath.Join(localPath, "discovery", item.Name()),
				},
			}
			pds = append(pds, pd)
		}
	}

	plugins, err := discoverPlugins(pds)
	if err != nil {
		return nil, err
	}

	for i := range plugins {
		plugins[i].Scope = common.PluginScopeStandalone
		plugins[i].Status = common.PluginStatusNotInstalled
		plugins[i].DiscoveryType = common.DiscoveryTypeLocal
	}
	return plugins, nil
}

// Clean deletes all plugins and tests.
func Clean() error {
	if err := catalog.CleanCatalogCache(); err != nil {
		return errors.Errorf("Failed to clean the catalog cache %v", err)
	}
	return os.RemoveAll(common.DefaultPluginRoot)
}

// getCLIPluginResourceWithLocalDistroFromPluginDescriptor return cliv1alpha1.CLIPlugin resource from the plugin descriptor
// Note: This function generates cliv1alpha1.CLIPlugin which contains only single local distribution type artifact for
// OS-ARCH where user is running the cli
// This function is only used to create CLIPlugin resource for local plugin installation with legacy directory structure
func getCLIPluginResourceWithLocalDistroFromPluginDescriptor(pd *cliv1alpha1.PluginDescriptor, pluginBinaryPath string) cliv1alpha1.CLIPlugin {
	return cliv1alpha1.CLIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: pd.Name,
		},
		Spec: cliv1alpha1.CLIPluginSpec{
			Description:        pd.Description,
			RecommendedVersion: pd.Version,
			Artifacts: map[string]cliv1alpha1.ArtifactList{
				pd.Version: []cliv1alpha1.Artifact{
					{
						URI:  pluginBinaryPath,
						Type: common.DistributionTypeLocal,
						OS:   runtime.GOOS,
						Arch: runtime.GOARCH,
					},
				},
			},
		},
	}
}

// discoverPluginsFromLocalSourceWithLegacyDirectoryStructure returns the available plugins
// that are discovered from the provided local path
func discoverPluginsFromLocalSourceWithLegacyDirectoryStructure(localPath string) ([]plugin.Discovered, error) {
	if localPath == "" {
		return nil, nil
	}

	// Get the plugin manifest object from manifest.yaml file
	manifest, err := getPluginManifestResource(filepath.Join(localPath, ManifestFileName))
	if err != nil {
		return nil, err
	}

	var discoveredPlugins []plugin.Discovered

	// Create plugin.Discovered object for all locally available plugin
	for _, p := range manifest.Plugins {
		if p.Name == cli.CoreName {
			continue
		}

		// Get the plugin descriptor from the plugin.yaml file
		pd, err := getPluginDescriptorResource(filepath.Join(localPath, p.Name, PluginFileName))
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal plugin.yaml: %v", err)
		}

		absLocalPath, err := filepath.Abs(localPath)
		if err != nil {
			return nil, err
		}
		// With legacy configuration directory structure creating the pluginBinary path from plugin descriptor
		// Sample path: cli/<plugin-name>/<plugin-version>/tanzu-<plugin-name>-<os>_<arch>
		// 				cli/login/v0.14.0/tanzu-login-darwin_amd64
		// As mentioned above, we expect the binary for user's OS-ARCH is present and hence creating path accordingly
		pluginBinaryPath := filepath.Join(absLocalPath, p.Name, pd.Version, fmt.Sprintf("tanzu-%s-%s_%s", p.Name, runtime.GOOS, runtime.GOARCH))
		if common.BuildArch().IsWindows() {
			pluginBinaryPath += exe
		}
		// Check if the pluginBinary file exists or not
		if _, err := os.Stat(pluginBinaryPath); errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrapf(err, "unable to find plugin binary for %q", p.Name)
		}

		p := getCLIPluginResourceWithLocalDistroFromPluginDescriptor(pd, pluginBinaryPath)

		// Create plugin.Discovered resource from CLIPlugin resource
		dp, err := discovery.DiscoveredFromK8sV1alpha1(&p)
		if err != nil {
			return nil, err
		}
		dp.DiscoveryType = common.DiscoveryTypeLocal
		dp.Scope = common.PluginScopeStandalone
		dp.Status = common.PluginStatusNotInstalled

		discoveredPlugins = append(discoveredPlugins, dp)
	}

	return discoveredPlugins, nil
}

// getPluginManifestResource returns cli.Manifest resource by reading manifest file
func getPluginManifestResource(manifestFilePath string) (*cli.Manifest, error) {
	b, err := os.ReadFile(manifestFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not find %s file: %v", filepath.Base(manifestFilePath), err)
	}

	var manifest cli.Manifest
	err = yaml.Unmarshal(b, &manifest)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal %s: %v", filepath.Base(manifestFilePath), err)
	}
	return &manifest, nil
}

// getPluginDescriptorResource returns cliv1alpha1.PluginDescriptor resource by reading plugin file
func getPluginDescriptorResource(pluginFilePath string) (*cliv1alpha1.PluginDescriptor, error) {
	b, err := os.ReadFile(pluginFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not find %s file: %v", filepath.Base(pluginFilePath), err)
	}

	var pd cliv1alpha1.PluginDescriptor
	err = yaml.Unmarshal(b, &pd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal %s: %v", filepath.Base(pluginFilePath), err)
	}
	return &pd, nil
}

// verifyPluginPreDownload verifies that the plugin distribution repo is trusted
// and returns error if the verification fails.
func verifyPluginPreDownload(p *plugin.Discovered) error {
	artifactInfo, err := p.Distribution.DescribeArtifact(p.RecommendedVersion, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if artifactInfo.Image != "" {
		return verifyRegistry(artifactInfo.Image)
	}
	if artifactInfo.URI != "" {
		return verifyArtifactLocation(artifactInfo.URI)
	}
	return errors.Errorf("no download information available for artifact \"%s:%s:%s:%s\"", p.Name, p.RecommendedVersion, runtime.GOOS, runtime.GOARCH)
}

// verifyRegistry verifies the authenticity of the registry from where cli is
// trying to download the plugins by comparing it with the list of trusted registries
func verifyRegistry(image string) error {
	trustedRegistries := config.GetTrustedRegistries()
	for _, tr := range trustedRegistries {
		// Verify fullname of the registry has trusted registry fullname as the prefix
		if tr != "" && strings.HasPrefix(image, tr) {
			return nil
		}
	}
	return errors.Errorf("untrusted registry detected with image %q. Allowed registries are %v", image, trustedRegistries)
}

// verifyArtifactLocation verifies the artifact location from where the cli is
// trying to download the plugins by comparing it with the list of trusted locations
func verifyArtifactLocation(uri string) error {
	art, err := artifact.NewURIArtifact(uri)
	if err != nil {
		return err
	}

	switch art.(type) {
	case *artifact.LocalArtifact:
		// trust local artifacts implicitly
		return nil

	default:
		trustedLocations := config.GetTrustedArtifactLocations()
		for _, tl := range trustedLocations {
			// Verify that the URI has a trusted location as the prefix
			if tl != "" && strings.HasPrefix(uri, tl) {
				return nil
			}
		}
		return errors.Errorf("untrusted artifact location detected with URI %q. Allowed locations are %v", uri, trustedLocations)
	}
}

// verifyPluginPostDownload compares the source digest of the plugin against the
// SHA256 hash of the downloaded binary to ensure that the binary was not altered
// during transit.
func verifyPluginPostDownload(p *plugin.Discovered, srcDigest string, b []byte) error {
	if srcDigest == "" {
		// Skip if the Distribution repo does not have the source digest.
		return nil
	}

	d := sha256.Sum256(b)
	actDigest := fmt.Sprintf("%x", d)
	if actDigest != srcDigest {
		return errors.Errorf("plugin %q has been corrupted during download. source digest: %s, actual digest: %s", p.Name, srcDigest, actDigest)
	}

	return nil
}
