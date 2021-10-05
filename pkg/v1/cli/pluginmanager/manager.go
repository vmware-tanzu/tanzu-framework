// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/discovery"
	"go.uber.org/multierr"
	"golang.org/x/mod/semver"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/catalog"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

const (
	// exe is an executable file extension
	exe = ".exe"
)

var (
	minConcurrent = 2
	// PluginRoot is the plugin root where plugins are installed
	pluginRoot = common.DefaultPluginRoot
	// Distro is set of plugins that should be included with the CLI.
	distro = common.DefaultDistro
)

// ValidatePlugin validates the plugin descriptor.
func ValidatePlugin(p *cliv1alpha1.PluginDescriptor) (err error) {
	// skip builder plugin for bootstrapping
	if p.Name == "builder" {
		return nil
	}
	if p.Name == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q name cannot be empty", p.Name))
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
		plugins[i].Scope = "Stand-Alone"
		plugins[i].Status = "not installed"
	}
	return
}

// DiscoverServerPlugins returns the available plugins associated with the given server
func DiscoverServerPlugins(serverName string) (plugins []plugin.Discovered, err error) {
	plugins = []plugin.Discovered{}
	server, e := config.GetServer(serverName)
	if e != nil {
		return
	}

	plugins, err = discoverPlugins(server.DiscoverySources)
	if err != nil {
		return
	}
	for i := range plugins {
		plugins[i].Scope = "Context"
		plugins[i].Status = "not installed"
	}
	return
}

// DiscoverPlugins returns the available plugins that can be used with the given server
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
func AvailablePlugins(serverName string) (availablePlugins []plugin.Discovered, err error) {
	discoveredServerPlugins, discoveredStandalonePlugins, err := DiscoverPlugins(serverName)
	if err != nil {
		return
	}
	installedSeverPluginDesc, installedStandalonePluginDesc, err := InstalledPlugins(serverName)
	if err != nil {
		return
	}

	availablePlugins = discoveredServerPlugins

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
				availablePlugins[j].Status = "installed"
			}
		}
	}

	for i := range installedStandalonePluginDesc {
		for j := range availablePlugins {
			if installedStandalonePluginDesc[i].Name == availablePlugins[j].Name &&
				installedStandalonePluginDesc[i].Discovery == availablePlugins[j].Source {
				// Match found, Check for update available and update status
				availablePlugins[j].Status = "installed"
			}
		}
	}
	return
}

// InstalledPlugins returns the installed plugins.
func InstalledPlugins(serverName string, exclude ...string) (serverPlugins, standalonePlugins []*cliv1alpha1.PluginDescriptor, err error) {
	return catalog.GetPluginsFromCatalogCache(serverName)
}

// DescribePlugin describes a plugin.
func DescribePlugin(serverName, pluginName string) (desc *cliv1alpha1.PluginDescriptor, err error) {
	pluginPath, err := catalog.GetPluginPath(serverName, pluginName)
	if err != nil {
		err = fmt.Errorf("could not get plugin path for plugin %q", pluginName)
	}

	log.Info(pluginPath)

	b, err := exec.Command(pluginPath, "info").Output()
	if err != nil {
		err = errors.Wrapf(err, "could not describe plugin %q", pluginName)
		return
	}

	var descriptor cliv1alpha1.PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		err = fmt.Errorf("could not unmarshal plugin %q description", pluginName)
	}
	return &descriptor, err
}

// InitializePlugin initializes the plugin configuration
func InitializePlugin(serverName, pluginName string) error {
	pluginPath, err := catalog.GetPluginPath(serverName, pluginName)
	if err != nil {
		err = fmt.Errorf("could not get plugin path for plugin %q", pluginName)
	}

	b, err := exec.Command(pluginPath, "post-install").CombinedOutput()

	// Note: If user is installing old version of plugin than it is possible that
	// the plugin does not implement post-install command. Ignoring the
	// errors if the command does not exist for a particular plugin.
	if err != nil && !strings.Contains(string(b), "unknown command") {
		log.Warningf("Warning: Failed to initialize plugin '%q' after installation. %v", pluginName, string(b))
	}

	return nil
}

// InstallPlugin installs a plugin from the given repository.
func InstallPlugin(serverName, pluginName, version string) error {
	availablePlugins, err := AvailablePlugins(serverName)
	if err != nil {
		return err
	}
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			if availablePlugins[i].Scope == "Stand-Alone" {
				serverName = ""
			}
			return installOrUpgradePlugin(serverName, availablePlugins[i], version)
		}
	}

	return errors.Errorf("unable to find plugin '%v'", pluginName)
}

// UpgradePlugin upgrades a plugin from the given repository.
func UpgradePlugin(serverName, pluginName, version string) error {
	availablePlugins, err := AvailablePlugins(serverName)
	if err != nil {
		return err
	}
	for i := range availablePlugins {
		if availablePlugins[i].Name == pluginName {
			if availablePlugins[i].Scope == "Stand-Alone" {
				serverName = ""
			}
			return installOrUpgradePlugin(serverName, availablePlugins[i], version)
		}
	}

	return errors.Errorf("unable to find plugin '%v'", pluginName)
}

func installOrUpgradePlugin(serverName string, p plugin.Discovered, version string) error {
	b, err := p.Distribution.Fetch(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	pluginName := p.Name
	pluginPath := filepath.Join(pluginRoot, pluginName, version)

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

	b, err = exec.Command(pluginPath, "info").Output()
	if err != nil {
		return fmt.Errorf("could not describe plugin %q", pluginName)
	}
	var descriptor cliv1alpha1.PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		err = fmt.Errorf("could not unmarshal plugin %q description", pluginName)
	}
	descriptor.InstallationPath = pluginPath

	err = catalog.UpsertPluginCacheEntry(serverName, pluginName, descriptor)
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
func DeletePlugin(serverName, pluginName string) error {
	pluginPath, err := catalog.GetPluginPath(serverName, pluginName)
	if err != nil {
		err = fmt.Errorf("could not get plugin path for plugin %q", pluginName)
	}

	err = catalog.DeletePluginCacheEntry(serverName, pluginName)
	if err != nil {
		log.Debugf("Plugin descriptor could not be deleted from cache %v", err)
	}

	return os.Remove(pluginPath)
}

// Clean deletes all plugins and tests.
func Clean() error {
	if err := catalog.CleanCatalogCache(); err != nil {
		return errors.Errorf("Failed to clean the catalog cache %v", err)
	}
	return os.RemoveAll(pluginRoot)
}
