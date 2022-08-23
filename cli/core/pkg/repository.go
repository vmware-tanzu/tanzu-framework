// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// Repository is a remote repository containing plugin artifacts.
type Repository interface {
	// List available plugins.
	List() ([]Plugin, error)

	// Describe a plugin.
	Describe(name string) (Plugin, error)

	// Fetch an artifact.
	Fetch(name, version string, arch Arch) ([]byte, error)

	// Fetch an artifact test.
	FetchTest(pluginName, version string, arch Arch) ([]byte, error)

	// Name of the repository.
	Name() string

	// Manifest retrieves the manifest for the repo.
	Manifest() (Manifest, error)

	// VersionSelector returns the version finder.
	VersionSelector() VersionSelector
}

// NewDefaultRepository returns the default repository.
func NewDefaultRepository() Repository {
	return NewGCPBucketRepository()
}

// Manifest is stored in the repository which gives an inventory of the artifacts.
type Manifest struct {
	// Created is the time the manifest was created.
	CreatedTime time.Time `json:"created" yaml:"created"`

	// Plugins is a list of plugin artifacts available.
	Plugins []Plugin `json:"plugins" yaml:"plugins"`

	// Deprecated: Version of the root CLI.
	Version string `json:"version" yaml:"version"`

	// CoreVersion of the root CLI.
	CoreVersion string `json:"coreVersion" yaml:"coreVersion"`
}

// GetCoreVersion returns the core version in a backwards compatible manner.
func (m *Manifest) GetCoreVersion() string {
	if m.Version != "" {
		return m.Version
	}
	return m.CoreVersion
}

// Plugin is an installable CLI plugin.
type Plugin struct {
	// Name is the name of the plugin.
	Name string `json:"name" yaml:"name"`

	// Description is the plugin's description.
	Description string `json:"description" yaml:"description"`

	// Versions available for plugin.
	Versions []string `json:"versions" yaml:"versions"`
}

// FindVersion finds the version using the version selector.
func (p *Plugin) FindVersion(selector VersionSelector) string {
	if selector == nil {
		selector = DefaultVersionSelector
	}
	return selector(p.Versions)
}

// Arch represents a system architecture.
type Arch string

// BuildArch returns compile time build arch or locates it.
func BuildArch() Arch {
	return Arch(fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
}

// IsWindows tells if an arch is windows.
func (a Arch) IsWindows() bool {
	if a == Win386 || a == WinAMD64 {
		return true
	}
	return false
}

const (
	// Linux386 arch.
	Linux386 Arch = "linux_386"
	// LinuxAMD64 arch.
	LinuxAMD64 Arch = "linux_amd64"
	// LinuxARM64 arch.
	LinuxARM64 Arch = "linux_arm64"
	// DarwinAMD64 arch.
	DarwinAMD64 Arch = "darwin_amd64"
	// DarwinARM64 arch.
	DarwinARM64 Arch = "darwin_arm64"
	// Win386 arch.
	Win386 Arch = "windows_386"
	// WinAMD64 arch.
	WinAMD64 Arch = "windows_amd64"

	// ManifestFileName is the file name for the manifest.
	ManifestFileName = "manifest.yaml"
	// PluginFileName is the file name for the plugin descriptor.
	PluginFileName = "plugin.yaml"
	// DefaultArtifactsDirectory is the root artifacts directory
	DefaultArtifactsDirectory = "artifacts"

	// AllPlugins is the keyword for all plugins.
	AllPlugins = "all"

	// DefaultManifestQueryTimeout is max time to wait for querying for a plugin manifest
	DefaultManifestQueryTimeout = 5 * time.Second
)

// GCPBucketRepository is a artifact repository utilizing a GCP bucket.
type GCPBucketRepository struct {
	bucketName      string
	rootPath        string
	name            string
	versionSelector VersionSelector
}

// LoadRepositories loads the repositories from the config file along with the known repositories.
func LoadRepositories(c *configv1alpha1.ClientConfig) []Repository {
	repos := []Repository{}
	if c.ClientOptions == nil {
		c.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configv1alpha1.CLIOptions{}
	}

	vs := LoadVersionSelector(c.ClientOptions.CLI.UnstableVersionSelector)
	for _, repo := range c.ClientOptions.CLI.Repositories {
		if repo.GCPPluginRepository == nil {
			continue
		}
		repos = append(repos, loadRepository(repo, vs))
	}
	return repos
}

func loadRepository(repo configv1alpha1.PluginRepository, versionSelector VersionSelector) Repository {
	opts := []Option{
		WithGCPBucket(repo.GCPPluginRepository.BucketName),
		WithName(repo.GCPPluginRepository.Name),
		WithVersionSelector(versionSelector),
	}
	if repo.GCPPluginRepository.RootPath != "" {
		opts = append(opts, WithGCPRootPath(repo.GCPPluginRepository.RootPath))
	}
	return NewGCPBucketRepository(opts...)
}

// LoadVersionSelector will return the correct VersionSelector for a VersionSelectorLevel
func LoadVersionSelector(selectorType configv1alpha1.VersionSelectorLevel) (versionSelector VersionSelector) {
	switch selectorType {
	case configv1alpha1.AllUnstableVersions:
		versionSelector = SelectVersionAny
	case configv1alpha1.AlphaUnstableVersions:
		versionSelector = SelectVersionAlpha
	case configv1alpha1.ExperimentalUnstableVersions:
		versionSelector = SelectVersionExperimental
	case configv1alpha1.NoUnstableVersions:
		versionSelector = DefaultVersionSelector
	default:
		versionSelector = DefaultVersionSelector
	}
	return
}

// NewGCPBucketRepository returns a new GCP bucket repository.
func NewGCPBucketRepository(options ...Option) Repository {
	opts := makeDefaultOptions(options...)

	return &GCPBucketRepository{
		bucketName:      opts.gcpBucket,
		rootPath:        opts.gcpRootPath,
		name:            opts.repoName,
		versionSelector: opts.versionSelector,
	}
}

// List available plugins.
func (g *GCPBucketRepository) List() (plugins []Plugin, err error) {
	manifest, err := g.Manifest()
	if err != nil {
		return plugins, err
	}
	for _, plugin := range manifest.Plugins {
		p, err := g.Describe(plugin.Name)
		if err != nil {
			return plugins, err
		}
		plugins = append(plugins, p)
	}
	return
}

// Describe a plugin.
func (g *GCPBucketRepository) Describe(name string) (plugin Plugin, err error) {
	ctx := context.Background()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return plugin, err
	}

	pluginPath := path.Join(g.rootPath, name, PluginFileName)

	obj := bkt.Object(pluginPath)
	if obj == nil {
		return plugin, fmt.Errorf("artifact %q not found", name)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return plugin, errors.Wrap(err, fmt.Sprintf("could not fetch artifact %q from repository", name))
	}
	defer r.Close()

	d := yaml.NewDecoder(r)

	err = d.Decode(&plugin)
	if err != nil {
		return plugin, errors.Wrap(err, fmt.Sprintf("could not decode plugin %q decriptor", name))
	}

	pluginPath = path.Join(g.rootPath, name)
	query := &storage.Query{Prefix: pluginPath}

	versionMap := map[string]string{}
	it := bkt.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return plugin, errors.Wrap(err, fmt.Sprintf("could not list versions for plugin %q", name))
		}

		// As of today the gcloud Go sdk doesn't allow for an effiecient means of listing by available
		// prefixes or "directories"
		pre := strings.TrimPrefix(attrs.Name, pluginPath)
		split := strings.Split(pre, "/")
		if len(split) < 1 {
			return plugin, fmt.Errorf("could not retrieve version from bucket")
		}
		version := split[1]

		if version != PluginFileName {
			versionMap[version] = ""
		}
	}
	versions := []string{}
	for version := range versionMap {
		versions = append(versions, version)
	}
	plugin.Versions = versions
	return plugin, err
}

// Fetch an artifact.
func (g *GCPBucketRepository) Fetch(name, version string, arch Arch) ([]byte, error) {
	ctx := context.Background()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return nil, err
	}

	if version == VersionLatest {
		plugin, err := g.Describe(name)
		if err != nil {
			return nil, err
		}
		version = plugin.FindVersion(g.versionSelector)
		if version == "" {
			return nil, fmt.Errorf("could not find a suitable version for plugin %q from versions %v", name, plugin.Versions)
		}
	}

	artifactPath := path.Join(g.rootPath, name, version, MakeArtifactName(name, arch))

	return g.fetch(ctx, artifactPath, bkt)
}

// FetchTest fetches a test artifact.
func (g *GCPBucketRepository) FetchTest(name, version string, arch Arch) ([]byte, error) {
	ctx := context.Background()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return nil, err
	}

	if version == VersionLatest {
		plugin, err := g.Describe(name)
		if err != nil {
			return nil, err
		}
		version = plugin.FindVersion(g.versionSelector)
		if version == "" {
			return nil, fmt.Errorf("could not find a suitable version for test plugin %q from versions %v", name, plugin.Versions)
		}
	}

	artifactPath := path.Join(g.rootPath, name, version, "test", MakeTestArtifactName(name, arch))
	return g.fetch(ctx, artifactPath, bkt)
}

func (g *GCPBucketRepository) fetch(ctx context.Context, artifactPath string, bkt *storage.BucketHandle) ([]byte, error) {
	obj := bkt.Object(artifactPath)
	if obj == nil {
		return nil, fmt.Errorf("artifact %q not found", artifactPath)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not read artifact %q", artifactPath))
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch artifact")
	}
	return b, nil
}

// Name of the repository.
func (g *GCPBucketRepository) Name() string {
	return g.name
}

// Manifest retrieves the manifest for a repository.
func (g *GCPBucketRepository) Manifest() (manifest Manifest, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultManifestQueryTimeout)
	defer cancel()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return manifest, err
	}

	manifestPath := path.Join(g.rootPath, ManifestFileName)

	obj := bkt.Object(manifestPath)
	if obj == nil {
		return manifest, fmt.Errorf("could not fetch manifest from repository %q", g.Name())
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return manifest, errors.Wrap(err, fmt.Sprintf("could not fetch manifest from repository %q", g.Name()))
	}
	defer r.Close()

	d := yaml.NewDecoder(r)

	err = d.Decode(&manifest)
	if err != nil {
		return manifest, errors.Wrap(err, "could not decode plugin decriptor")
	}
	return manifest, nil
}

func (g *GCPBucketRepository) getBucket(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to repository")
	}
	bkt := client.Bucket(g.bucketName)
	if bkt == nil {
		return nil, fmt.Errorf("could not connect to repository")
	}
	return bkt, nil
}

// VersionSelector returns the current default version finder.
func (g *GCPBucketRepository) VersionSelector() VersionSelector {
	return g.versionSelector
}

// LocalRepository is a artifact repository utilizing a local host os.
type LocalRepository struct {
	path            string
	name            string
	versionSelector VersionSelector
}

// DefaultLocalRepository is the default local repository.
var DefaultLocalRepository = &LocalRepository{
	path: fmt.Sprintf("./%s", DefaultArtifactsDirectory),
}

// NewLocalRepository returns a new local repository.
func NewLocalRepository(name, localPath string, options ...Option) Repository {
	opts := makeDefaultOptions(options...)
	return &LocalRepository{
		path:            localPath,
		name:            name,
		versionSelector: opts.versionSelector,
	}
}

// List available plugins.
func (l *LocalRepository) List() (plugins []Plugin, err error) {
	manifest, err := l.Manifest()
	if err != nil {
		return plugins, err
	}
	for _, plugin := range manifest.Plugins {
		p, err := l.Describe(plugin.Name)
		if err != nil {
			return plugins, err
		}
		plugins = append(plugins, p)
	}
	return
}

// Describe a plugin.
func (l *LocalRepository) Describe(name string) (plugin Plugin, err error) {
	b, err := os.ReadFile(filepath.Join(l.path, name, PluginFileName))
	if err != nil {
		err = fmt.Errorf("could not find plugin.yaml file for plugin %q: %v", name, err)
		return
	}

	err = yaml.Unmarshal(b, &plugin)
	if err != nil {
		return plugin, fmt.Errorf("could not unmarshal manifest.yaml: %v", err)
	}
	infos, err := os.ReadDir(filepath.Join(l.path, name))
	if err != nil {
		return plugin, err
	}

	versions := []string{}
	for _, info := range infos {
		if info.IsDir() {
			versions = append(versions, info.Name())
		}
	}
	plugin.Versions = versions
	return
}

// Fetch an artifact.
func (l *LocalRepository) Fetch(name, version string, arch Arch) ([]byte, error) {
	if version == "" {
		return nil, fmt.Errorf("version cannot be empty for plugin %q", name)
	}
	if version == VersionLatest {
		plugin, err := l.Describe(name)
		if err != nil {
			return nil, err
		}
		version = plugin.FindVersion(l.versionSelector)
		if version == "" {
			return nil, fmt.Errorf("could not find a suitable version for plugin %q from versions %v", name, plugin.Versions)
		}
	}
	b, err := os.ReadFile(filepath.Join(l.path, name, version, MakeArtifactName(name, arch)))
	if err != nil {
		return nil, errors.Wrap(err, "could not find artifact at given path")
	}
	return b, nil
}

// FetchTest fetches an artifact test.
func (l *LocalRepository) FetchTest(name, version string, arch Arch) ([]byte, error) {
	if version == "" {
		return nil, fmt.Errorf("version cannot be empty for plugin %q", name)
	}
	if version == VersionLatest {
		plugin, err := l.Describe(name)
		if err != nil {
			return nil, err
		}
		version = plugin.FindVersion(l.versionSelector)
		if version == "" {
			return nil, fmt.Errorf("could not find a suitable version for test plugin %q from versions %v", name, plugin.Versions)
		}
	}
	b, err := os.ReadFile(filepath.Join(l.path, name, version, "test", MakeTestArtifactName(name, arch)))
	if err != nil {
		return nil, errors.Wrap(err, "could not find artifact at given path")
	}
	return b, nil
}

// Name of the repository.
func (l *LocalRepository) Name() string {
	return l.name
}

// Manifest returns the manifest for a local repository.
func (l *LocalRepository) Manifest() (manifest Manifest, err error) {
	b, err := os.ReadFile(filepath.Join(l.path, ManifestFileName))
	if err != nil {
		err = fmt.Errorf("could not find manifest.yaml file: %v", err)
		return
	}

	err = yaml.Unmarshal(b, &manifest)
	if err != nil {
		err = fmt.Errorf("could not unmarshal manifest.yaml: %v", err)
	}
	return
}

// VersionSelector returns the current default version finder.
func (l *LocalRepository) VersionSelector() VersionSelector {
	return l.versionSelector
}
