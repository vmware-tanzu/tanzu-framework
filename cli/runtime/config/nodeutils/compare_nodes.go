// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

// Equal checks whether the passed two nodes are equal
func Equal(node1, node2 *yaml.Node) (bool, error) {
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

// NotEqual checks whether the passed two nodes are not deep equal
func NotEqual(node1, node2 *yaml.Node) (bool, error) {
	equal, err := Equal(node1, node2)
	if err != nil {
		return false, err
	}
	return !equal, nil
}
