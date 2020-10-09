package cli

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func newTestRepo(t *testing.T, name string) Repository {
	path, err := filepath.Abs(filepath.Join(basepath, "../../../test/cli/mock/", name))
	require.NoError(t, err)
	return NewLocalRepository(name, path)
}

func newTestCatalog(t *testing.T) *Catalog {
	pluginRoot := filepath.Join(xdg.DataHome, "tanzu-cli-test")
	c, err := NewCatalog(WithPluginRoot(pluginRoot), WithDistro([]string{"foo", "bar", "baz"}))
	require.NoError(t, err)
	return c
}

func TestCatalog(t *testing.T) {
	catalog := newTestCatalog(t)
	repo := newTestRepo(t, "artifacts-new")

	err := catalog.InstallAll(repo)
	require.NoError(t, err)

	plugins, err := catalog.List()
	require.NoError(t, err)

	require.Len(t, plugins, 3)

	err = catalog.Delete("foo")
	require.NoError(t, err)

	plugins, err = catalog.List()
	require.NoError(t, err)

	require.Len(t, plugins, 2)

	_, err = catalog.Describe("bar")
	require.NoError(t, err)

	_, err = catalog.Describe("foo")
	require.Error(t, err)

	altRepo := newTestRepo(t, "artifacts-alt")

	multi := NewMultiRepo(repo, altRepo)

	err = catalog.InstallAllMulti(multi)
	require.NoError(t, err)
}
