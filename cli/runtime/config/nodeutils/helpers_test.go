// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUniqNode(t *testing.T) {
	tests := []struct {
		name  string
		nodes []*yaml.Node
		count int
	}{
		{
			name:  "success 2 uniq nodes",
			count: 2,
			nodes: []*yaml.Node{
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
			name:  "success 1 uniq nodes",
			count: 1,
			nodes: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Value: "test",
					Style: 0,
					Tag:   "!!str",
				},
				{
					Kind:  yaml.ScalarNode,
					Value: "test",
					Style: 0,
					Tag:   "!!str",
				},
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			nodes := UniqNodes(spec.nodes)
			assert.Equal(t, spec.count, len(nodes))
		})
	}
}
