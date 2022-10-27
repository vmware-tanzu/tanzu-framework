package nodeutils

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

func Equal(node1 *yaml.Node, node2 *yaml.Node) (bool, error) {

	m1, err := ConvertNodeToMapInterface(node1)
	if err != nil {
		return false, err
	}
	m2, err := ConvertNodeToMapInterface(node2)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(m1, m2), nil
}

func NotEqual(node1 *yaml.Node, node2 *yaml.Node) (bool, error) {

	equal, err := Equal(node1, node2)
	if err != nil {
		return false, err
	}

	return !equal, nil
}
