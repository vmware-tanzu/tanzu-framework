package nodeutils

import (
	"gopkg.in/yaml.v3"
)

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
