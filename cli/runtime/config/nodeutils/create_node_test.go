// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCreateNode(t *testing.T) {
	tests := []struct {
		name string
		key  Key
		dst  []*yaml.Node
	}{
		{
			name: "success create scalar node",
			key: Key{
				Name:  "test",
				Type:  yaml.ScalarNode,
				Value: "",
			},
			dst: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Value: "test",
					Style: 0,
					Tag:   "!!str",
				},
				{
					Kind:  yaml.ScalarNode,
					Value: "",
					Style: 0,
					Tag:   "!!str",
				},
			},
		},
		{
			name: "success create sequence node",
			key: Key{
				Name: "test",
				Type: yaml.SequenceNode,
			},
			dst: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Value: "test",
					Style: 0,
				},
				{
					Kind:  yaml.SequenceNode,
					Style: 0,
				},
			},
		},
		{
			name: "success create mapping node",
			key: Key{
				Name: "test",
				Type: yaml.MappingNode,
			},
			dst: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Value: "test",
					Style: 0,
				},
				{
					Kind:  yaml.MappingNode,
					Style: 0,
				},
			},
		},
		{
			name: "should return nil for not supported node type",
			key: Key{
				Name: "test",
				Type: yaml.DocumentNode,
			},
			dst: nil,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			node := CreateNode(spec.key)
			assert.Equal(t, node, spec.dst)
		})
	}
}
