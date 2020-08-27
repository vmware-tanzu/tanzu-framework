package repo

import (
	"fmt"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

// MockRepository is a mock repository
type MockRepository struct {
	plugins     []cli.PluginDescriptor
	coreVersion string
}

// MockRepoOpts are options for the mock repository.
type MockRepoOpts func(*MockRepository)

// NewMockRepository is a new mock repository.
func NewMockRepository(opts ...MockRepoOpts) *MockRepository {
	m := &MockRepository{plugins: defaultMockPlugins}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithMockPlugins adds mock plugins to the mock repo.
func WithMockPlugins(plugins []cli.PluginDescriptor) func(*MockRepository) {
	return func(m *MockRepository) {
		m.plugins = plugins
	}
}

// WithCoreVersion sets the core version for the mock repo.
func WithCoreVersion(coreVersion string) func(*MockRepository) {
	return func(m *MockRepository) {
		m.coreVersion = coreVersion
	}
}

var defaultMockPlugins = []cli.PluginDescriptor{
	cli.PluginDescriptor{
		Name:        "foo",
		Description: "A foo plugin",
		Version:     "1.0",
	},
	cli.PluginDescriptor{
		Name:        "bar",
		Description: "A bar plugin",
		Version:     "2.0",
	},
	cli.PluginDescriptor{
		Name:        "baz",
		Description: "A baz plugin",
		Version:     "3.0",
	},
}

// List available plugins.
func (m *MockRepository) List() ([]cli.PluginDescriptor, error) {
	return m.plugins, nil
}

// Describe a plugin.
func (m *MockRepository) Describe(name string) (desc cli.PluginDescriptor, err error) {
	for _, plugin := range m.plugins {
		if plugin.Name == name {
			return plugin, nil
		}
	}
	return desc, fmt.Errorf("plugin %q not found", name)
}

// Fetch an artifact.
func (m *MockRepository) Fetch(name, version string, arch cli.Arch) ([]byte, error) {
	return []byte(fmt.Sprintf("name: %v version: %v arch: %v", name, version, arch)), nil
}

// Manifest retrieves the manifest for the repo.
func (m *MockRepository) Manifest() (cli.Manifest, error) {
	return cli.Manifest{
		Plugins: m.plugins,
		Version: m.coreVersion,
	}, nil
}
