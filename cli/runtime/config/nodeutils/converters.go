package nodeutils

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func ConvertFromNode[T any](node *yaml.Node) (obj *T, err error) {
	err = node.Decode(&obj)
	if err != nil {
		return nil, err
	}
	return obj, err
}

func ConvertToNode[T any](obj *T) (*yaml.Node, error) {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func ConvertNodeToMap(node *yaml.Node) (envs map[string]string, err error) {
	err = node.Decode(&envs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode node to client config")
	}
	return envs, err
}

func ConvertMapToNode(envs map[string]string) (*yaml.Node, error) {
	bytes, err := yaml.Marshal(envs)
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, err
	}
	node.Style = 0
	node.Content[0].Style = 0
	return &node, nil
}

func ConvertNodeToMapInterface(node *yaml.Node) (envs map[string]interface{}, err error) {
	err = node.Decode(&envs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode node to client config")
	}
	return envs, err
}
