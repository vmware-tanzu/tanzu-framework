// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

const (
	// catalogCacheFileName is the name of the file which holds Catalog cache
	// TODO: Use the original catalog file instead of using v2 once the feature is enabled by default
	catalogCacheFileName = "catalog_v2.yaml"
)

var (
	// PluginRoot is the plugin root where plugins are installed
	pluginRoot = common.DefaultPluginRoot
)

// ContextCatalog denotes a local plugin catalog for a given context or
// stand-alone.
type ContextCatalog struct {
	sharedCatalog *cliv1alpha1.Catalog
	plugins       cliv1alpha1.PluginAssociation
}

// NewContextCatalog creates context-aware catalog
func NewContextCatalog(context string) (*ContextCatalog, error) {
	sc, err := getCatalogCache()
	if err != nil {
		return nil, err
	}

	var plugins cliv1alpha1.PluginAssociation
	if context == "" {
		plugins = sc.StandAlonePlugins
	} else {
		var ok bool
		plugins, ok = sc.ServerPlugins[context]
		if !ok {
			plugins = make(cliv1alpha1.PluginAssociation)
			sc.ServerPlugins[context] = plugins
		}
	}

	return &ContextCatalog{
		sharedCatalog: sc,
		plugins:       plugins,
	}, nil
}

// Upsert inserts/updates the given plugin.
func (c *ContextCatalog) Upsert(plugin *cliv1alpha1.PluginDescriptor) error {
	c.plugins[plugin.Name] = plugin.InstallationPath
	c.sharedCatalog.IndexByPath[plugin.InstallationPath] = *plugin

	if !utils.ContainsString(c.sharedCatalog.IndexByName[plugin.Name], plugin.InstallationPath) {
		c.sharedCatalog.IndexByName[plugin.Name] = append(c.sharedCatalog.IndexByName[plugin.Name], plugin.InstallationPath)
	}

	return saveCatalogCache(c.sharedCatalog)
}

// Get looks up the descriptor of a plugin given its name.
func (c *ContextCatalog) Get(plugin string) (cliv1alpha1.PluginDescriptor, bool) {
	pd := cliv1alpha1.PluginDescriptor{}
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
func (c *ContextCatalog) List() []cliv1alpha1.PluginDescriptor {
	pds := make([]cliv1alpha1.PluginDescriptor, 0)
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
func newSharedCatalog() (*cliv1alpha1.Catalog, error) {
	c := &cliv1alpha1.Catalog{
		IndexByPath:       map[string]cliv1alpha1.PluginDescriptor{},
		IndexByName:       map[string][]string{},
		StandAlonePlugins: map[string]string{},
		ServerPlugins:     map[string]cliv1alpha1.PluginAssociation{},
	}

	err := ensureRoot()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// getCatalogCache retrieves the catalog from from the local directory.
func getCatalogCache() (catalog *cliv1alpha1.Catalog, err error) {
	b, err := os.ReadFile(getCatalogCachePath())
	if err != nil {
		catalog, err = newSharedCatalog()
		if err != nil {
			return nil, err
		}
		return catalog, nil
	}
	scheme, err := cliv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
		apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	var c cliv1alpha1.Catalog
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode catalog file")
	}

	if c.IndexByPath == nil {
		c.IndexByPath = map[string]cliv1alpha1.PluginDescriptor{}
	}
	if c.IndexByName == nil {
		c.IndexByName = map[string][]string{}
	}
	if c.StandAlonePlugins == nil {
		c.StandAlonePlugins = map[string]string{}
	}
	if c.ServerPlugins == nil {
		c.ServerPlugins = map[string]cliv1alpha1.PluginAssociation{}
	}

	return &c, nil
}

// saveCatalogCache saves the catalog in the local directory.
func saveCatalogCache(catalog *cliv1alpha1.Catalog) error {
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

	scheme, err := cliv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
		apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	catalog.GetObjectKind().SetGroupVersionKind(cliv1alpha1.GroupVersionKindCatalog)
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
	if err := os.Remove(getCatalogCachePath()); err != nil {
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
