// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pluginmanager is responsible for plugin discovery and installation
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

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/artifact"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/catalog"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/plugin"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	configlib "github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
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

type DeletePluginOptions struct {
	Target      cliv1alpha1.Target
	PluginName  string
	ForceDelete bool
}

// ValidatePlugin validates the plugin descriptor.
func ValidatePlugin(p *cliapi.PluginDescriptor) (err error) {
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

func discoverPlugins(pd []configapi.PluginDiscovery) ([]plugin.Discovered, error) {
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
	cfg, e := configlib.GetClientConfig()
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
func DiscoverServerPlugins() ([]plugin.Discovered, error) {
	// If the context and target feature is enabled, discover plugins from all currentContexts
	// Else discover plugin based on current Server
	if configlib.IsFeatureActivated(config.FeatureContextCommand) {
		return discoverServerPluginsBasedOnAllCurrentContexts()
	}
	return discoverServerPluginsBasedOnCurrentServer()
}

func discoverServerPluginsBasedOnAllCurrentContexts() ([]plugin.Discovered, error) {
	var plugins []plugin.Discovered

	currentContextMap, err := configlib.GetAllCurrentContextsMap()
	if err != nil {
		return nil, err
	}
	if len(currentContextMap) == 0 {
		return plugins, nil
	}

	for _, context := range currentContextMap {
		var discoverySources []configapi.PluginDiscovery
		discoverySources = append(discoverySources, context.DiscoverySources...)
		discoverySources = append(discoverySources, defaultDiscoverySourceBasedOnContext(context)...)
		discoveredPlugins, err := discoverPlugins(discoverySources)
		if err != nil {
			return discoveredPlugins, err
		}
		for i := range discoveredPlugins {
			discoveredPlugins[i].Scope = common.PluginScopeContext
			discoveredPlugins[i].Status = common.PluginStatusNotInstalled
			discoveredPlugins[i].ServerName = context.Name

			// Associate Target of the plugin based on the Context Type of the Context
			switch context.Type {
			case configapi.CtxTypeTMC:
				discoveredPlugins[i].Target = cliv1alpha1.TargetTMC
			case configapi.CtxTypeK8s:
				discoveredPlugins[i].Target = cliv1alpha1.TargetK8s
			}
		}
		plugins = append(plugins, discoveredPlugins...)
	}
	return plugins, nil
}

// discoverServerPluginsBasedOnCurrentServer returns the available plugins associated with the given server
func discoverServerPluginsBasedOnCurrentServer() ([]plugin.Discovered, error) {
	var plugins []plugin.Discovered

	server, err := configlib.GetCurrentServer()
	if err != nil || server == nil {
		// If servername is not specified than returning empty list
		// as there are no server plugins that can be discovered
		return plugins, nil
	}
	var discoverySources []configapi.PluginDiscovery
	discoverySources = append(discoverySources, server.DiscoverySources...)
	discoverySources = append(discoverySources, defaultDiscoverySourceBasedOnServer(server)...)

	plugins, err = discoverPlugins(discoverySources)
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
func DiscoverPlugins() ([]plugin.Discovered, []plugin.Discovered) {
	serverPlugins, err := DiscoverServerPlugins()
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
func AvailablePlugins() ([]plugin.Discovered, error) {
	discoveredServerPlugins, discoveredStandalonePlugins := DiscoverPlugins()
	return availablePlugins(discoveredServerPlugins, discoveredStandalonePlugins)
}

// AvailablePluginsFromLocalSource returns the list of available plugins from local source
func AvailablePluginsFromLocalSource(localPath string) ([]plugin.Discovered, error) {
	localStandalonePlugins, err := DiscoverPluginsFromLocalSource(localPath)
	if err != nil {
		log.Warningf("Unable to discover standalone plugins from local source, %v", err.Error())
	}
	return availablePlugins([]plugin.Discovered{}, localStandalonePlugins)
}

func availablePlugins(discoveredServerPlugins, discoveredStandalonePlugins []plugin.Discovered) ([]plugin.Discovered, error) {
	installedServerPluginDesc, installedStandalonePluginDesc, err := InstalledPlugins()
	if err != nil {
		return nil, err
	}

	availablePlugins := availablePluginsFromStandaloneAndServerPlugins(discoveredServerPlugins, discoveredStandalonePlugins)

	setAvailablePluginsStatus(availablePlugins, installedServerPluginDesc)
	setAvailablePluginsStatus(availablePlugins, installedStandalonePluginDesc)

	installedButNotDiscoveredPlugins := getInstalledButNotDiscoveredStandalonePlugins(availablePlugins, installedStandalonePluginDesc)
	availablePlugins = append(availablePlugins, installedButNotDiscoveredPlugins...)

	availablePlugins = combineDuplicatePlugins(availablePlugins)

	return availablePlugins, nil
}

// combineDuplicatePlugins combines same plugins to eliminate duplicates
// When there is a plugin name conflicts and target of both the plugins are same, remove one.
//
// Considering, we are adding `k8s` targeted plugins as root level commands as well for backward compatibility
// A plugin 'foo' getting discovered/installed with `<none>` target and a plugin `foo` getting discovered
// with `k8s` discovery (having `k8s` target) should be treated as same plugin.
// This function takes this case into consideration and ignores `<none>` targeted plugin for above the mentioned scenario.
func combineDuplicatePlugins(availablePlugins []plugin.Discovered) []plugin.Discovered {
	mapOfSelectedPlugins := make(map[string]plugin.Discovered)

	combinePluginInstallationStatus := func(plugin1, plugin2 plugin.Discovered) plugin.Discovered {
		// Combine the installation status and installedVersion result when combining plugins
		if plugin2.Status == common.PluginStatusInstalled {
			plugin1.Status = common.PluginStatusInstalled
		}
		if plugin2.InstalledVersion != "" {
			plugin1.InstalledVersion = plugin2.InstalledVersion
		}
		return plugin1
	}

	for i := range availablePlugins {
		if availablePlugins[i].Target == "" {
			key := fmt.Sprintf("%s_%s", availablePlugins[i].Name, cliv1alpha1.TargetK8s)
			dp, exists := mapOfSelectedPlugins[key]
			if !exists {
				mapOfSelectedPlugins[key] = availablePlugins[i]
			} else {
				mapOfSelectedPlugins[key] = combinePluginInstallationStatus(dp, availablePlugins[i])
			}
		} else {
			key := fmt.Sprintf("%s_%s", availablePlugins[i].Name, availablePlugins[i].Target)
			dp, exists := mapOfSelectedPlugins[key]
			if !exists {
				mapOfSelectedPlugins[key] = availablePlugins[i]
			} else if availablePlugins[i].Target == cliv1alpha1.TargetK8s || availablePlugins[i].Scope == common.PluginScopeContext {
				mapOfSelectedPlugins[key] = combinePluginInstallationStatus(availablePlugins[i], dp)
			}
		}
	}

	var selectedPlugins []plugin.Discovered
	for key := range mapOfSelectedPlugins {
		selectedPlugins = append(selectedPlugins, mapOfSelectedPlugins[key])
	}

	return selectedPlugins
}

func getInstalledButNotDiscoveredStandalonePlugins(availablePlugins []plugin.Discovered, installedPluginDesc []cliapi.PluginDescriptor) []plugin.Discovered {
	var newPlugins []plugin.Discovered
	for i := range installedPluginDesc {
		found := false
		for j := range availablePlugins {
			if installedPluginDesc[i].Name == availablePlugins[j].Name && installedPluginDesc[i].Target == availablePlugins[j].Target {
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
			p.InstalledVersion = installedPluginDesc[i].Version
			newPlugins = append(newPlugins, p)
		}
	}
	return newPlugins
}

// DiscoveredFromPluginDescriptor returns discovered plugin object from k8sV1alpha1
func DiscoveredFromPluginDescriptor(p *cliapi.PluginDescriptor) plugin.Discovered {
	dp := plugin.Discovered{
		Name:               p.Name,
		Description:        p.Description,
		RecommendedVersion: p.Version,
		Source:             p.Discovery,
		SupportedVersions:  []string{p.Version},
	}
	return dp
}

func setAvailablePluginsStatus(availablePlugins []plugin.Discovered, installedPluginDesc []cliapi.PluginDescriptor) {
	for i := range installedPluginDesc {
		for j := range availablePlugins {
			if installedPluginDesc[i].Name == availablePlugins[j].Name && installedPluginDesc[i].Target == availablePlugins[j].Target {
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
		matchIndex := pluginIndexForName(availablePlugins, &discoveredStandalonePlugins[i])

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

func pluginIndexForName(availablePlugins []plugin.Discovered, p *plugin.Discovered) int {
	for j := range availablePlugins {
		if p != nil && p.Name == availablePlugins[j].Name && p.Target == availablePlugins[j].Target {
			return j
		}
	}
	return -1 // haven't found a match
}

// InstalledPlugins returns the installed plugins.
// If serverName is empty(""), return only installed standalone plugins
func InstalledPlugins() (serverPlugins, standalonePlugins []cliapi.PluginDescriptor, err error) {
	serverPlugins, err = InstalledServerPlugins()
	if err != nil {
		return nil, nil, err
	}
	standalonePlugins, err = InstalledStandalonePlugins()
	if err != nil {
		return nil, nil, err
	}
	return
}

// InstalledStandalonePlugins returns the installed standalone plugins.
func InstalledStandalonePlugins() ([]cliapi.PluginDescriptor, error) {
	standAloneCatalog, err := catalog.NewContextCatalog("")
	if err != nil {
		return nil, err
	}
	standalonePlugins := standAloneCatalog.List()
	return standalonePlugins, nil
}

// InstalledServerPlugins returns the installed server plugins.
func InstalledServerPlugins() ([]cliapi.PluginDescriptor, error) {
	serverNames, err := configlib.GetAllCurrentContextsList()
	if err != nil {
		return nil, err
	}

	var serverPlugins []cliapi.PluginDescriptor
	for _, serverName := range serverNames {
		if serverName != "" {
			serverCatalog, err := catalog.NewContextCatalog(serverName)
			if err != nil {
				return nil, err
			}
			serverPlugins = append(serverPlugins, serverCatalog.List()...)
		}
	}

	return serverPlugins, nil
}

// DescribePlugin describes a plugin.
func DescribePlugin(pluginName string, target cliv1alpha1.Target) (desc *cliapi.PluginDescriptor, err error) {
	serverPlugins, standalonePlugins, err := InstalledPlugins()
	if err != nil {
		return nil, err
	}
	var matchedPlugins []cliapi.PluginDescriptor

	pds := serverPlugins
	pds = append(pds, standalonePlugins...)
	for i := range pds {
		if pds[i].Name == pluginName {
			matchedPlugins = append(matchedPlugins, pds[i])
		}
	}

	if len(matchedPlugins) == 0 {
		return nil, errors.Errorf("unable to find plugin '%v'", pluginName)
	}
	if len(matchedPlugins) == 1 {
		return &matchedPlugins[0], nil
	}

	for i := range matchedPlugins {
		if matchedPlugins[i].Target == target {
			return &matchedPlugins[i], nil
		}
	}

	return nil, errors.Errorf("unable to uniquely identify plugin '%v'. Please specify Target of the plugin", pluginName)
}

// InitializePlugin initializes the plugin configuration
func InitializePlugin(descriptor *cliapi.PluginDescriptor) error {
	if descriptor == nil {
		return fmt.Errorf("could not get plugin information")
	}

	b, err := execCommand(descriptor.InstallationPath, "post-install").CombinedOutput()

	// Note: If user is installing old version of plugin than it is possible that
	// the plugin does not implement post-install command. Ignoring the
	// errors if the command does not exist for a particular plugin.
	if err != nil && !strings.Contains(string(b), "unknown command") {
		log.Warningf("Warning: Failed to initialize plugin '%q' after installation. %v", descriptor.Name, string(b))
	}

	return nil
}

// InstallPlugin installs a plugin from the given repository.
func InstallPlugin(pluginName, version string, target cliv1alpha1.Target) error {
	availablePlugins, err := AvailablePlugins()
	if err != nil {
		return err
	}

	var matchedPlugins []plugin.Discovered
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			matchedPlugins = append(matchedPlugins, availablePlugins[i])
		}
	}
	if len(matchedPlugins) == 0 {
		return errors.Errorf("unable to find plugin '%v'", pluginName)
	}

	if len(matchedPlugins) == 1 {
		return installOrUpgradePlugin(&matchedPlugins[0], version, false)
	}

	for i := range matchedPlugins {
		if matchedPlugins[i].Target == target {
			return installOrUpgradePlugin(&matchedPlugins[i], version, false)
		}
	}

	return errors.Errorf("unable to uniquely identify plugin '%v'. Please specify Target of the plugin", pluginName)
}

// UpgradePlugin upgrades a plugin from the given repository.
func UpgradePlugin(pluginName, version string, target cliv1alpha1.Target) error {
	return InstallPlugin(pluginName, version, target)
}

// GetRecommendedVersionOfPlugin returns recommended version of the plugin
func GetRecommendedVersionOfPlugin(pluginName string, target cliv1alpha1.Target) (string, error) {
	availablePlugins, err := AvailablePlugins()
	if err != nil {
		return "", err
	}

	var matchedPlugins []plugin.Discovered
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			matchedPlugins = append(matchedPlugins, availablePlugins[i])
		}
	}
	if len(matchedPlugins) == 0 {
		return "", errors.Errorf("unable to find plugin '%v'", pluginName)
	}

	if len(matchedPlugins) == 1 {
		return matchedPlugins[0].RecommendedVersion, nil
	}

	for i := range matchedPlugins {
		if matchedPlugins[i].Target == target {
			return matchedPlugins[i].RecommendedVersion, nil
		}
	}
	return "", errors.Errorf("unable to uniquely identify plugin '%v'. Please specify Target of the plugin", pluginName)
}

func installOrUpgradePlugin(p *plugin.Discovered, version string, installTestPlugin bool) error {
	if p.Target == "" {
		log.Infof("Installing plugin '%v:%v'", p.Name, version)
	} else {
		log.Infof("Installing plugin '%v:%v' with target '%v'", p.Name, version, p.Target)
	}

	binary, err := fetchAndVerifyPlugin(p, version)
	if err != nil {
		return err
	}

	descriptor, err := installAndDescribePlugin(p, version, binary)
	if err != nil {
		return err
	}

	if installTestPlugin {
		if err := doInstallTestPlugin(p, version); err != nil {
			return err
		}
	}

	return updateDescriptorAndInitializePlugin(p, descriptor)
}

func fetchAndVerifyPlugin(p *plugin.Discovered, version string) ([]byte, error) {
	// verify plugin before download
	err := verifyPluginPreDownload(p)
	if err != nil {
		return nil, errors.Wrapf(err, "%q plugin pre-download verification failed", p.Name)
	}

	b, err := p.Distribution.Fetch(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch the plugin metadata for plugin %q", p.Name)
	}

	// verify plugin after download but before installation
	d, err := p.Distribution.GetDigest(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}
	err = verifyPluginPostDownload(p, d, b)
	if err != nil {
		return nil, errors.Wrapf(err, "%q plugin post-download verification failed", p.Name)
	}
	return b, nil
}

func installAndDescribePlugin(p *plugin.Discovered, version string, binary []byte) (*cliapi.PluginDescriptor, error) {
	pluginFileName := fmt.Sprintf("%s_%x_%s", version, sha256.Sum256(binary), p.Target)
	pluginPath := filepath.Join(common.DefaultPluginRoot, p.Name, pluginFileName)

	if err := os.MkdirAll(filepath.Dir(pluginPath), os.ModePerm); err != nil {
		return nil, err
	}

	if common.BuildArch().IsWindows() {
		pluginPath += exe
	}

	if err := os.WriteFile(pluginPath, binary, 0755); err != nil {
		return nil, errors.Wrap(err, "could not write file")
	}
	bytesInfo, err := execCommand(pluginPath, "info").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "could not describe plugin %q", p.Name)
	}

	var descriptor cliapi.PluginDescriptor
	if err = json.Unmarshal(bytesInfo, &descriptor); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal plugin %q description", p.Name)
	}
	descriptor.InstallationPath = pluginPath
	descriptor.Discovery = p.Source
	descriptor.DiscoveredRecommendedVersion = p.RecommendedVersion
	descriptor.Target = p.Target

	return &descriptor, nil
}

func doInstallTestPlugin(p *plugin.Discovered, version string) error {
	log.Infof("Installing test plugin for '%v:%v'", p.Name, version)
	binary, err := p.Distribution.FetchTest(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return errors.Wrapf(err, "unable to install test plugin for '%v:%v'", p.Name, version)
	}
	testpluginPath := filepath.Join(common.DefaultPluginRoot, p.Name, fmt.Sprintf("test-%s", version))
	err = os.WriteFile(testpluginPath, binary, 0755)
	if err != nil {
		return errors.Wrap(err, "error while saving test plugin binary")
	}
	return nil
}

func updateDescriptorAndInitializePlugin(p *plugin.Discovered, descriptor *cliapi.PluginDescriptor) error {
	c, err := catalog.NewContextCatalog(p.ServerName)
	if err != nil {
		return err
	}
	if err := c.Upsert(descriptor); err != nil {
		log.Info("Plugin descriptor could not be updated in cache")
	}
	if err := InitializePlugin(descriptor); err != nil {
		log.Infof("could not initialize plugin after installing: %v", err.Error())
	}
	if err := config.ConfigureDefaultFeatureFlagsIfMissing(descriptor.DefaultFeatureFlags); err != nil {
		log.Infof("could not configure default featureflags for the plugin: %v", err.Error())
	}
	return nil
}

// DeletePlugin deletes a plugin.
func DeletePlugin(options DeletePluginOptions) error {
	serverNames, err := configlib.GetAllCurrentContextsList()
	if err != nil {
		return err
	}

	// Add empty serverName for standalone plugins
	serverNames = append(serverNames, "")
	var catalogContainingPlugin *catalog.ContextCatalog

	for _, serverName := range serverNames {
		c, err := catalog.NewContextCatalog(serverName)
		if err != nil {
			continue
		}
		_, ok := c.Get(catalog.PluginNameTarget(options.PluginName, options.Target))
		if ok {
			catalogContainingPlugin = c
			break
		}
	}

	if catalogContainingPlugin == nil {
		return fmt.Errorf("could not get plugin path for plugin %q", options.PluginName)
	}

	if !options.ForceDelete {
		if err := component.AskForConfirmation(fmt.Sprintf("Deleting Plugin '%s'. Are you sure?", options.PluginName)); err != nil {
			return err
		}
	}
	err = catalogContainingPlugin.Delete(catalog.PluginNameTarget(options.PluginName, options.Target))
	if err != nil {
		return fmt.Errorf("plugin %q could not be deleted from cache", options.PluginName)
	}

	// TODO: delete the plugin binary if it is not used by any server

	return nil
}

// SyncPlugins automatically downloads all available plugins to users machine
func SyncPlugins() error {
	log.Info("Checking for required plugins...")
	plugins, err := AvailablePlugins()
	if err != nil {
		return err
	}

	installed := false

	errList := make([]error, 0)
	for idx := range plugins {
		if plugins[idx].Status != common.PluginStatusInstalled {
			installed = true
			err = InstallPlugin(plugins[idx].Name, plugins[idx].RecommendedVersion, plugins[idx].Target)
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
func InstallPluginsFromLocalSource(pluginName, version string, target cliv1alpha1.Target, localPath string, installTestPlugin bool) error {
	// Set default local plugin distro to local-path as while installing the plugin
	// from local source we should take t
	common.DefaultLocalPluginDistroDir = localPath

	availablePlugins, err := DiscoverPluginsFromLocalSource(localPath)
	if err != nil {
		return errors.Wrap(err, "unable to discover plugins")
	}

	var errList []error

	var matchedPlugins []plugin.Discovered
	for i := range availablePlugins {
		if pluginName == cli.AllPlugins || availablePlugins[i].Name == pluginName {
			matchedPlugins = append(matchedPlugins, availablePlugins[i])
		}
	}
	if len(matchedPlugins) == 0 {
		return errors.Errorf("unable to find plugin '%v'", pluginName)
	}

	if len(matchedPlugins) == 1 {
		return installOrUpgradePlugin(&matchedPlugins[0], FindVersion(matchedPlugins[0].RecommendedVersion, version), installTestPlugin)
	}

	for i := range matchedPlugins {
		// Install all plugins otherwise include all matching plugins
		if pluginName == cli.AllPlugins || matchedPlugins[i].Target == target {
			err = installOrUpgradePlugin(&matchedPlugins[i], FindVersion(matchedPlugins[i].RecommendedVersion, version), installTestPlugin)
			if err != nil {
				errList = append(errList, err)
			}
		}
	}

	err = kerrors.NewAggregate(errList)
	if err != nil {
		return err
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

	var pds []configapi.PluginDiscovery

	items, err := os.ReadDir(filepath.Join(localPath, "discovery"))
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading local plugin manifest directory")
	}
	for _, item := range items {
		if item.IsDir() {
			pd := configapi.PluginDiscovery{
				Local: &configapi.LocalDiscovery{
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
func getCLIPluginResourceWithLocalDistroFromPluginDescriptor(pd *cliapi.PluginDescriptor, pluginBinaryPath string) cliv1alpha1.CLIPlugin {
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

// getPluginDescriptorResource returns cliapi.PluginDescriptor resource by reading plugin file
func getPluginDescriptorResource(pluginFilePath string) (*cliapi.PluginDescriptor, error) {
	b, err := os.ReadFile(pluginFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not find %s file: %v", filepath.Base(pluginFilePath), err)
	}

	var pd cliapi.PluginDescriptor
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

func FindVersion(recommendedPluginVersion, requestedVersion string) string {
	if requestedVersion == "" || requestedVersion == cli.VersionLatest {
		return recommendedPluginVersion
	}
	return requestedVersion
}
