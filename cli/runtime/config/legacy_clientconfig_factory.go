// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

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
		cfg = &configapi.ClientConfig{}
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
	node, err := getClientConfigNodeNoLock()
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
		_, err = setAPIVersion(node, cfg.APIVersion)
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
	err = clientConfigSetCurrentContext(cfg, node)
	if err != nil {
		return err
	}
	err = clientConfigSetClientOptions(cfg, node)
	if err != nil {
		return err
	}
	return persistConfig(node)
}

func clientConfigSetClientOptions(cfg *configapi.ClientConfig, node *yaml.Node) error {
	if cfg.ClientOptions != nil {
		err := clientConfigSetFeatures(cfg, node)
		if err != nil {
			return err
		}
		err = clientConfigSetEnvs(cfg, node)
		if err != nil {
			return err
		}
		err = clientConfigSetCLI(cfg, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func clientConfigSetCLI(cfg *configapi.ClientConfig, node *yaml.Node) (err error) {
	if cfg.ClientOptions.CLI != nil {
		err = clientConfigSetCLIRepositories(cfg, node)
		if err != nil {
			return err
		}
		err = clientConfigSetCLIDiscoverySources(cfg, node)
		if err != nil {
			return err
		}
		if cfg.ClientOptions.CLI.UnstableVersionSelector != "" {
			setUnstableVersionSelector(node, string(cfg.ClientOptions.CLI.UnstableVersionSelector))
		}
		//nolint:staticcheck
		// Disable deprecated lint warning
		if cfg.ClientOptions.CLI.Edition != "" {
			setEdition(node, string(cfg.ClientOptions.CLI.Edition))
		}
		//nolint:staticcheck
		// Disable deprecated lint warning
		if cfg.ClientOptions.CLI.BOMRepo != "" {
			setBomRepo(node, cfg.ClientOptions.CLI.BOMRepo)
		}
		//nolint:staticcheck
		// Disable deprecated lint warning
		if cfg.ClientOptions.CLI.CompatibilityFilePath != "" {
			setCompatibilityFilePath(node, cfg.ClientOptions.CLI.CompatibilityFilePath)
		}
	}
	return nil
}

func clientConfigSetCLIDiscoverySources(cfg *configapi.ClientConfig, node *yaml.Node) error {
	if cfg.ClientOptions.CLI.DiscoverySources != nil && len(cfg.ClientOptions.CLI.DiscoverySources) != 0 {
		err := setCLIDiscoverySources(node, cfg.ClientOptions.CLI.DiscoverySources)
		if err != nil {
			return err
		}
	}
	return nil
}

func clientConfigSetCLIRepositories(cfg *configapi.ClientConfig, node *yaml.Node) error {
	if cfg.ClientOptions.CLI.Repositories != nil && len(cfg.ClientOptions.CLI.Repositories) != 0 {
		err := setCLIRepositories(node, cfg.ClientOptions.CLI.Repositories)
		if err != nil {
			return err
		}
	}
	return nil
}

func clientConfigSetEnvs(cfg *configapi.ClientConfig, node *yaml.Node) error {
	if cfg.ClientOptions.Env != nil {
		for key, value := range cfg.ClientOptions.Env {
			_, err := setEnv(node, key, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func clientConfigSetFeatures(cfg *configapi.ClientConfig, node *yaml.Node) error {
	if cfg.ClientOptions.Features != nil {
		for plugin := range cfg.ClientOptions.Features {
			for key, value := range cfg.ClientOptions.Features[plugin] {
				_, err := setFeature(node, plugin, key, value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func clientConfigSetCurrentContext(cfg *configapi.ClientConfig, node *yaml.Node) error {
	if cfg.CurrentContext != nil {
		for _, contextName := range cfg.CurrentContext {
			ctx, contextErr := cfg.GetContext(contextName)
			if contextErr != nil {
				return contextErr
			}
			_, err := setCurrentContext(node, ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
