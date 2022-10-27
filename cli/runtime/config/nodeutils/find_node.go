package nodeutils

import (
	"gopkg.in/yaml.v3"
)

func FindNode(node *yaml.Node, opts ...Options) (*yaml.Node, error) {
	nodeConfig := &Config{}
	for _, opt := range opts {
		opt(nodeConfig)
	}

	parent := node
	for _, nodeKey := range nodeConfig.Keys {
		child := GetNode(parent, nodeKey.Name)
		if child == nil {
			if nodeConfig.ForceCreate {
				parent.Content = append(parent.Content, CreateNode(nodeKey)...)
				child = GetNode(parent, nodeKey.Name)
			} else {
				return nil, nil
			}
		}
		parent = child
		parent.Style = 0
	}
	return parent, nil
}

func GetNode(node *yaml.Node, key string) *yaml.Node {
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
