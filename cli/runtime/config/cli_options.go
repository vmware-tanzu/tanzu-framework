// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// GetEdition retrieves ClientOptions Edition
func GetEdition() (string, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return "", err
	}
	return getEdition(node)
}

// SetEdition adds or updates edition value
func SetEdition(val string) (err error) {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// Add or Update edition in the yaml node
	persist := setEdition(node, val)

	// Persist the config node to the file
	if persist {
		return persistConfig(node)
	}
	return err
}

func setEdition(node *yaml.Node, val string) (persist bool) {
	editionNode := getCLIOptionsChildNode(KeyEdition, node)
	if editionNode != nil && editionNode.Value != val {
		editionNode.Value = val
		persist = true
	}
	return persist
}

func getEdition(node *yaml.Node) (string, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return "", err
	}
	if cfg != nil && cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil {
		//nolint:staticcheck
		return string(cfg.ClientOptions.CLI.Edition), nil
	}
	return "", errors.New("edition not found")
}

func setUnstableVersionSelector(node *yaml.Node, name string) (persist bool) {
	unstableVersionSelectorNode := getCLIOptionsChildNode(KeyUnstableVersionSelector, node)
	if unstableVersionSelectorNode != nil && unstableVersionSelectorNode.Value != name {
		unstableVersionSelectorNode.Value = name
		persist = true
	}
	return persist
}

func setBomRepo(node *yaml.Node, repo string) (persist bool) {
	bomRepoNode := getCLIOptionsChildNode(KeyBomRepo, node)
	if bomRepoNode != nil && bomRepoNode.Value != repo {
		bomRepoNode.Value = repo
		persist = true
	}
	return persist
}

func setCompatibilityFilePath(node *yaml.Node, filepath string) (persist bool) {
	compatibilityFilePathNode := getCLIOptionsChildNode(KeyCompatibilityFilePath, node)
	if compatibilityFilePathNode.Value != filepath {
		compatibilityFilePathNode.Value = filepath
		persist = true
	}
	return persist
}

// getCLIOptionsChildNode parses the yaml node and returns the matched node based on configOptions
func getCLIOptionsChildNode(key string, node *yaml.Node) *yaml.Node {
	configOptions := func(c *nodeutils.CfgNode) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: key, Type: yaml.ScalarNode, Value: ""},
		}
	}
	keyNode := nodeutils.FindNode(node.Content[0], configOptions)
	return keyNode
}
