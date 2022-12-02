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

	return getSettings(node)
}

func GetConfigMetadataSetting(key string) (string, error) {
	// Retrieve Metadata config node
	node, err := getMetadataNode()
	if err != nil {
		return "", err
	}

	return getSetting(node, key)
}

// IsConfigMetadataSettingsEnabled checks and returns whether specific plugin and key is true
func IsConfigMetadataSettingsEnabled(key string) (bool, error) {
	node, err := getMetadataNode()
	if err != nil {
		return false, err
	}
	val, err := getSetting(node, key)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(val, "true"), nil
}

// UseUnifiedConfig checks useUnifiedConfig feature flag
func UseUnifiedConfig() (bool, error) {
	return IsConfigMetadataSettingsEnabled(SettingUseUnifiedConfig)
}

// DeleteConfigMetadataSetting delete the env entry of specified key
func DeleteConfigMetadataSetting(key string) error {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return err
	}

	err = deleteSetting(node, key)
	if err != nil {
		return err
	}

	return persistConfigMetadata(node)
}

// SetConfigMetadataSetting add or update a env key and value
func SetConfigMetadataSetting(key, value string) (err error) {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return err
	}

	persist, err := setSetting(node, key, value)

	if persist {
		return persistConfigMetadata(node)
	}

	return err
}

func getSettings(node *yaml.Node) (map[string]string, error) {
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

func getSetting(node *yaml.Node, key string) (string, error) {
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

func deleteSetting(node *yaml.Node, key string) (err error) {
	// find settings node
	keys := []nodeutils.Key{
		{Name: KeyConfigMetadata, Type: yaml.MappingNode},
		{Name: KeySettings, Type: yaml.MappingNode},
	}
	settingsNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if settingsNode == nil {
		return err
	}

	// convert settings nodes to map
	settings, err := nodeutils.ConvertNodeToMap(settingsNode)
	if err != nil {
		return err
	}

	// delete the specified entry in the map
	delete(settings, key)

	// convert updated map to settings node
	newSettingNode, err := nodeutils.ConvertMapToNode(settings)
	if err != nil {
		return err
	}
	settingsNode.Content = newSettingNode.Content[0].Content
	return nil
}

//nolint:dupl
func setSetting(node *yaml.Node, key, value string) (persist bool, err error) {
	// find settings node
	keys := []nodeutils.Key{
		{Name: KeyConfigMetadata, Type: yaml.MappingNode},
		{Name: KeySettings, Type: yaml.MappingNode},
	}
	settingsNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if settingsNode == nil {
		return persist, err
	}
	// convert settings node to map
	settings, err := nodeutils.ConvertNodeToMap(settingsNode)
	if err != nil {
		return persist, err
	}
	// add or update the settings map per specified key value pair
	if len(settings) == 0 || settings[key] != value {
		settings[key] = value
		persist = true
	}
	// convert map to settings node
	newSettingNode, err := nodeutils.ConvertMapToNode(settings)
	if err != nil {
		return persist, err
	}
	settingsNode.Content = newSettingNode.Content[0].Content
	return persist, err
}
