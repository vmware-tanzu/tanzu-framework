// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// GetCLIDiscoverySources retrieves cli discovery sources
func GetCLIDiscoverySources() ([]configapi.PluginDiscovery, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getCLIDiscoverySources(node)
}

// GetCLIDiscoverySource retrieves cli discovery source by name
func GetCLIDiscoverySource(name string) (*configapi.PluginDiscovery, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getCLIDiscoverySource(node, name)
}

// SetCLIDiscoverySources Add/Update array of cli discovery sources to the yaml node
func SetCLIDiscoverySources(discoverySources []configapi.PluginDiscovery) (err error) {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// Loop through each discovery source and add or update existing node
	for _, discoverySource := range discoverySources {
		persist, err := setCLIDiscoverySource(node, discoverySource)
		if err != nil {
			return err
		}
		// Persist the config node to the file
		if persist {
			err = persistConfig(node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SetCLIDiscoverySource add or update a cli discoverySource
func SetCLIDiscoverySource(discoverySource configapi.PluginDiscovery) (err error) {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// Add/Update cli discovery source in the yaml node
	persist, err := setCLIDiscoverySource(node, discoverySource)
	if err != nil {
		return err
	}

	// Persist the config node to the file
	if persist {
		return persistConfig(node)
	}

	return err
}

// DeleteCLIDiscoverySource delete cli discoverySource by name
func DeleteCLIDiscoverySource(name string) error {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// Delete the matching cli discovery source from the yaml node
	err = deleteCLIDiscoverySource(node, name)
	if err != nil {
		return err
	}

	// Persist the config node to the file
	return persistConfig(node)
}

func getCLIDiscoverySources(node *yaml.Node) ([]configapi.PluginDiscovery, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.DiscoverySources != nil {
		return cfg.ClientOptions.CLI.DiscoverySources, nil
	}
	return nil, errors.New("cli discovery sources not found")
}

func getCLIDiscoverySource(node *yaml.Node, name string) (*configapi.PluginDiscovery, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.DiscoverySources != nil {
		for _, discoverySource := range cfg.ClientOptions.CLI.DiscoverySources {
			_, discoverySourceName := getDiscoverySourceTypeAndName(discoverySource)
			if discoverySourceName == name {
				return &discoverySource, nil
			}
		}
	}
	return nil, errors.New("cli discovery source not found")
}

// setCLIDiscoverySources Add/Update array of cli discovery sources to the yaml node
func setCLIDiscoverySources(node *yaml.Node, discoverySources []configapi.PluginDiscovery) (err error) {
	for _, discoverySource := range discoverySources {
		_, err = setCLIDiscoverySource(node, discoverySource)
		if err != nil {
			return err
		}
	}
	return err
}

// setCLIDiscoverySource Add/Update cli discovery source in the yaml node
func setCLIDiscoverySource(node *yaml.Node, discoverySource configapi.PluginDiscovery) (persist bool, err error) {
	// Retrieve the patch strategies from config metadata
	patchStrategies, err := GetConfigMetadataPatchStrategy()
	if err != nil {
		patchStrategies = make(map[string]string)
	}

	// Find the cli discovery sources node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions, Type: yaml.MappingNode},
		{Name: KeyCLI, Type: yaml.MappingNode},
		{Name: KeyDiscoverySources, Type: yaml.SequenceNode},
	}
	discoverySourcesNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if discoverySourcesNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}

	// Add or Update cli discovery source to discovery sources node based on patch strategy
	key := fmt.Sprintf("%v.%v.%v", KeyClientOptions, KeyCLI, KeyDiscoverySources)
	return setDiscoverySource(discoverySourcesNode, discoverySource, nodeutils.WithPatchStrategyKey(key), nodeutils.WithPatchStrategies(patchStrategies))
}

func deleteCLIDiscoverySource(node *yaml.Node, name string) error {
	// Find cli discovery sources node in the yaml node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions},
		{Name: KeyCLI},
		{Name: KeyDiscoverySources},
	}
	cliDiscoverySourcesNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if cliDiscoverySourcesNode == nil {
		return nil
	}

	// Get matching cli discovery source from the yaml node
	discoverySource, err := getCLIDiscoverySource(node, name)
	if err != nil {
		return err
	}
	discoverySourceType, discoverySourceName := getDiscoverySourceTypeAndName(*discoverySource)
	var result []*yaml.Node
	for _, discoverySourceNode := range cliDiscoverySourcesNode.Content {
		// Find discovery source matched by discoverySourceType
		if discoverySourceIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content, discoverySourceType); discoverySourceIndex != -1 {
			// Find matching discovery source
			if discoverySourceFieldIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content[discoverySourceIndex].Content, "name"); discoverySourceFieldIndex != -1 && discoverySourceNode.Content[discoverySourceIndex].Content[discoverySourceFieldIndex].Value == discoverySourceName {
				continue
			}
		}
		result = append(result, discoverySourceNode)
	}
	cliDiscoverySourcesNode.Style = 0
	cliDiscoverySourcesNode.Content = result
	return nil
}
