// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

func Test_ContextCatalog_With_Empty_Context(t *testing.T) {
	common.DefaultCacheDir = filepath.Join(os.TempDir(), "test")

	assert := assert.New(t)

	// Create catalog without context
	cc, err := NewContextCatalog("")
	assert.Nil(err)
	assert.NotNil(cc)

	pd1 := cliapi.PluginDescriptor{
		Name:             "fakeplugin1",
		InstallationPath: "/path/to/plugin/fakeplugin1",
		Version:          "1.0.0",
	}

	err = cc.Upsert(&pd1)
	assert.Nil(err)

	pd, exists := cc.Get("fakeplugin1")
	assert.True(exists)
	assert.Equal(pd.Name, "fakeplugin1")
	assert.Equal(pd.InstallationPath, "/path/to/plugin/fakeplugin1")
	assert.Equal(pd.Version, "1.0.0")

	pd2 := cliapi.PluginDescriptor{
		Name:             "fakeplugin2",
		InstallationPath: "/path/to/plugin/fakeplugin2",
		Version:          "2.0.0",
	}
	err = cc.Upsert(&pd2)
	assert.Nil(err)

	pd, exists = cc.Get("fakeplugin2")
	assert.True(exists)
	assert.Equal(pd.Name, "fakeplugin2")
	assert.Equal(pd.InstallationPath, "/path/to/plugin/fakeplugin2")
	assert.Equal(pd.Version, "2.0.0")

	pds := cc.List()
	assert.Equal(len(pds), 2)
	assert.ElementsMatch([]string{pds[0].Name, pds[1].Name}, []string{"fakeplugin1", "fakeplugin2"})
	assert.ElementsMatch([]string{pds[0].InstallationPath, pds[1].InstallationPath}, []string{"/path/to/plugin/fakeplugin1", "/path/to/plugin/fakeplugin2"})
	assert.ElementsMatch([]string{pds[0].Version, pds[1].Version}, []string{"1.0.0", "2.0.0"})

	err = cc.Delete("fakeplugin2")
	assert.Nil(err)

	pd, exists = cc.Get("fakeplugin2")
	assert.False(exists)
	assert.NotEqual(pd.Name, "fakeplugin2")

	pds = cc.List()
	assert.Equal(len(pds), 1)

	// Create another catalog without context
	// The new catalog should also have the same information
	cc2, err := NewContextCatalog("")
	assert.Nil(err)
	assert.NotNil(cc2)

	pd, exists = cc2.Get("fakeplugin1")
	assert.True(exists)
	assert.Equal(pd.Name, "fakeplugin1")
	assert.Equal(pd.InstallationPath, "/path/to/plugin/fakeplugin1")
	assert.Equal(pd.Version, "1.0.0")

	pds = cc2.List()
	assert.Equal(len(pds), 1)

	os.RemoveAll(common.DefaultPluginRoot)
}

func Test_ContextCatalog_With_Context(t *testing.T) {
	common.DefaultCacheDir = filepath.Join(os.TempDir(), "test")

	assert := assert.New(t)

	cc, err := NewContextCatalog("server")
	assert.Nil(err)
	assert.NotNil(cc)

	pd1 := cliapi.PluginDescriptor{
		Name:             "fakeplugin1",
		InstallationPath: "/path/to/plugin/fakeplugin1",
		Version:          "1.0.0",
	}

	err = cc.Upsert(&pd1)
	assert.Nil(err)

	pd, exists := cc.Get("fakeplugin1")
	assert.True(exists)
	assert.Equal(pd.Name, "fakeplugin1")
	assert.Equal(pd.InstallationPath, "/path/to/plugin/fakeplugin1")
	assert.Equal(pd.Version, "1.0.0")

	pd2 := cliapi.PluginDescriptor{
		Name:             "fakeplugin2",
		InstallationPath: "/path/to/plugin/fakeplugin2",
		Version:          "2.0.0",
	}
	err = cc.Upsert(&pd2)
	assert.Nil(err)

	pd, exists = cc.Get("fakeplugin2")
	assert.True(exists)
	assert.Equal(pd.Name, "fakeplugin2")
	assert.Equal(pd.InstallationPath, "/path/to/plugin/fakeplugin2")
	assert.Equal(pd.Version, "2.0.0")

	pds := cc.List()
	assert.Equal(len(pds), 2)
	assert.ElementsMatch([]string{pds[0].Name, pds[1].Name}, []string{"fakeplugin1", "fakeplugin2"})
	assert.ElementsMatch([]string{pds[0].InstallationPath, pds[1].InstallationPath}, []string{"/path/to/plugin/fakeplugin1", "/path/to/plugin/fakeplugin2"})
	assert.ElementsMatch([]string{pds[0].Version, pds[1].Version}, []string{"1.0.0", "2.0.0"})

	err = cc.Delete("fakeplugin2")
	assert.Nil(err)

	pd, exists = cc.Get("fakeplugin2")
	assert.False(exists)
	assert.NotEqual(pd.Name, "fakeplugin2")

	pds = cc.List()
	assert.Equal(len(pds), 1)

	// Create another catalog with same context
	// The new catalog should also have the same information
	cc2, err := NewContextCatalog("server")
	assert.Nil(err)
	assert.NotNil(cc2)

	pd, exists = cc2.Get("fakeplugin1")
	assert.True(exists)
	assert.Equal(pd.Name, "fakeplugin1")
	assert.Equal(pd.InstallationPath, "/path/to/plugin/fakeplugin1")
	assert.Equal(pd.Version, "1.0.0")

	pds = cc2.List()
	assert.Equal(len(pds), 1)

	// Create another catalog with different context
	// The new catalog should not have the same information
	cc3, err := NewContextCatalog("server2")
	assert.Nil(err)
	assert.NotNil(cc3)

	pd, exists = cc3.Get("fakeplugin1")
	assert.False(exists)

	os.RemoveAll(common.DefaultPluginRoot)
}

// Test_CatalogCacheFileName tests we default to catalog.yaml file when
// the featuregate is configured to true by default
func Test_CatalogCacheFileName(t *testing.T) {
	assert := assert.New(t)
	if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
		assert.Equal(catalogCacheFileName, "catalog.yaml")
	}
}
