// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

const (
	// catalogCacheDirName is the name of the local directory in which tanzu state is stored.
	catalogCacheDirName = ".cache/tanzu"
	// catalogCacheFileName is the name of the file which holds Catalog cache
	catalogCacheFileName = "catalog.yaml"
	// exe is an executable file extension
	exe = ".exe"
)

var (
	minConcurrent = 2
	// PluginRoot is the plugin root where plugins are installed
	pluginRoot = DefaultPluginRoot
	// Distro is set of plugins that should be included with the CLI.
	distro = DefaultDistro
)

// getCatalogCacheDir returns the local directory in which tanzu state is stored.
func getCatalogCacheDir() (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate user home directory")
	}
	path = filepath.Join(home, catalogCacheDirName)
	return
}

// HasPluginUpdateIn checks if the plugin has an update in any of the given repositories.
func HasPluginUpdateIn(repos *MultiRepo, p *cliapi.PluginDescriptor) (update bool, repo Repository, version string, err error) {
	for _, repo := range repos.repositories {
		versionSelector := repo.VersionSelector()
		update, version, err := HasPluginUpdate(repo, versionSelector, p)
		if err != nil {
			log.Debugf("could not check for update for plugin %q in repo %q: %v", p.Name, repo.Name, err)
			continue
		}
		if update {
			return update, repo, version, err
		}
	}
	return false, nil, "", nil
}

// HasPluginUpdate tells whether the plugin descriptor has an update available in the given repository.
func HasPluginUpdate(repo Repository, versionSelector VersionSelector, p *cliapi.PluginDescriptor) (update bool, version string, err error) {
	if versionSelector == nil {
		versionSelector = repo.VersionSelector()
	}
	plugin, err := repo.Describe(p.Name)
	if err != nil {
		return update, version, err
	}
	valid := semver.IsValid(p.Version)
	if !valid {
		err = fmt.Errorf("local plugin version %q is not a valid semantic version", p.Version)
		return
	}
	latest := plugin.FindVersion(versionSelector)
	valid = semver.IsValid(latest)
	if !valid {
		err = fmt.Errorf("remote plugin version %q is not a valid semantic version", latest)
		return
	}
	compared := semver.Compare(latest, p.Version)
	if compared == 1 {
		return true, latest, nil
	}
	return false, version, nil
}

// IsDistributionSatisfied tells if a distribution is satisfied by the plugin list.
func IsDistributionSatisfied(desc []*cliapi.PluginDescriptor) bool {
	for _, dist := range distro {
		var contains bool
		for _, plugin := range desc {
			if dist == plugin.Name {
				contains = true
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

// NewCatalog creates an instance of Catalog.
func NewCatalog() (*cliapi.Catalog, error) {
	c := &cliapi.Catalog{}

	err := ensureRoot()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// getCatalogCache retrieves the catalog from from the local directory.
func getCatalogCache() (catalog *cliapi.Catalog, err error) {
	catalogCachePath, err := getCatalogCachePath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(catalogCachePath)
	if err != nil {
		catalog, err = NewCatalog()
		if err != nil {
			return nil, err
		}
		return catalog, nil
	}
	scheme, err := cliapi.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
		apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	var c cliapi.Catalog
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode catalog file")
	}
	return &c, nil
}

// saveCatalogCache saves the catalog in the local directory.
func saveCatalogCache(catalog *cliapi.Catalog) error {
	catalogCachePath, err := getCatalogCachePath()
	if err != nil {
		return err
	}
	_, err = os.Stat(catalogCachePath)
	if os.IsNotExist(err) {
		catalogCacheDir, err := getCatalogCacheDir()
		if err != nil {
			return errors.Wrap(err, "could not find tanzu cache dir for OS")
		}
		err = os.MkdirAll(catalogCacheDir, 0755)
		if err != nil {
			return errors.Wrap(err, "could not make tanzu cache directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not create catalog cache path")
	}

	scheme, err := cliapi.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
		apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	catalog.GetObjectKind().SetGroupVersionKind(cliapi.GroupVersionKindCatalog)
	buf := new(bytes.Buffer)
	if err := s.Encode(catalog, buf); err != nil {
		return errors.Wrap(err, "failed to encode catalog cache file")
	}
	if err = os.WriteFile(catalogCachePath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write catalog cache file")
	}
	return nil
}

// ListPlugins returns the available plugins.
// This code will be removed as part of TKG-13912, and replaced with pluginmanager.AvailablePluginsFromLocalSource or pluginmanager.AvailablePlugins instead
func ListPlugins(exclude ...string) (list []*cliapi.PluginDescriptor, err error) {
	pluginDescriptors, err := getPluginsFromCatalogCache()
	if err != nil {
		log.Debugf("could not get plugin descriptors %v", err)
	} else {
		return pluginDescriptors, nil
	}

	infos, err := os.ReadDir(pluginRoot)
	if err != nil {
		log.Debug("no plugins currently found")
		return list, nil
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		if inExclude(PluginNameFromBin(info.Name()), exclude) {
			continue
		}
		descriptor, err := DescribePlugin(PluginNameFromBin(info.Name()))
		if err != nil {
			return list, err
		}
		list = append(list, descriptor)
	}

	if err := savePluginsToCatalogCache(list); err != nil {
		log.Debugf("Plugin descriptors could not be saved to cache", err)
	}
	return list, nil
}

// savePluginsToCatalogCache saves plugins to catalog cache
func savePluginsToCatalogCache(list []*cliapi.PluginDescriptor) error {
	catalog, err := getCatalogCache()
	if err != nil {
		catalog, err = NewCatalog()
		if err != nil {
			return err
		}
	}
	catalog.PluginDescriptors = list
	if err := saveCatalogCache(catalog); err != nil {
		return err
	}
	return nil
}

// getPluginsFromCatalogCache gets plugins from catalog cache
func getPluginsFromCatalogCache() (list []*cliapi.PluginDescriptor, err error) {
	catalog, err := getCatalogCache()
	if err != nil {
		return nil, err
	}
	if len(catalog.PluginDescriptors) == 0 {
		return nil, errors.New("could not retrieve plugin descriptors from catalog cache")
	}
	return catalog.PluginDescriptors, nil
}

// upsertPluginCacheEntry inserts or updates a plugin entry in catalog cache
func upsertPluginCacheEntry(name string) error {
	list, err := getPluginsFromCatalogCache()
	if err != nil {
		return err
	}
	list = remove(list, name)
	descriptor, err := DescribePlugin(PluginNameFromBin(name))
	if err != nil {
		return err
	}
	list = append(list, descriptor)
	if err := savePluginsToCatalogCache(list); err != nil {
		return err
	}
	return nil
}

// deletePluginCacheEntry deletes plugin entry in catalog cache
func deletePluginCacheEntry(name string) error {
	list, err := getPluginsFromCatalogCache()
	if err != nil {
		return err
	}
	list = remove(list, name)
	if err := savePluginsToCatalogCache(list); err != nil {
		return err
	}
	return nil
}

// CleanCatalogCache cleans the catalog cache
func CleanCatalogCache() error {
	catalogCachePath, err := getCatalogCachePath()
	if err != nil {
		return err
	}
	if err := os.Remove(catalogCachePath); err != nil {
		return err
	}
	return nil
}

// getCatalogCachePath gets the catalog cache path
func getCatalogCachePath() (string, error) {
	catalogCacheDir, err := getCatalogCacheDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate catalog cache directory")
	}
	return filepath.Join(catalogCacheDir, catalogCacheFileName), nil
}

func remove(list []*cliapi.PluginDescriptor, name string) []*cliapi.PluginDescriptor {
	i := 0
	for _, v := range list {
		if v != nil && v.Name != name {
			list[i] = v
			i++
		}
	}
	list = list[:i]
	return list
}

func inExclude(name string, exclude []string) bool {
	for _, e := range exclude {
		if name == e {
			return true
		}
	}
	return false
}

// ListTestPlugins returns the available test plugins.
func ListTestPlugins() (list []*cliapi.PluginDescriptor, err error) {
	infos, err := os.ReadDir(testPath())
	if err != nil {
		log.Debug("no plugins currently found")
		return list, nil
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		descriptor, err := DescribeTestPlugin(PluginNameFromTestBin(info.Name()))
		if err != nil {
			return list, err
		}
		list = append(list, descriptor)
	}

	return list, nil
}

// DescribePlugin describes a plugin.
func DescribePlugin(name string) (desc *cliapi.PluginDescriptor, err error) {
	pluginPath := pluginPath(name)

	b, err := exec.Command(pluginPath, "info").Output()
	if err != nil {
		err = fmt.Errorf("could not describe plugin %q", name)
		return
	}

	var descriptor cliapi.PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		err = fmt.Errorf("could not unmarshal plugin %q description", name)
	}
	return &descriptor, err
}

// DescribeTestPlugin describes a test plugin.
func DescribeTestPlugin(pluginName string) (desc *cliapi.PluginDescriptor, err error) {
	pluginPath := testPluginPath(pluginName)
	b, err := exec.Command(pluginPath, "info").Output()
	if err != nil {
		err = fmt.Errorf("could not describe test plugin %q", pluginName)
		return
	}

	var descriptor cliapi.PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		err = fmt.Errorf("could not unmarshal plugin %q description", pluginName)
	}
	return &descriptor, err
}

// InitializePlugin initializes the plugin configuration
func InitializePlugin(name string) error {
	pluginPath := pluginPath(name)

	b, err := exec.Command(pluginPath, "post-install").CombinedOutput()

	// Note: If user is installing old version of plugin than it is possible that
	// the plugin does not implement post-install command. Ignoring the
	// errors if the command does not exist for a particular plugin.
	if err != nil && !strings.Contains(string(b), "unknown command") {
		log.Warningf("Warning: Failed to initialize plugin '%q' after installation. %v", name, string(b))
	}

	return nil
}

// InstallPlugin installs a plugin from the given repository.
func InstallPlugin(name, version string, repo Repository) error {
	return installOrUpgradePlugin(name, version, repo)
}

// UpgradePlugin upgrades a plugin from the given repository.
func UpgradePlugin(name, version string, repo Repository) error {
	return installOrUpgradePlugin(name, version, repo)
}

func installOrUpgradePlugin(name, version string, repo Repository) error {
	if name == CoreName {
		return fmt.Errorf("cannot install core as a plugin")
	}
	b, err := repo.Fetch(name, version, BuildArch())
	if err != nil {
		return err
	}

	pluginPath := pluginPath(name)

	if BuildArch().IsWindows() {
		pluginPath += exe
	}

	err = os.WriteFile(pluginPath, b, 0755)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	err = upsertPluginCacheEntry(name)
	if err != nil {
		log.Debug("Plugin descriptor could not be updated in cache")
	}
	err = InitializePlugin(name)
	if err != nil {
		log.Infof("could not initialize plugin after installing: %v", err.Error())
	}
	return nil
}

// InstallAllPlugins plugins with the given version finder.
func InstallAllPlugins(repo Repository) error {
	versionSelector := repo.VersionSelector()
	plugins, err := repo.List()
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		// TODO (pbarker): there is likely a better way of doing this
		if plugin.Name == CoreName {
			continue
		}
		err := InstallPlugin(plugin.Name, plugin.FindVersion(versionSelector), repo)
		if err != nil {
			return err
		}
	}
	return nil
}

// InstallAllMulti installs all the plugins at the latest version in all the given repositories.
func InstallAllMulti(repos *MultiRepo) error {
	pluginMap, err := repos.ListPlugins()
	if err != nil {
		return err
	}
	for repoName, descs := range pluginMap {
		repo, err := repos.GetRepository(repoName)
		if err != nil {
			return err
		}
		versionSelector := repo.VersionSelector()
		for _, plugin := range descs {
			if plugin.Name == CoreName {
				continue
			}
			err := InstallPlugin(plugin.Name, plugin.FindVersion(versionSelector), repo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeletePlugin deletes a plugin.
func DeletePlugin(name string) error {
	err := deletePluginCacheEntry(name)
	if err != nil {
		log.Debugf("Plugin descriptor could not be deleted from cache %v", err)
	}
	return os.Remove(pluginPath(name))
}

// Clean deletes all plugins and tests.
func Clean() error {
	if err := CleanCatalogCache(); err != nil {
		log.Debugf("Failed to clean the catalog cache %v", err)
	}
	return os.RemoveAll(pluginRoot)
}

// EnsureDistro ensures that all the distro plugins are installed.
func EnsureDistro(repos *MultiRepo) error {
	fatalErrors := make(chan error, len(distro))

	// Limit the number of concurrent operations we perform so we don't
	// overwhelm the system.
	maxConcurrent := runtime.NumCPU() / 2
	if maxConcurrent < minConcurrent {
		maxConcurrent = 2
	}
	guard := make(chan struct{}, maxConcurrent)

	// capture list of already installed plugins
	installedPlugins, err := ListPlugins()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, pluginName := range distro {
		// if plugin exists on user's system, do not (re)install
		if isPluginInstalled(installedPlugins, pluginName) {
			continue
		}
		wg.Add(1)
		guard <- struct{}{}
		go func(pluginName string) {
			repo, err := repos.Find(pluginName)
			if err != nil {
				fatalErrors <- err
			} else {
				err = InstallPlugin(pluginName, VersionLatest, repo)
				if err != nil {
					fatalErrors <- err
				}
				log.Debugf("done installing: %s", pluginName)
			}
			wg.Done()
			<-guard
		}(pluginName)
	}

	wg.Wait()

	select {
	case err := <-fatalErrors:
		close(fatalErrors)
		return err
	default:
		break
	}
	return nil
}

// InstallTest installs the test for the given plugin name
func InstallTest(pluginName, version string, repo Repository) error {
	b, err := repo.FetchTest(pluginName, version, BuildArch())
	if err != nil {
		return err
	}

	pluginPath := testPluginPath(pluginName)

	if BuildArch().IsWindows() {
		pluginPath += exe
	}

	err = os.WriteFile(pluginPath, b, 0755)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	return nil
}

// EnsureTest ensures the right version of the test is present for the plugin.
func EnsureTest(plugin *cliapi.PluginDescriptor, repos *MultiRepo) error {
	testDesc, err := DescribeTestPlugin(plugin.Name)
	if err == nil {
		if testDesc.BuildSHA == plugin.BuildSHA {
			return nil
		}
	}
	repo, err := repos.Find(plugin.Name)
	if err != nil {
		return err
	}
	log.Infof("installing test for plugin %q", plugin.Name)
	err = InstallTest(plugin.Name, plugin.Version, repo)
	if err != nil {
		log.Debugf("could not install test for plugin %q", plugin.Name)
	}
	return nil
}

// EnsureTests ensures the plugin tests are installed.
func EnsureTests(repos *MultiRepo, exclude ...string) error {
	descs, err := ListPlugins(exclude...)
	if err != nil {
		return err
	}
	for _, desc := range descs {
		err = EnsureTest(desc, repos)
		if err != nil {
			return err
		}
	}
	return nil
}

// Returns the local path for a plugin.
func pluginPath(name string) string {
	binName := BinFromPluginName(name)
	return filepath.Join(pluginRoot, binName)
}

// Returns the local path for a plugin.
func testPluginPath(name string) string {
	binName := BinTestFromPluginName(name)
	return filepath.Join(pluginRoot, "test", binName)
}

// Returns the test path relative to the plugin root
func testPath() string {
	return filepath.Join(pluginRoot, "test")
}

// Ensure the root directory exists.
func ensureRoot() error {
	_, err := os.Stat(testPath())
	if os.IsNotExist(err) {
		err := os.MkdirAll(testPath(), 0755)
		return errors.Wrap(err, "could not make root plugin directory")
	}
	return err
}

// isPluginInstalled takes a list of PluginDescriptors representing installed plugins.
// When the pluginName entered matches a plugin in the descriptor list, true is returned
// A list of installed plugins can be captured by calling Catalog's List method.
func isPluginInstalled(installedPlugin []*cliapi.PluginDescriptor, pluginName string) bool {
	for _, p := range installedPlugin {
		if p.Name == pluginName {
			return true
		}
	}
	return false
}
