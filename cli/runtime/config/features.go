// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"

	"gopkg.in/yaml.v3"
)

// IsFeatureEnabled checks and returns whether specific plugin and key is true
func IsFeatureEnabled(plugin, key string) (bool, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return false, err
	}
	val, err := getFeature(node, plugin, key)
	if err != nil {
		return false, err
	}
	if strings.EqualFold(val, "true") {
		return true, nil
	}
	return false, nil
}

func getFeature(node *yaml.Node, plugin, key string) (string, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return "", err
	}
	if cfg.ClientOptions == nil || cfg.ClientOptions.Features == nil || cfg.ClientOptions.Features[plugin] == nil {
		return "", errors.New("not found")
	}
	if val, ok := cfg.ClientOptions.Features[plugin][key]; ok {
		return val, nil
	}
	return "", errors.New("not found")
}

// DeleteFeature deletes the specified plugin key
func DeleteFeature(plugin, key string) error {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	err = deleteFeature(node, plugin, key)
	if err != nil {
		return err
	}
	return persistConfig(node)
}

func deleteFeature(node *yaml.Node, plugin, key string) error {
	// Find plugin node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions},
		{Name: KeyFeatures},
		{Name: plugin},
	}
	pluginNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if pluginNode == nil {
		return nil
	}
	plugins, err := nodeutils.ConvertNodeToMap(pluginNode)
	if err != nil {
		return err
	}
	delete(plugins, key)
	newPluginsNode, err := nodeutils.ConvertMapToNode(plugins)
	if err != nil {
		return err
	}
	pluginNode.Content = newPluginsNode.Content[0].Content
	return nil
}

// SetFeature add or update plugin key value
func SetFeature(plugin, key, value string) (err error) {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	// Add or Update Feature plugin
	persist, err := setFeature(node, plugin, key, value)
	if err != nil {
		return err
	}
	if persist {
		return persistConfig(node)
	}
	return err
}

func setFeature(node *yaml.Node, plugin, key, value string) (persist bool, err error) {
	// find plugin node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions, Type: yaml.MappingNode},
		{Name: KeyFeatures, Type: yaml.MappingNode},
		{Name: plugin, Type: yaml.MappingNode},
	}
	pluginNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if pluginNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}
	if index := nodeutils.GetNodeIndex(pluginNode.Content, key); index != -1 {
		if pluginNode.Content[index].Value != value {
			pluginNode.Content[index].Tag = "!!str"
			pluginNode.Content[index].Value = value
			persist = true
		}
	} else {
		pluginNode.Content = append(pluginNode.Content, nodeutils.CreateScalarNode(key, value)...)
		persist = true
	}
	return persist, err
}

// ConfigureDefaultFeatureFlagsIfMissing add or update plugin features based on specified default feature flags
func ConfigureDefaultFeatureFlagsIfMissing(plugin string, defaultFeatureFlags map[string]bool) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	// find plugin node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions, Type: yaml.MappingNode},
		{Name: KeyFeatures, Type: yaml.MappingNode},
		{Name: plugin, Type: yaml.MappingNode},
	}
	pluginNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if pluginNode == nil {
		return nodeutils.ErrNodeNotFound
	}
	for key, value := range defaultFeatureFlags {
		val := strconv.FormatBool(value)
		if index := nodeutils.GetNodeIndex(pluginNode.Content, key); index != -1 {
			pluginNode.Content[index].Value = val
		} else {
			pluginNode.Content = append(pluginNode.Content, nodeutils.CreateScalarNode(key, val)...)
		}
	}
	return nil
}

// IsFeatureActivated returns true if the given feature is activated
// User can set this CLI feature flag using `tanzu config set features.global.<feature> true`
func IsFeatureActivated(feature string) bool {
	cfg, err := GetClientConfig()
	if err != nil {
		return false
	}
	status, err := cfg.IsConfigFeatureActivated(feature)
	if err != nil {
		return false
	}
	return status
}
