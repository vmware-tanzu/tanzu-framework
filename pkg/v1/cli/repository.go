package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

// Repository is a remote repository containing plugin artifacts.
type Repository interface {
	// List available plugins.
	List() ([]PluginDescriptor, error)

	// Describe a plugin.
	Describe(name string) (PluginDescriptor, error)

	// Fetch an artifact.
	Fetch(name, version string, arch Arch) ([]byte, error)

	// Name of the repository.
	Name() string

	// Manifest retrieves the manifest for the repo.
	Manifest() (Manifest, error)
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
	Plugins []PluginDescriptor `json:"plugins" yaml:"plugins"`

	// Version of the root CLI.
	Version string `json:"version" yaml:"version"`
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
	// DarwinAMD64 arch.
	DarwinAMD64 Arch = "darwin_amd64"
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

	// VersionLatest is the latest version.
	VersionLatest = "latest"
	// AllPlugins is the keyword for all plugins.
	AllPlugins = "all"
)

// GCPBucketRepository is a artifact repository utilizing a GCP bucket.
type GCPBucketRepository struct {
	bucketName string
	rootPath   string
	name       string
}

// CommunityRepositoryName is the community repository name.
const CommunityRepositoryName = "community"

// CommunityGCPBucketRepository is the default GCP bucket repository.
var CommunityGCPBucketRepository = &GCPBucketRepository{
	bucketName: "tanzu-cli",
	rootPath:   DefaultArtifactsDirectory,
	name:       CommunityRepositoryName,
}

// TMCGCPBucketRepository is the default GCP bucket repository.
var TMCGCPBucketRepository = &GCPBucketRepository{
	bucketName: "tmc-cli-plugins",
	rootPath:   DefaultArtifactsDirectory,
	name:       "manage",
}

// NewGCPBucketRepository returns a new GCP bucket repository.
func NewGCPBucketRepository(options ...Option) Repository {
	opts := makeDefaultOptions(options...)

	return &GCPBucketRepository{
		bucketName: opts.gcpBucket,
		rootPath:   opts.gcpRootPath,
	}
}

// List available plugins.
func (g *GCPBucketRepository) List() (desc []PluginDescriptor, err error) {
	manifest, err := g.Manifest()
	if err != nil {
		return desc, err
	}
	desc = manifest.Plugins
	return
}

// Describe a plugin.
func (g *GCPBucketRepository) Describe(name string) (desc PluginDescriptor, err error) {
	ctx := context.Background()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return desc, err
	}

	pluginPath := filepath.Join(g.rootPath, name, PluginFileName)

	obj := bkt.Object(pluginPath)
	if obj == nil {
		return desc, fmt.Errorf("artifact %q not found", name)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return desc, errors.Wrap(err, "could not fetch artifact from repository")
	}
	defer r.Close()

	d := yaml.NewDecoder(r)

	err = d.Decode(&desc)
	if err != nil {
		return desc, errors.Wrap(err, "could not decode plugin decriptor")
	}
	return
}

// Fetch an artifact.
func (g *GCPBucketRepository) Fetch(name, version string, arch Arch) ([]byte, error) {
	ctx := context.Background()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return nil, err
	}

	if version == VersionLatest {
		desc, err := g.Describe(name)
		if err != nil {
			return nil, err
		}
		version = desc.Version
	}

	artifactPath := filepath.Join(g.rootPath, name, version, MakeArtifactName(name, arch))
	obj := bkt.Object(artifactPath)
	if obj == nil {
		return nil, fmt.Errorf("artifact %q not found", name)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not read artifact")
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
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
	ctx := context.Background()

	bkt, err := g.getBucket(ctx)
	if err != nil {
		return manifest, err
	}

	manifestPath := filepath.Join(g.rootPath, ManifestFileName)

	obj := bkt.Object(manifestPath)
	if obj == nil {
		return manifest, fmt.Errorf("could not fetch manifest from repository")
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return manifest, errors.Wrap(err, "could not fetch manifest from repository")
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

// LocalRepository is a artifact repository utilizing a local host os.
type LocalRepository struct {
	path string
	name string
}

// DefaultLocalRepository is the default local repository.
var DefaultLocalRepository = &LocalRepository{
	path: fmt.Sprintf("./%s", DefaultArtifactsDirectory),
}

// NewLocalRepository returns a new local repository.
func NewLocalRepository(name, path string) Repository {
	return &LocalRepository{
		path: path,
		name: name,
	}
}

// List available plugins.
func (l *LocalRepository) List() (desc []PluginDescriptor, err error) {
	manifest, err := l.Manifest()
	if err != nil {
		return desc, err
	}
	desc = manifest.Plugins
	return
}

// Describe a plugin.
func (l *LocalRepository) Describe(name string) (desc PluginDescriptor, err error) {
	b, err := ioutil.ReadFile(filepath.Join(l.path, name, PluginFileName))
	if err != nil {
		err = fmt.Errorf("could not find plugin.yaml file for plugin %q: %v", name, err)
		return
	}

	err = yaml.Unmarshal(b, &desc)
	if err != nil {
		err = fmt.Errorf("could not unmarshal manifest.yaml: %v", err)
	}
	return
}

// Fetch an artifact.
func (l *LocalRepository) Fetch(name, version string, arch Arch) ([]byte, error) {
	if version == VersionLatest {
		desc, err := l.Describe(name)
		if err != nil {
			return nil, err
		}
		version = desc.Version
	}
	b, err := ioutil.ReadFile(filepath.Join(l.path, name, version, MakeArtifactName(name, arch)))
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
	b, err := ioutil.ReadFile(filepath.Join(l.path, ManifestFileName))
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
