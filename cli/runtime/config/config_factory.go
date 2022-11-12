// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config Provide API methods to Read/Write specific stanza of config file
package config

import (
	"os"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// getMultiConfig retrieves combined config.yaml and config-ng.yaml
func getMultiConfig() (*yaml.Node, error) {
	node1, err := getClientConfig()
	if err != nil {
		return node1, err
	}

	node2, err := getClientConfigNextGenNode()
	if err != nil {
		return node2, err
	}

	return makeMultiFileCfg(node1, node2)
}

// getMultiConfigNoLock retrieves combined config.yaml and config-ng.yaml
func getMultiConfigNoLock() (*yaml.Node, error) {
	cfgNode, err := getClientConfigNoLock()
	if err != nil {
		return cfgNode, err
	}

	cfgNextGenNode, err := getClientConfigNextGenNodeNoLock()
	if err != nil {
		return cfgNextGenNode, err
	}

	return makeMultiFileCfg(cfgNode, cfgNextGenNode)
}

// Construct multi file node from config.yaml and config-ng.yaml nodes
func makeMultiFileCfg(node1, node2 *yaml.Node) (*yaml.Node, error) {
	// config items that goes to config-ng.yaml
	nextGenCfgItems := []string{
		KeyContexts,
		KeyCurrentContext,
	}

	cfgNode := node1.Content[0]

	for _, nextGenItem := range nextGenCfgItems {
		// Find the next gen item node from the config.yaml
		nextGenItemNodeIndex := nodeutils.GetNodeIndex(cfgNode.Content, nextGenItem)
		if nextGenItemNodeIndex == -1 {
			continue
		}
		// Delete the next gen item node from the config.yaml
		cfgNode.Content = append(cfgNode.Content[:nextGenItemNodeIndex-1], cfgNode.Content[nextGenItemNodeIndex:]...)
		cfgNode.Content = append(cfgNode.Content[:nextGenItemNodeIndex-1], cfgNode.Content[nextGenItemNodeIndex:]...)
	}

	// Construct Multi Bytes data
	var multiBytes []byte
	var multiNode yaml.Node

	// construct node1 bytes and append to multiBytes
	if node1.Content[0].Content != nil && len(node1.Content[0].Content) > 0 {
		node1Bytes, err := yaml.Marshal(node1)
		if err != nil {
			return nil, err
		}

		if len(node1Bytes) != 0 {
			multiBytes = append(multiBytes, node1Bytes...)
		}
	}

	// construct node 2 bytes and append to multiBytes
	if node2.Content[0].Content != nil && len(node2.Content[0].Content) > 0 {
		node2Bytes, err := yaml.Marshal(node2)
		if err != nil {
			return nil, err
		}

		if len(node2Bytes) != 0 {
			multiBytes = append(multiBytes, node2Bytes...)
		}
	}

	// create new yaml node if multiBytes contains no data or empty
	if multiBytes == nil {
		multiNode, err := newClientConfigNode()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new client config node")
		}

		return multiNode, err
	}

	// construct the multi node from multi bytes data
	if len(multiBytes) != 0 {
		err := yaml.Unmarshal(multiBytes, &multiNode)
		if err != nil {
			return nil, err
		}
	}

	return &multiNode, nil
}

func persistConfig(node *yaml.Node) error {
	migrate, err := UseUnifiedConfig()
	if err != nil {
		migrate = false
	}

	if migrate {
		return persistClientConfigNextGen(node)
	}

	cfgNode, err := getClientConfigNoLock()
	if err != nil {
		return err
	}

	// deep copy of change node
	var cfgNodeToPersist yaml.Node
	data, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &cfgNodeToPersist)
	if err != nil {
		return err
	}

	cfgNextGenNode, err := getClientConfigNextGenNodeNoLock()
	if err != nil {
		return err
	}

	// config items that goes to config-ng.yaml
	nextGenItemNodes := []*yaml.Node{
		{Value: KeyContexts, Kind: yaml.SequenceNode},
		{Value: KeyCurrentContext, Kind: yaml.MappingNode},
	}

	for _, nextGenItem := range nextGenItemNodes {
		// Find the nextGenItem node from the updated node
		itemNode := nodeutils.FindNode(cfgNodeToPersist.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys([]nodeutils.Key{
			{Name: nextGenItem.Value, Type: nextGenItem.Kind},
		}))

		// Find the nextGenItem node from config.yaml
		itemCfgNode := nodeutils.FindNode(cfgNode.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys([]nodeutils.Key{
			{Name: nextGenItem.Value, Type: nextGenItem.Kind},
		}))

		// Find the nextGenItem node from config-ng.yaml
		itemCfgNextGenNode := nodeutils.FindNode(cfgNextGenNode.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys([]nodeutils.Key{
			{Name: nextGenItem.Value, Type: nextGenItem.Kind},
		}))

		// Update nextGenItem node in config-ng
		*itemCfgNextGenNode = *itemNode

		// Reset nextGenItem node in config.yaml
		*itemNode = *itemCfgNode
	}

	// Store the non nextGenItem config data to config.yaml
	err = persistClientConfig(&cfgNodeToPersist)
	if err != nil {
		return err
	}

	// Store the nextGenItem config data to config-ng.yaml
	err = persistClientConfigNextGen(cfgNextGenNode)
	if err != nil {
		return err
	}

	// Store the config data to legacy client config file/location
	err = persistLegacyClientConfig(node)
	if err != nil {
		return err
	}

	return nil
}

// persistNode stores/writes the yaml node to config path specified in CfgOpts
func persistNode(node *yaml.Node, opts ...CfgOpts) error {
	configurations := &CfgOptions{}
	for _, opt := range opts {
		opt(configurations)
	}
	cfgPathExists, err := fileExists(configurations.CfgPath)
	if err != nil {
		return errors.Wrap(err, "failed to check config path existence")
	}
	if !cfgPathExists {
		localDir, err := LocalDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	}
	data, err := yaml.Marshal(node)
	if err != nil {
		return errors.Wrap(err, "failed to marshal nodeutils")
	}
	err = os.WriteFile(configurations.CfgPath, data, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write the config to file")
	}
	return nil
}
