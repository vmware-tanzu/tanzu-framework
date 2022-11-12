// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"gopkg.in/yaml.v3"
)

// FindNode parse the yaml node and return the matched node based on Config Options passed
func FindNode(node *yaml.Node, opts ...Options) *yaml.Node {
	nodeConfig := &CfgNode{}
	for _, opt := range opts {
		opt(nodeConfig)
	}
	parent := node
	for _, nodeKey := range nodeConfig.Keys {
		child := getNode(parent, nodeKey.Name)
		if child == nil {
			if nodeConfig.ForceCreate {
				parent.Content = append(parent.Content, CreateNode(nodeKey)...)
				child = getNode(parent, nodeKey.Name)
			} else {
				return nil
			}
		}
		parent = child
		parent.Style = 0
	}
	return parent
}

// getNode parse the yaml node and return the node matched by key
func getNode(node *yaml.Node, key string) *yaml.Node {
	if node.Content == nil || len(node.Content) == 0 {
		return nil
	}
	nodeIndex := GetNodeIndex(node.Content, key)
	if nodeIndex == -1 {
		return nil
	}
	foundNode := node.Content[nodeIndex]
	return foundNode
}

// GetNodeIndex retrieves node index of specific node object value; yaml nodes are stored in arrays for all value types.
// Ex: refer https://pkg.go.dev/gopkg.in/yaml.v3#pkg-overview
func GetNodeIndex(node []*yaml.Node, key string) int {
	appIdx := -1
	for i, k := range node {
		if i%2 == 0 && k.Value == key {
			appIdx = i + 1
			break
		}
	}
	return appIdx
}
