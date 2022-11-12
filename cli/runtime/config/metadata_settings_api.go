// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

const (
	SettingUseUnifiedConfig = "useUnifiedConfig"
)

// GetConfigMetadataSettings retrieves feature flags
func GetConfigMetadataSettings() (map[string]string, error) {
	// Retrieve Metadata config node
	node, err := getMetadataNode()
	if err != nil {
		return nil, err
	}

	return getConfigMetadataSettings(node)
}

func GetConfigMetadataSpecificSetting(key string) (string, error) {
	// Retrieve Metadata config node
	node, err := getMetadataNode()
	if err != nil {
		return "", err
	}

	return getConfigMetadataSpecificSetting(node, key)
}

// IsConfigMetadataSpecificSettingEnabled checks and returns whether specific plugin and key is true
func IsConfigMetadataSpecificSettingEnabled(key string) (bool, error) {
	node, err := getMetadataNode()
	if err != nil {
		return false, err
	}
	val, err := getConfigMetadataSpecificSetting(node, key)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(val, "true"), nil
}

// UseUnifiedConfig checks migrateToNewConfig feature flag
func UseUnifiedConfig() (bool, error) {
	return IsConfigMetadataSpecificSettingEnabled(SettingUseUnifiedConfig)
}

// DeleteConfigMetadataSpecificSetting delete the env entry of specified key
func DeleteConfigMetadataSpecificSetting(key string) error {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return err
	}

	err = deleteSpecificSetting(node, key)
	if err != nil {
		return err
	}

	return persistConfigMetadata(node)
}

// SetConfigMetadataSpecificSetting add or update a env key and value
func SetConfigMetadataSpecificSetting(key, value string) (err error) {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return err
	}

	persist, err := setConfigMetadataSpecificSetting(node, key, value)

	if persist {
		return persistConfigMetadata(node)
	}

	return err
}

func getConfigMetadataSettings(node *yaml.Node) (map[string]string, error) {
	cfgMetadata, err := convertNodeToMetadata(node)
	if err != nil {
		return nil, err
	}
	if cfgMetadata != nil && cfgMetadata.ConfigMetadata != nil &&
		cfgMetadata.ConfigMetadata.Settings != nil {
		return cfgMetadata.ConfigMetadata.Settings, nil
	}
	return nil, nil
}

func getConfigMetadataSpecificSetting(node *yaml.Node, key string) (string, error) {
	cfgMetadata, err := convertNodeToMetadata(node)
	if err != nil {
		return "", err
	}

	if cfgMetadata == nil || cfgMetadata.ConfigMetadata == nil ||
		cfgMetadata.ConfigMetadata.Settings == nil {
		return "", errors.New("not found")
	}

	if val, ok := cfgMetadata.ConfigMetadata.Settings[key]; ok {
		return val, nil
	}
	return "", errors.New("not found")
}

func deleteSpecificSetting(node *yaml.Node, key string) (err error) {
	// find feature flags node
	keys := []nodeutils.Key{
		{Name: KeyConfigMetadata, Type: yaml.MappingNode},
		{Name: KeySettings, Type: yaml.MappingNode},
	}
	featureFlagsNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if featureFlagsNode == nil {
		return err
	}

	// convert env nodes to map
	featureFlags, err := nodeutils.ConvertNodeToMap(featureFlagsNode)
	if err != nil {
		return err
	}

	// delete the specified entry in the map
	delete(featureFlags, key)

	// convert updated map to env node
	newFeatureFlagNode, err := nodeutils.ConvertMapToNode(featureFlags)
	if err != nil {
		return err
	}
	featureFlagsNode.Content = newFeatureFlagNode.Content[0].Content
	return nil
}

//nolint:dupl
func setConfigMetadataSpecificSetting(node *yaml.Node, key, value string) (persist bool, err error) {
	// find feature flags stanza node
	keys := []nodeutils.Key{
		{Name: KeyConfigMetadata, Type: yaml.MappingNode},
		{Name: KeySettings, Type: yaml.MappingNode},
	}
	settingsNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if settingsNode == nil {
		return persist, err
	}
	// convert env node to map
	settings, err := nodeutils.ConvertNodeToMap(settingsNode)
	if err != nil {
		return persist, err
	}
	// add or update the envs map per specified key value pair
	if len(settings) == 0 || settings[key] != value {
		settings[key] = value
		persist = true
	}
	// convert map to yaml node
	newFeatureFlagsNode, err := nodeutils.ConvertMapToNode(settings)
	if err != nil {
		return persist, err
	}
	settingsNode.Content = newFeatureFlagsNode.Content[0].Content
	return persist, err
}
