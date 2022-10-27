package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func NewClientConfig() *configapi.ClientConfig {
	c := newClientConfig()
	return c
}

func newClientConfig() *configapi.ClientConfig {
	c := &configapi.ClientConfig{}

	// Check if the lock is acquired by the current process or not
	// If not try to acquire the lock before Storing the client config
	// and release the lock after updating the config
	if !IsTanzuConfigLockAcquired() {
		AcquireTanzuConfigLock()
		defer ReleaseTanzuConfigLock()
	}

	return c
}

// GetClientConfig retrieves the config from the local directory with file lock
func GetClientConfig() (cfg *configapi.ClientConfig, err error) {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	return GetClientConfigNoLock()
}

// GetClientConfigNoLock retrieves the config from the local directory without acquiring the lock
func GetClientConfigNoLock() (cfg *configapi.ClientConfig, err error) {

	cfgPath, err := ClientConfigPath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(cfgPath)
	if err != nil || len(b) == 0 {
		_ = fmt.Errorf("failed to read in config: %v\n", err)
		cfg = NewClientConfig()

		return cfg, nil
	}
	// Logging

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct struct from config data")
	}

	return cfg, nil
}

// StoreClientConfig stores the config in the local directory.
// Make sure to Acquire and Release tanzu lock when reading/writing to the
// tanzu client configuration
// Deprecated: StoreClientConfig is deprecated. Use New Config API methods
func StoreClientConfig(cfg *configapi.ClientConfig) error {
	// new plugins would be setting only contexts, so populate servers for backwards compatibility
	populateServers(cfg)
	// old plugins would be setting only servers, so populate contexts for forwards compatibility
	PopulateContexts(cfg)

	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	if cfg.Kind != "" {
		_, err = setKind(node, cfg.Kind)
		if err != nil {
			return err
		}
	}

	if cfg.APIVersion != "" {
		_, err = setApiVersion(node, cfg.APIVersion)
		if err != nil {
			return err
		}
	}

	err = setServers(node, cfg.KnownServers)
	if err != nil {
		return err
	}

	if cfg.CurrentServer != "" {
		_, err = setCurrentServer(node, cfg.CurrentServer)
		if err != nil {
			return err
		}

	}

	err = setContexts(node, cfg.KnownContexts)
	if err != nil {
		return err
	}

	if cfg.CurrentContext != nil {
		for _, contextName := range cfg.CurrentContext {
			ctx, contextErr := cfg.GetContext(contextName)
			if contextErr != nil {
				return contextErr
			}
			_, err = setCurrentContext(node, ctx)
			if err != nil {
				return err
			}
		}

	}

	if cfg.ClientOptions != nil {
		if cfg.ClientOptions.Features != nil {
			for plugin := range cfg.ClientOptions.Features {
				for key, value := range cfg.ClientOptions.Features[plugin] {
					_, err = setFeature(node, plugin, key, value)
					if err != nil {
						return err
					}
				}
			}
		}

		if cfg.ClientOptions.Env != nil {
			for key, value := range cfg.ClientOptions.Env {
				_, err = setEnv(node, key, value)
				if err != nil {
					return err
				}
			}
		}

		if cfg.ClientOptions.CLI != nil {
			// TODO : Test Set Repositories
			if cfg.ClientOptions.CLI.Repositories != nil && len(cfg.ClientOptions.CLI.Repositories) != 0 {
				err = setCLIRepositories(node, cfg.ClientOptions.CLI.Repositories)
				if err != nil {
					return err
				}
			}

			if cfg.ClientOptions.CLI.DiscoverySources != nil && len(cfg.ClientOptions.CLI.DiscoverySources) != 0 {
				err = setCLIDiscoverySources(node, cfg.ClientOptions.CLI.DiscoverySources)
				if err != nil {
					return err
				}
			}

			if cfg.ClientOptions.CLI.UnstableVersionSelector != "" {
				_, err = setUnstableVersionSelector(node, string(cfg.ClientOptions.CLI.UnstableVersionSelector))
				if err != nil {
					return err
				}

			}

			if cfg.ClientOptions.CLI.Edition != "" {
				_, err = setEdition(node, string(cfg.ClientOptions.CLI.Edition))
				if err != nil {
					return err
				}

			}

			if cfg.ClientOptions.CLI.BOMRepo != "" {
				_, err = setBomRepo(node, cfg.ClientOptions.CLI.BOMRepo)
				if err != nil {
					return err
				}

			}

			if cfg.ClientOptions.CLI.CompatibilityFilePath != "" {
				_, err = setCompatibilityFilePath(node, cfg.ClientOptions.CLI.CompatibilityFilePath)
				if err != nil {
					return err
				}

			}

		}

	}

	return PersistNode(node)
}

// DeleteClientConfig deletes the config from the local directory.
func DeleteClientConfig() error {
	cfgPath, err := ClientConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(cfgPath)
	if err != nil {
		return errors.Wrap(err, "could not remove config")
	}
	return nil
}
