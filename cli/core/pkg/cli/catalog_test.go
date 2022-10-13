// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//nolint:deadcode,unused
package cli

// TODO: This is legacy code tests and will be removed soon as part of TKG-13912
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	_, b, _, _     = runtime.Caller(0)
	basepath       = filepath.Dir(b)
	mockPluginList = []string{"foo", "bar", "baz"}
)

func newTestRepo(t *testing.T, name string) Repository {
	path, err := filepath.Abs(filepath.Join(basepath, "../../../test/cli/mock/", name))
	require.NoError(t, err)
	return NewLocalRepository(name, path)
}

func newTestCatalog(t *testing.T) {
	distro = mockPluginList
	pluginRoot = filepath.Join(xdg.DataHome, "tanzu-cli-test")
	os.RemoveAll(pluginRoot)
	_, err := NewCatalog()
	require.NoError(t, err)
}

func setupCatalogCache() error {
	catalogCachePath, err := getCatalogCachePath()
	if err != nil {
		return errors.Wrap(err, "could not get plugin descriptor path")
	}
	_, err = os.Stat(catalogCachePath)
	if os.IsNotExist(err) {
		localDir, err := getCatalogCacheDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		err = os.MkdirAll(localDir, 0755)
		if err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not create plugin descriptors path")
	} else {
		if err := CleanCatalogCache(); err != nil {
			return errors.Wrap(err, "could not clean plugin descriptors cache")
		}
	}
	return nil
}

func testMultiRepo(t *testing.T, multi *MultiRepo) {
	err := InstallAllMulti(multi)
	require.NoError(t, err)

	err = EnsureTests(multi)
	require.NoError(t, err)

	err = EnsureDistro(multi)
	require.NoError(t, err)
}

func testByDownGrading(t *testing.T) int {
	plugins, err := ListPlugins()
	require.NoError(t, err)

	numPluginsDowngraded := 1
	repoOld := newTestRepo(t, "artifacts-old")
	// downgrades from v0.0.4 to v0.0.3
	err = InstallPlugin("baz", "v0.0.3", repoOld)
	require.NoError(t, err)
	pluginsAfterDowngrade, err := ListPlugins()
	require.NoError(t, err)
	require.NotEqual(t, plugins, pluginsAfterDowngrade)
	return numPluginsDowngraded
}

func testHasUpdate(t *testing.T, multi *MultiRepo, numPluginsDowngraded int) {
	plugins, err := ListPlugins()
	require.NoError(t, err)

	numPluginsRequiringUpdate := 0
	for _, p := range plugins {
		hasUpdate, repo, version, err := HasPluginUpdateIn(multi, p)
		require.NoError(t, err)

		if hasUpdate {
			numPluginsRequiringUpdate++

			hasUpdate2, version2, err2 := HasPluginUpdate(repo, nil, p)
			require.NoError(t, err2)

			require.Equal(t, hasUpdate, hasUpdate2)
			require.Equal(t, version, version2)
		}
	}
	require.Equal(t, numPluginsRequiringUpdate, numPluginsDowngraded)
}

// TODO: This is legacy code tests and will be removed soon. This test and implementation file would be removed soon
//func TestCatalog(t *testing.T) {
//	newTestCatalog(t)
//
//	//setup cache
//	err := setupCatalogCache()
//	require.NoError(t, err)
//
//	// clean cache
//	defer func() {
//		err := CleanCatalogCache()
//		require.NoError(t, err)
//	}()
//
//	repo := newTestRepo(t, "artifacts-new")
//
//	err = InstallAllPlugins(repo)
//	require.NoError(t, err)
//
//	err = InstallPlugin("foo", "v0.0.3", repo)
//	require.NoError(t, err)
//
//	err = InstallPlugin("foo", "v0.0.0-missingversion", repo)
//	require.Error(t, err)
//
//	err = UpgradePlugin("foo", "v0.0.4", repo)
//	require.Error(t, err)
//
//	plugins, _ := ListPlugins()
//	require.Len(t, plugins, 3)
//
//	err = InstallPlugin("notpresent", "v0.0.0", repo)
//	require.Error(t, err)
//
//	plugins, err = ListPlugins()
//	require.NoError(t, err)
//	require.Len(t, plugins, 3)
//
//	pluginsInCatalogCache, err := getPluginsFromCatalogCache()
//	require.NoError(t, err)
//	require.Len(t, pluginsInCatalogCache, 3)
//
//	err = DeletePlugin("foo")
//	require.NoError(t, err)
//
//	plugins, err = ListPlugins()
//	require.NoError(t, err)
//
//	require.Len(t, plugins, 2)
//
//	pluginsInCatalogCache, err = getPluginsFromCatalogCache()
//	require.NoError(t, err)
//	require.Len(t, pluginsInCatalogCache, 2)
//
//	_, err = DescribePlugin("bar")
//	require.NoError(t, err)
//
//	altRepo := newTestRepo(t, "artifacts-alt")
//
//	multi := NewMultiRepo(repo, altRepo)
//
//	testMultiRepo(t, multi)
//
//	numPluginsDowngraded := testByDownGrading(t)
//
//	testHasUpdate(t, multi, numPluginsDowngraded)
//
//	err = EnsureDistro(multi)
//	require.NoError(t, err)
//	pluginsAfterReensure, err := ListPlugins()
//	require.NoError(t, err)
//	// ensure does not update/upgrade the plugin to v0.0.4
//	// thus the plugins installed in the catalog and the plugins
//	// on the user's file system do not match
//	require.NotEqual(t, plugins, pluginsAfterReensure)
//
//	invalidPluginList := append(mockPluginList, "notpresent")
//	distro = invalidPluginList
//	err = EnsureDistro(multi)
//	require.Error(t, err)
//
//	// clean test plugin root
//	err = os.RemoveAll(pluginRoot)
//	require.NoError(t, err)
//}
