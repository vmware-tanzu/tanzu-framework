// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

//nolint:all
// Disable unused lint
// setTypeMeta add or update typemeta stanza in the node config
func setTypeMeta(node *yaml.Node, typeMeta metav1.TypeMeta) (persist bool, err error) {
	persist, err = setKind(node, typeMeta.Kind)
	if err != nil {
		return persist, err
	}
	persist, err = setAPIVersion(node, typeMeta.APIVersion)
	if err != nil {
		return persist, err
	}
	return persist, err
}

func setKind(node *yaml.Node, kind string) (persist bool, err error) {
	return setScalarNode(node, KeyKind, kind)
}

func setAPIVersion(node *yaml.Node, apiVersion string) (persist bool, err error) {
	return setScalarNode(node, KeyAPIVersion, apiVersion)
}

// setScalarNode adds or updated scalar node value in yaml node
func setScalarNode(node *yaml.Node, key, value string) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: key, Type: yaml.ScalarNode, Value: ""},
		}
	}
	targetNode := nodeutils.FindNode(node.Content[0], configOptions)
	if targetNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}
	if targetNode.Value != value {
		targetNode.Value = value
		persist = true
	}
	return persist, err
}
