package cli

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

// optionsConfig is where the options are configured.
type optionsConfig struct {
	// pluginRoot is the directory that contains the plugins.
	pluginRoot string

	// gcpBucket is the bucket name for the gcp artifact repository.
	gcpBucket string

	// gcpRootPath is the root bucket path for the gcp artifact repository.
	gcpRootPath string

	// repoName is the repository name.
	repoName string

	// distro is the plugin distro to install with the CLI.
	distro Distro
}

var (
	// DefaultPluginRoot is the default plugin root.
	DefaultPluginRoot = filepath.Join(xdg.DataHome, "tanzu-cli")
)

// makeDefaultOptions creates the default options for this namespace.
func makeDefaultOptions(list ...Option) optionsConfig {
	opts := optionsConfig{
		// by default, the plugin root is at $XDG_DATA_HOME/tanzu-cli
		pluginRoot:  DefaultPluginRoot,
		gcpBucket:   CommunityGCPBucketRepository.bucketName,
		gcpRootPath: CommunityGCPBucketRepository.rootPath,
		distro:      DefaultDistro,
	}

	for _, o := range list {
		o(&opts)
	}

	return opts
}

// Option is a filesystem store option.
type Option func(o *optionsConfig)

// WithPluginRoot sets the root which directory plugins live in.
func WithPluginRoot(root string) Option {
	return func(o *optionsConfig) {
		o.pluginRoot = root
	}
}

// WithGCPBucket sets the gcp bucket to use for the artifact repository.
func WithGCPBucket(name string) Option {
	return func(o *optionsConfig) {
		o.gcpBucket = name
	}
}

// WithGCPRootPath sets the gcp bucket root path to use for the artifact repository.
func WithGCPRootPath(path string) Option {
	return func(o *optionsConfig) {
		o.gcpRootPath = path
	}
}

// WithDistro sets the distro that should be installed with the CLI
func WithDistro(distro Distro) Option {
	return func(o *optionsConfig) {
		o.distro = distro
	}
}

// WithName sets the name
func WithName(name string) Option {
	return func(o *optionsConfig) {
		o.repoName = name
	}
}
