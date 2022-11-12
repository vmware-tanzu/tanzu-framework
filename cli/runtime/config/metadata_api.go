// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// GetMetadata retrieves Metadata
func GetMetadata() (*configapi.Metadata, error) {
	// Retrieve config metadata node
	node, err := getMetadataNode()
	if err != nil {
		return nil, err
	}
	return getMetadata(node)
}

// GetConfigMetadata retrieves configMetadata
func GetConfigMetadata() (*configapi.ConfigMetadata, error) {
	// Retrieve config metadata node
	node, err := getMetadataNode()
	if err != nil {
		return nil, err
	}
	return getConfigMetadata(node)
}

// GetConfigMetadataPatchStrategy retrieves patch strategies
func GetConfigMetadataPatchStrategy() (map[string]string, error) {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return nil, err
	}
	return getConfigMetadataPatchStrategy(node)
}

// SetConfigMetadataPatchStrategy add or update patch strategy specified by key-value pair
func SetConfigMetadataPatchStrategy(key, value string) error {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return err
	}

	// Add or update patch strategy
	err = setConfigMetadataPatchStrategy(node, key, value)
	if err != nil {
		return err
	}
	return persistConfigMetadata(node)
}

// SetConfigMetadataPatchStrategies add or update map of patch strategies
func SetConfigMetadataPatchStrategies(patchStrategies map[string]string) error {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	node, err := getMetadataNodeNoLock()
	if err != nil {
		return err
	}

	// Add or update patch strategies
	err = setConfigMetadataPatchStrategies(node, patchStrategies)
	if err != nil {
		return err
	}
	return persistConfigMetadata(node)
}

func getConfigMetadata(node *yaml.Node) (*configapi.ConfigMetadata, error) {
	metadata, err := convertNodeToMetadata(node)
	if err != nil {
		return nil, err
	}

	if metadata != nil && metadata.ConfigMetadata != nil {
		return metadata.ConfigMetadata, nil
	}

	return nil, errors.New("config metadata not found")
}

func getConfigMetadataPatchStrategy(node *yaml.Node) (map[string]string, error) {
	metadata, err := convertNodeToMetadata(node)
	if err != nil {
		return nil, err
	}
	if metadata != nil && metadata.ConfigMetadata != nil &&
		metadata.ConfigMetadata.PatchStrategy != nil {
		return metadata.ConfigMetadata.PatchStrategy, nil
	}
	return nil, errors.New("config metadata patch strategy not found")
}

func getMetadata(node *yaml.Node) (*configapi.Metadata, error) {
	metadata, err := convertNodeToMetadata(node)
	if err != nil {
		return nil, err
	}
	return metadata, nil
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
		return errors.New("allowed values are replace or merge")
	}

	// find patch strategy node
	keys := []nodeutils.Key{
		{Name: KeyConfigMetadata, Type: yaml.MappingNode},
		{Name: KeyPatchStrategy, Type: yaml.MappingNode},
	}
	patchStrategyNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
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
