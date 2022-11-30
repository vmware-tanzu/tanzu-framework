// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"

	"gopkg.in/yaml.v3"
)

// GetAllEnvs retrieves all env values from config
func GetAllEnvs() (map[string]string, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getAllEnvs(node)
}

func getAllEnvs(node *yaml.Node) (map[string]string, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.Env != nil {
		return cfg.ClientOptions.Env, nil
	}
	return nil, errors.New("not found")
}

// GetEnv retrieves env value by key
func GetEnv(key string) (string, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return "", err
	}
	return getEnv(node, key)
}

func getEnv(node *yaml.Node, key string) (string, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return "", err
	}
	if cfg.ClientOptions == nil && cfg.ClientOptions.Env == nil {
		return "", errors.New("not found")
	}
	if val, ok := cfg.ClientOptions.Env[key]; ok {
		return val, nil
	}
	return "", errors.New("not found")
}

// DeleteEnv delete the env entry of specified key
func DeleteEnv(key string) error {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	err = deleteEnv(node, key)
	if err != nil {
		return err
	}
	return persistConfig(node)
}

func deleteEnv(node *yaml.Node, key string) (err error) {
	// find env node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions},
		{Name: KeyEnv},
	}
	envsNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if envsNode == nil {
		return err
	}

	// convert env nodes to map
	envs, err := nodeutils.ConvertNodeToMap(envsNode)
	if err != nil {
		return err
	}

	// delete the specified entry in the map
	delete(envs, key)

	// convert updated map to env node
	newEnvsNode, err := nodeutils.ConvertMapToNode(envs)
	if err != nil {
		return err
	}
	envsNode.Content = newEnvsNode.Content[0].Content
	return nil
}

// SetEnv add or update a env key and value
func SetEnv(key, value string) (err error) {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// add or update env map
	persist, err := setEnv(node, key, value)
	if err != nil {
		return err
	}
	if persist {
		return persistConfig(node)
	}
	return err
}

func setEnv(node *yaml.Node, key, value string) (persist bool, err error) {
	// find env node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions, Type: yaml.MappingNode},
		{Name: KeyEnv, Type: yaml.MappingNode},
	}
	envsNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if envsNode == nil {
		return persist, err
	}

	// convert env node to map
	envs, err := nodeutils.ConvertNodeToMap(envsNode)
	if err != nil {
		return persist, err
	}

	// add or update the envs map per specified key value pair
	if len(envs) == 0 || envs[key] != value {
		envs[key] = value
		persist = true
	}

	// convert map to yaml node
	newEnvsNode, err := nodeutils.ConvertMapToNode(envs)
	if err != nil {
		return false, err
	}

	envsNode.Content = newEnvsNode.Content[0].Content
	return persist, err
}

// GetEnvConfigurations returns a map of configured environment variables
// to values as part of tanzu configuration file
// it returns nil if configuration is not yet defined
func GetEnvConfigurations() map[string]string {
	envs, err := GetAllEnvs()
	if err != nil {
		return make(map[string]string)
	}
	return envs
}
