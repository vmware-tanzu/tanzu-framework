// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"
)

// ConvertNodeToMap converts yaml node to map[string]string
func ConvertNodeToMap(node *yaml.Node) (envs map[string]string, err error) {
	err = node.Decode(&envs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert node to map")
	}
	return envs, err
}

// ConvertMapToNode converts map[string]string to yaml node
func ConvertMapToNode(envs map[string]string) (*yaml.Node, error) {
	bytes, err := yaml.Marshal(envs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal map")
	}
	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal bytes to node")
	}
	node.Style = 0
	node.Content[0].Style = 0
	return &node, nil
}

// ConvertNodeToMapInterface converts yaml node to map[string]interface
func ConvertNodeToMapInterface(node *yaml.Node) (envs map[string]interface{}, err error) {
	err = node.Decode(&envs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert node to map interface")
	}
	return envs, err
}
