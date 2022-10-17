// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"gopkg.in/yaml.v3"
)

// CreateNode creates yaml node based on node key Type
func CreateNode(nodeKey Key) []*yaml.Node {
	switch nodeKey.Type {
	case yaml.MappingNode:
		return CreateMappingNode(nodeKey.Name)
	case yaml.SequenceNode:
		return CreateSequenceNode(nodeKey.Name)
	case yaml.ScalarNode:
		return CreateScalarNode(nodeKey.Name, nodeKey.Value)
	}
	return nil
}

// CreateSequenceNode creates yaml node based on node key
func CreateSequenceNode(key string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: 0,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.SequenceNode,
		Style: 0,
	}
	return []*yaml.Node{keyNode, valueNode}
}

// CreateScalarNode creates yaml node based on node key and value
func CreateScalarNode(key, value string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
		Style: 0,
		Tag:   "!!str",
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
		Style: 0,
		Tag:   "!!str",
	}
	return []*yaml.Node{keyNode, valueNode}
}

// CreateMappingNode creates yaml node based on node key
func CreateMappingNode(key string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
		Style: 0,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.MappingNode,
		Style: 0,
	}
	return []*yaml.Node{keyNode, valueNode}
}
