// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config Provide API methods to Read/Write specific stanza of config file
package config

import (
	"os"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/collectionutils"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// LegacyConfigNodeKeys config nodes that goes to config.yaml
var LegacyConfigNodeKeys = []string{
	KeyAPIVersion,
	KeyKind,
	KeyMetadata,
	KeyClientOptions,
	KeyServers,
	KeyCurrentServer,
}

var DiscardedConfigNodeKeys = []string{
	KeyContexts,
	KeyCurrentContext,
}

// getMultiConfig retrieves combined config.yaml and config-ng.yaml
func getMultiConfig() (*yaml.Node, error) {
	cfgNode, err := getClientConfig()
	if err != nil {
		return cfgNode, err
	}

	cfgNextGenNode, err := getClientConfigNextGenNode()
	if err != nil {
		return cfgNextGenNode, err
	}

	return makeMultiFileCfg(cfgNode, cfgNextGenNode)
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
func makeMultiFileCfg(cfgNode, cfgNextGenNode *yaml.Node) (*yaml.Node, error) {
	// Create a root config document node
	rootCfgNode, err := newClientConfigNode()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new client config node")
	}

	// Loop through each config items and construct the root config node
	for _, cfgItem := range LegacyConfigNodeKeys {
		// Find the cfg item node from the config.yaml
		cfgNodeIndex := nodeutils.GetNodeIndex(cfgNode.Content[0].Content, cfgItem)
		if cfgNodeIndex == -1 {
			continue
		}
		// Add matching key node and value node from config.yaml to root cfg node
		rootCfgNode.Content[0].Content = append(rootCfgNode.Content[0].Content, cfgNode.Content[0].Content[cfgNodeIndex-1:cfgNodeIndex+1]...)
	}

	// Append the config-ng.yaml nodes to root config node
	rootCfgNode.Content[0].Content = append(rootCfgNode.Content[0].Content, cfgNextGenNode.Content[0].Content...)

	// return the construct root node that contains both config.yaml with cfgItems and all of config-ng.yaml
	return rootCfgNode, nil
}

// persistConfig write the updated node data to config.yaml and config-ng.yaml based on cfgItems
func persistConfig(node *yaml.Node) error {
	// check to persist multi file or to config-ng yaml
	useUnifiedConfig, err := UseUnifiedConfig()
	if err != nil {
		useUnifiedConfig = false
	}

	// If useUnifiedConfig is set to true write to config-ng.yaml
	if useUnifiedConfig {
		return persistClientConfigNextGen(node)
	}

	// config node from config.yaml
	cfgNode, err := getClientConfigNoLock()
	if err != nil {
		return err
	}

	// config next gen node from config-ng.yaml
	cfgNextGenNode, err := getClientConfigNextGenNodeNoLock()
	if err != nil {
		return err
	}

	// for each of the change node update the respective node in cfg and cfg-ng
	for index, changeNode := range node.Content[0].Content {
		if index%2 == 0 {
			// If contains then add or update the config.yaml
			if collectionutils.Contains(LegacyConfigNodeKeys, changeNode.Value) {
				// update in config-yaml
				currentChangeNodeIndex := nodeutils.GetNodeIndex(cfgNode.Content[0].Content, changeNode.Value)
				if currentChangeNodeIndex != -1 {
					// update the existing node value
					updated := node.Content[0].Content[index+1]
					cfgNode.Content[0].Content[currentChangeNodeIndex] = updated
				} else {
					// append the new node key and value
					cfgNode.Content[0].Content = append(cfgNode.Content[0].Content, node.Content[0].Content[index:index+2]...)
				}
				// If it doesn't contain then add or update the config-ng.yaml
			} else {
				// update in config-ng yaml
				currentChangeNextGenNodeIndex := nodeutils.GetNodeIndex(cfgNextGenNode.Content[0].Content, changeNode.Value)
				if currentChangeNextGenNodeIndex != -1 {
					// update the existing vaalue
					updated := node.Content[0].Content[index+1]
					cfgNextGenNode.Content[0].Content[currentChangeNextGenNodeIndex] = updated
				} else {
					// append the new node key and value
					cfgNextGenNode.Content[0].Content = append(cfgNextGenNode.Content[0].Content, node.Content[0].Content[index:index+2]...)
				}
			}
		}
	}

	// Discard nodes from config.yaml
	for _, discardedCfgNodeKey := range DiscardedConfigNodeKeys {
		// Discard node from config.yaml
		legacyDiscardedContextsNodeIndex := nodeutils.GetNodeIndex(cfgNode.Content[0].Content, discardedCfgNodeKey)
		if legacyDiscardedContextsNodeIndex != -1 {
			// remove discarded node
			cfgNode.Content[0].Content = append(cfgNode.Content[0].Content[:legacyDiscardedContextsNodeIndex-1], cfgNode.Content[0].Content[legacyDiscardedContextsNodeIndex+1:]...)
		}
	}

	// Store the non nextGenItem config data to config.yaml
	err = persistClientConfig(cfgNode)
	if err != nil {
		return err
	}

	// Store the nextGenItem config data to config-ng.yaml
	err = persistClientConfigNextGen(cfgNextGenNode)
	if err != nil {
		return err
	}

	// Store the config data to legacy client config file/location
	err = persistLegacyClientConfig(cfgNode)
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
