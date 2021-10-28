// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pluginmanager is resposible for plugin discovery and installation
package pluginmanager

import (
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
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/catalog"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/discovery"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

const (
	// exe is an executable file extension
	exe = ".exe"
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
			return nil, errors.Wrapf(err, "unable to list plugin from discovery '%v'", discObject.Name())
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
func DiscoverServerPlugins(serverName string) (plugins []plugin.Discovered, err error) {
	plugins = []plugin.Discovered{}
	if serverName == "" {
		return
	}

	server, e := config.GetServer(serverName)
	if e != nil {
		return
	}

	plugins, err = discoverPlugins(server.DiscoverySources)
	if err != nil {
		return
	}
	for i := range plugins {
		plugins[i].Scope = common.PluginScopeContext
		plugins[i].Status = common.PluginStatusNotInstalled
	}
	return
}

// DiscoverPlugins returns the available plugins that can be used with the given server
// If serverName is empty(""), return only standalone plugins
func DiscoverPlugins(serverName string) (serverPlugins, standalonePlugins []plugin.Discovered, err error) {
	serverPlugins, err = DiscoverServerPlugins(serverName)
	if err != nil {
		err = errors.Wrapf(err, "unable to discover server plugins")
		return
	}
	standalonePlugins, err = DiscoverStandalonePlugins()
	if err != nil {
		err = errors.Wrapf(err, "unable to discover server plugins")
		return
	}

	// TODO(anuj): Remove duplicate plugins with server plugins getting higher priority
	return
}

// AvailablePlugins returns the list of available plugins including discovered and installed plugins
// If serverName is empty(""), return only available standalone plugins
func AvailablePlugins(serverName string) ([]plugin.Discovered, error) {
	discoveredServerPlugins, discoveredStandalonePlugins, err := DiscoverPlugins(serverName)
	if err != nil {
		return nil, err
	}
	installedSeverPluginDesc, installedStandalonePluginDesc, err := InstalledPlugins(serverName)
	if err != nil {
		return nil, err
	}

	availablePlugins := discoveredServerPlugins

	for i := range discoveredStandalonePlugins {
		exists := false
		for j := range availablePlugins {
			if discoveredStandalonePlugins[i].Name == availablePlugins[j].Name {
				exists = true
				break
			}
		}
		if !exists {
			availablePlugins = append(availablePlugins, discoveredStandalonePlugins[i])
		}
	}

	for i := range installedSeverPluginDesc {
		for j := range availablePlugins {
			if installedSeverPluginDesc[i].Name == availablePlugins[j].Name &&
				installedSeverPluginDesc[i].Discovery == availablePlugins[j].Source {
				// Match found, Check for update available and update status
				availablePlugins[j].Status = common.PluginStatusInstalled
			}
		}
	}

	for i := range installedStandalonePluginDesc {
		for j := range availablePlugins {
			if installedStandalonePluginDesc[i].Name == availablePlugins[j].Name &&
				installedStandalonePluginDesc[i].Discovery == availablePlugins[j].Source {
				// Match found, Check for update available and update status
				availablePlugins[j].Status = common.PluginStatusInstalled
			}
		}
	}
	return availablePlugins, nil
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
	log.Info("Installing plugin", p.Name)

	b, err := p.Distribution.Fetch(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
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

// Clean deletes all plugins and tests.
func Clean() error {
	if err := catalog.CleanCatalogCache(); err != nil {
		return errors.Errorf("Failed to clean the catalog cache %v", err)
	}
	return os.RemoveAll(common.DefaultPluginRoot)
}
