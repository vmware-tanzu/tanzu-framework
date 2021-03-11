// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	configv1alpha1 "github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
)

const (
	// EnvConfigKey is the environment variable that points to a tanzu config.
	EnvConfigKey = "TANZU_CONFIG"

	// EnvEndpointKey is the environment variable that overrides the tanzu endpoint.
	EnvEndpointKey = "TANZU_ENDPOINT"

	//nolint:gosec // Avoid "hardcoded credentials" false positive.
	// EnvAPITokenKey is the environment variable that overrides the tanzu API token for global auth.
	EnvAPITokenKey = "TANZU_API_TOKEN"

	// ConfigName is the name of the config
	ConfigName = "config.yaml"
)

// LocalDirName is the name of the local directory in which tanzu state is stored.
var LocalDirName = ".tanzu"

// LocalDir returns the local directory in which tanzu state is stored.
func LocalDir() (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate local tanzu dir")
	}
	path = filepath.Join(home, LocalDirName)
	return
}

// ConfigPath returns the tanzu config path, checking for environment overrides.
func ConfigPath() (path string, err error) {
	localDir, err := LocalDir()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigKey)
	if !ok {
		path = filepath.Join(localDir, ConfigName)
		return
	}
	return
}

// NewConfig returns a new config.
func NewConfig() (*configv1alpha1.Config, error) {
	c := &configv1alpha1.Config{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories: DefaultRepositories,
			},
		},
	}
	err := StoreConfig(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ConfigNotExistError is thown when a tanzu config cannot be found.
type ConfigNotExistError struct {
	s string
}

// Error is the error message.
func (c *ConfigNotExistError) Error() string {
	return c.s
}

// NewConfigNotExistError returns a new ConfigNotExistError.
func NewConfigNotExistError(err error) *ConfigNotExistError {
	return &ConfigNotExistError{errors.Wrap(err, "failed to read config file").Error()}
}

// GetConfig retrieves the config from the local directory.
func GetConfig() (cfg *configv1alpha1.Config, err error) {
	cfgPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		cfg, err = NewConfig()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	scheme, err := configv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	var c configv1alpha1.Config
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode config file")
	}
	return &c, nil
}

// StoreConfig stores the config in the local directory.
func StoreConfig(cfg *configv1alpha1.Config) error {
	cfgPath, err := ConfigPath()
	if err != nil {
		return errors.Wrap(err, "could not find config path")
	}

	_, err = os.Stat(cfgPath)
	if os.IsNotExist(err) {
		localDir, err := LocalDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		err = os.MkdirAll(localDir, 0755)
		if err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not create config path")
	}

	scheme, err := configv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	buf := new(bytes.Buffer)
	if err := s.Encode(cfg, buf); err != nil {
		return errors.Wrap(err, "failed to encode config file")
	}
	// TODO (pbarker): need to consider races.
	if err = os.WriteFile(cfgPath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}
	return nil
}

// DeleteConfig deletes the config from the local directory.
func DeleteConfig() error {
	cfgPath, err := ConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(cfgPath)
	if err != nil {
		return errors.Wrap(err, "could not remove config")
	}
	return nil
}

// GetServer by name.
func GetServer(name string) (s *configv1alpha1.Server, err error) {
	cfg, err := GetConfig()
	if err != nil {
		return s, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == name {
			return server, nil
		}
	}
	return s, fmt.Errorf("could not find server %q", name)
}

// ServerExists tells whether the server by the given name exists.
func ServerExists(name string) (bool, error) {
	cfg, err := GetConfig()
	if err != nil {
		return false, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// AddServer adds a server to the config.
func AddServer(s *configv1alpha1.Server, setCurrent bool) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == s.Name {
			return fmt.Errorf("server %q already exists", s.Name)
		}
	}
	cfg.KnownServers = append(cfg.KnownServers, s)
	if setCurrent {
		cfg.CurrentServer = s.Name
	}
	return StoreConfig(cfg)
}

// PutServer adds or updates the server.
func PutServer(s *configv1alpha1.Server, setCurrent bool) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}
	newServers := []*configv1alpha1.Server{s}
	for _, server := range cfg.KnownServers {
		if server.Name == s.Name {
			continue
		}
		newServers = append(newServers, server)
	}
	cfg.KnownServers = newServers
	if setCurrent {
		cfg.CurrentServer = s.Name
	}
	return StoreConfig(cfg)
}

// RemoveServer adds a server to the config.
func RemoveServer(name string) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	newServers := []*configv1alpha1.Server{}
	for _, server := range cfg.KnownServers {
		if server.Name != name {
			newServers = append(newServers, server)
		}
	}
	cfg.KnownServers = newServers

	if cfg.CurrentServer == name {
		cfg.CurrentServer = ""
	}

	err = StoreConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// SetCurrentServer sets the current server.
func SetCurrentServer(name string) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}
	var exists bool
	for _, server := range cfg.KnownServers {
		if server.Name == name {
			exists = true
		}
	}
	if !exists {
		return fmt.Errorf("could not set current server; %q is not a known server", name)
	}
	cfg.CurrentServer = name
	err = StoreConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentServer sets the current server.
func GetCurrentServer() (s *configv1alpha1.Server, err error) {
	cfg, err := GetConfig()
	if err != nil {
		return s, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == cfg.CurrentServer {
			return server, nil
		}
	}
	return s, fmt.Errorf("current server %q not found in tanzu config", cfg.CurrentServer)
}

// EndpointFromServer returns the endpoint from server.
func EndpointFromServer(s *configv1alpha1.Server) (endpoint string, err error) {
	switch s.Type {
	case configv1alpha1.ManagementClusterServerType:
		return s.ManagementClusterOpts.Endpoint, nil
	case configv1alpha1.GlobalServerType:
		return s.GlobalOpts.Endpoint, nil
	default:
		return endpoint, fmt.Errorf("unknown server type %q", s.Type)
	}
}
