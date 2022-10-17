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
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCLIDiscoverySources(node)
}

// GetCLIDiscoverySource retrieves cli discovery source by name
func GetCLIDiscoverySource(name string) (*configapi.PluginDiscovery, error) {
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCLIDiscoverySource(node, name)
}

// SetCLIDiscoverySource add or update a cli discoverySource
func SetCLIDiscoverySource(discoverySource configapi.PluginDiscovery) (err error) {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	persist, err := setCLIDiscoverySource(node, discoverySource)
	if err != nil {
		return err
	}
	if persist {
		return persistConfig(node)
	}
	return err
}

// DeleteCLIDiscoverySource delete cli discoverySource by name
func DeleteCLIDiscoverySource(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	err = deleteCLIDiscoverySource(node, name)
	if err != nil {
		return err
	}
	return persistConfig(node)
}

func getCLIDiscoverySources(node *yaml.Node) ([]configapi.PluginDiscovery, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.DiscoverySources != nil {
		return cfg.ClientOptions.CLI.DiscoverySources, nil
	}
	return nil, errors.New("cli discovery sources not found")
}

//nolint:dupl
// Skip duplicate lint to merge get cli discovery source and cli repository code into one.
func getCLIDiscoverySource(node *yaml.Node, name string) (*configapi.PluginDiscovery, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
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

func setCLIDiscoverySources(node *yaml.Node, discoverySources []configapi.PluginDiscovery) (err error) {
	for _, discoverySource := range discoverySources {
		_, err = setCLIDiscoverySource(node, discoverySource)
		if err != nil {
			return err
		}
	}
	return err
}

func setCLIDiscoverySource(node *yaml.Node, discoverySource configapi.PluginDiscovery) (persist bool, err error) {
	patchStrategies, err := GetConfigMetadataPatchStrategy()
	if err != nil {
		patchStrategies = make(map[string]string)
	}
	patchStrategyOptions := &nodeutils.PatchStrategyOptions{
		Key:             fmt.Sprintf("%v.%v.%v", KeyClientOptions, KeyCLI, KeyDiscoverySources),
		PatchStrategies: patchStrategies,
	}
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyDiscoverySources, Type: yaml.SequenceNode},
		}
	}
	discoverySourcesNode := nodeutils.FindNode(node.Content[0], configOptions)
	if discoverySourcesNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}
	persist, err = setDiscoverySource(discoverySourcesNode, discoverySource, patchStrategyOptions)
	if err != nil {
		return persist, err
	}
	return persist, err
}

//nolint:dupl
// Skip duplicate lint to merge delete cli discovery source and cli repository code into one.
func deleteCLIDiscoverySource(node *yaml.Node, name string) error {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = false
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions},
			{Name: KeyCLI},
			{Name: KeyDiscoverySources},
		}
	}
	cliDiscoverySourcesNode := nodeutils.FindNode(node.Content[0], configOptions)
	if cliDiscoverySourcesNode == nil {
		return nil
	}
	discoverySource, err := getCLIDiscoverySource(node, name)
	if err != nil {
		return err
	}
	discoverySourceType, discoverySourceName := getDiscoverySourceTypeAndName(*discoverySource)
	var result []*yaml.Node
	for _, discoverySourceNode := range cliDiscoverySourcesNode.Content {
		if discoverySourceIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content, discoverySourceType); discoverySourceIndex != -1 {
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
