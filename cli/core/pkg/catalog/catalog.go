// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"

	"github.com/pkg/errors"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/utils"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

const (
	// catalogCacheFileName is the name of the file which holds Catalog cache
	catalogCacheFileName = "catalog.yaml"
)

var (
	// PluginRoot is the plugin root where plugins are installed
	pluginRoot = common.DefaultPluginRoot
)

// ContextCatalog denotes a local plugin catalog for a given context or
// stand-alone.
type ContextCatalog struct {
	sharedCatalog *cliapi.Catalog
	plugins       cliapi.PluginAssociation
}

// NewContextCatalog creates context-aware catalog
func NewContextCatalog(context string) (*ContextCatalog, error) {
	sc, err := getCatalogCache()
	if err != nil {
		return nil, err
	}

	var plugins cliapi.PluginAssociation
	if context == "" {
		plugins = sc.StandAlonePlugins
	} else {
		var ok bool
		plugins, ok = sc.ServerPlugins[context]
		if !ok {
			plugins = make(cliapi.PluginAssociation)
			sc.ServerPlugins[context] = plugins
		}
	}

	return &ContextCatalog{
		sharedCatalog: sc,
		plugins:       plugins,
	}, nil
}

// Upsert inserts/updates the given plugin.
func (c *ContextCatalog) Upsert(plugin *cliapi.PluginDescriptor) error {
	pluginNameTarget := PluginNameTarget(plugin.Name, plugin.Target)

	c.plugins[pluginNameTarget] = plugin.InstallationPath
	c.sharedCatalog.IndexByPath[plugin.InstallationPath] = *plugin

	if !utils.ContainsString(c.sharedCatalog.IndexByName[pluginNameTarget], plugin.InstallationPath) {
		c.sharedCatalog.IndexByName[pluginNameTarget] = append(c.sharedCatalog.IndexByName[pluginNameTarget], plugin.InstallationPath)
	}

	return saveCatalogCache(c.sharedCatalog)
}

// Get looks up the descriptor of a plugin given its name.
func (c *ContextCatalog) Get(plugin string) (cliapi.PluginDescriptor, bool) {
	pd := cliapi.PluginDescriptor{}
	path, ok := c.plugins[plugin]
	if !ok {
		return pd, false
	}

	pd, ok = c.sharedCatalog.IndexByPath[path]
	if !ok {
		return pd, false
	}

	return pd, true
}

// List returns the list of active plugins.
// Active plugin means the plugin that are available to the user
// based on the current logged-in server.
func (c *ContextCatalog) List() []cliapi.PluginDescriptor {
	pds := make([]cliapi.PluginDescriptor, 0)
	for _, installationPath := range c.plugins {
		pd := c.sharedCatalog.IndexByPath[installationPath]
		pds = append(pds, pd)
	}
	return pds
}

// Delete deletes the given plugin from the catalog, but it does not delete
// the installation.
func (c *ContextCatalog) Delete(plugin string) error {
	_, ok := c.plugins[plugin]
	if ok {
		delete(c.plugins, plugin)
	}

	return saveCatalogCache(c.sharedCatalog)
}

// getCatalogCacheDir returns the local directory in which tanzu state is stored.
func getCatalogCacheDir() (path string) {
	return common.DefaultCacheDir
}

// newSharedCatalog creates an instance of the shared catalog file.
func newSharedCatalog() (*cliapi.Catalog, error) {
	c := &cliapi.Catalog{
		IndexByPath:       map[string]cliapi.PluginDescriptor{},
		IndexByName:       map[string][]string{},
		StandAlonePlugins: map[string]string{},
		ServerPlugins:     map[string]cliapi.PluginAssociation{},
	}

	err := ensureRoot()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// getCatalogCache retrieves the catalog from from the local directory.
func getCatalogCache() (catalog *cliapi.Catalog, err error) {
	b, err := os.ReadFile(getCatalogCachePath())
	if err != nil {
		catalog, err = newSharedCatalog()
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

	if c.IndexByPath == nil {
		c.IndexByPath = map[string]cliapi.PluginDescriptor{}
	}
	if c.IndexByName == nil {
		c.IndexByName = map[string][]string{}
	}
	if c.StandAlonePlugins == nil {
		c.StandAlonePlugins = map[string]string{}
	}
	if c.ServerPlugins == nil {
		c.ServerPlugins = map[string]cliapi.PluginAssociation{}
	}

	return &c, nil
}

// saveCatalogCache saves the catalog in the local directory.
func saveCatalogCache(catalog *cliapi.Catalog) error {
	catalogCachePath := getCatalogCachePath()
	_, err := os.Stat(catalogCachePath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(getCatalogCacheDir(), 0755)
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

// CleanCatalogCache cleans the catalog cache
func CleanCatalogCache() error {
	if err := os.Remove(getCatalogCachePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// getCatalogCachePath gets the catalog cache path
func getCatalogCachePath() string {
	return filepath.Join(getCatalogCacheDir(), catalogCacheFileName)
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

// Returns the test path relative to the plugin root
func testPath() string {
	return filepath.Join(pluginRoot, "test")
}

// UpdateCatalogCache when updating the core CLI from v0.x.x to v1.x.x. This is
// needed to group the standalone plugins by context type.
func UpdateCatalogCache() error {
	c, err := getCatalogCache()
	if err != nil {
		return err
	}

	return saveCatalogCache(c)
}

func PluginNameTarget(pluginName string, target cliv1alpha1.Target) string {
	if target == "" {
		return pluginName
	}
	return fmt.Sprintf("%s-%s", pluginName, target)
}
