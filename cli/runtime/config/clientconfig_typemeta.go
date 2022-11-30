// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"

	"gopkg.in/yaml.v3"
)

func setKind(node *yaml.Node, kind string) (persist bool, err error) {
	return setScalarNode(node, KeyKind, kind)
}

func setAPIVersion(node *yaml.Node, apiVersion string) (persist bool, err error) {
	return setScalarNode(node, KeyAPIVersion, apiVersion)
}

// setScalarNode adds or updated scalar node value in yaml node
func setScalarNode(node *yaml.Node, key, value string) (persist bool, err error) {
	keys := []nodeutils.Key{
		{Name: key, Type: yaml.ScalarNode, Value: ""},
	}
	targetNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if targetNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}
	if targetNode.Value != value {
		targetNode.Value = value
		persist = true
	}
	return persist, err
}
