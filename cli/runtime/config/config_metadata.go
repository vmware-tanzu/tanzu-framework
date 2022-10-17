// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"github.com/pkg/errors"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
	"gopkg.in/yaml.v3"
)

// GetConfigMetadata retrieves configMetadata
func GetConfigMetadata() (*configapi.Metadata, error) {
	node, err := GetConfigMetadataNode()
	if err != nil {
		return nil, err
	}
	return getConfigMetadata(node)
}

// GetConfigMetadataPatchStrategy retrieves patch strategies
func GetConfigMetadataPatchStrategy() (map[string]string, error) {
	node, err := GetConfigMetadataNode()
	if err != nil {
		return nil, err
	}
	return getConfigMetadataPatchStrategy(node)
}

// SetConfigMetadataPatchStrategy add or update patch strategy specified by key-value pair
func SetConfigMetadataPatchStrategy(key, value string) error {
	node, err := GetConfigMetadataNode()
	if err != nil {
		return err
	}
	err = setConfigMetadataPatchStrategy(node, key, value)
	if err != nil {
		return err
	}
	return persistConfigMetadata(node)
}

// SetConfigMetadataPatchStrategies add or update map of patch strategies
func SetConfigMetadataPatchStrategies(patchStrategies map[string]string) error {
	node, err := GetConfigMetadataNode()
	if err != nil {
		return err
	}
	err = setConfigMetadataPatchStrategies(node, patchStrategies)
	if err != nil {
		return err
	}
	return persistConfigMetadata(node)
}

func getConfigMetadataPatchStrategy(node *yaml.Node) (map[string]string, error) {
	cfgMetadata, err := nodeutils.ConvertFromNode[configapi.Metadata](node)
	if err != nil {
		return nil, err
	}
	if cfgMetadata != nil && cfgMetadata.ConfigMetadata != nil &&
		cfgMetadata.ConfigMetadata.PatchStrategy != nil {
		return cfgMetadata.ConfigMetadata.PatchStrategy, nil
	}
	return nil, nil
}

func getConfigMetadata(node *yaml.Node) (*configapi.Metadata, error) {
	cfgMetadata, err := nodeutils.ConvertFromNode[configapi.Metadata](node)
	if err != nil {
		return nil, err
	}
	return cfgMetadata, nil
}

func setConfigMetadataPatchStrategies(node *yaml.Node, patchStrategies map[string]string) error {
	for key, value := range patchStrategies {
		err := setConfigMetadataPatchStrategy(node, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func setConfigMetadataPatchStrategy(node *yaml.Node, key, value string) error {
	if !strings.EqualFold(value, "replace") && !strings.EqualFold(value, "merge") {
		return errors.New("allowed values replace or merge")
	}
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyConfigMetadata, Type: yaml.MappingNode},
			{Name: KeyPatchStrategy, Type: yaml.MappingNode},
		}
	}
	patchStrategyNode := nodeutils.FindNode(node.Content[0], configOptions)
	if patchStrategyNode == nil {
		return nodeutils.ErrNodeNotFound
	}
	if index := nodeutils.GetNodeIndex(patchStrategyNode.Content, key); index != -1 {
		patchStrategyNode.Content[index].Tag = nodeutils.NodeTagStr
		patchStrategyNode.Content[index].Value = value
	} else {
		patchStrategyNode.Content = append(patchStrategyNode.Content, nodeutils.CreateScalarNode(key, value)...)
	}
	return nil
}
