// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

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

// ClientConfigPath returns the tanzu config path, checking for environment overrides.
func ClientConfigPath() (path string, err error) {
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

// NewClientConfig returns a new config.
func NewClientConfig() (*configv1alpha1.ClientConfig, error) {
	c := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories: DefaultRepositories,
			},
		},
	}
	err := StoreClientConfig(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ClientConfigNotExistError is thrown when a tanzu config cannot be found.
type ClientConfigNotExistError struct {
	s string
}

// Error is the error message.
func (c *ClientConfigNotExistError) Error() string {
	return c.s
}

// NewConfigNotExistError returns a new ClientConfigNotExistError.
func NewConfigNotExistError(err error) *ClientConfigNotExistError {
	return &ClientConfigNotExistError{errors.Wrap(err, "failed to read config file").Error()}
}

// GetClientConfig retrieves the config from the local directory.
func GetClientConfig() (cfg *configv1alpha1.ClientConfig, err error) {
	cfgPath, err := ClientConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		cfg, err = NewClientConfig()
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
	var c configv1alpha1.ClientConfig
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode config file")
	}
	return &c, nil
}

// StoreClientConfig stores the config in the local directory.
func StoreClientConfig(cfg *configv1alpha1.ClientConfig) error {
	cfgPath, err := ClientConfigPath()
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
	// Set GVK explicitly as encoder does not do it.
	cfg.GetObjectKind().SetGroupVersionKind(configv1alpha1.GroupVersionKind)
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

// GetServer by name.
func GetServer(name string) (s *configv1alpha1.Server, err error) {
	cfg, err := GetClientConfig()
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
	cfg, err := GetClientConfig()
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
	cfg, err := GetClientConfig()
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
	return StoreClientConfig(cfg)
}

// PutServer adds or updates the server.
func PutServer(s *configv1alpha1.Server, setCurrent bool) error {
	cfg, err := GetClientConfig()
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
	return StoreClientConfig(cfg)
}

// RemoveServer adds a server to the config.
func RemoveServer(name string) error {
	cfg, err := GetClientConfig()
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

	err = StoreClientConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// SetCurrentServer sets the current server.
func SetCurrentServer(name string) error {
	cfg, err := GetClientConfig()
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
	err = StoreClientConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentServer sets the current server.
func GetCurrentServer() (s *configv1alpha1.Server, err error) {
	cfg, err := GetClientConfig()
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
