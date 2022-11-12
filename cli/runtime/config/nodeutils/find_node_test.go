// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestFindNode(t *testing.T) {
	tests := []struct {
		name      string
		in        *yaml.Node
		opts      Options
		out       *yaml.Node
		nullValue bool
	}{
		{
			name: "should return node from the already existing yaml node tree",
			in: &yaml.Node{
				Content: []*yaml.Node{
					{
						Value: "clientOptions",
					},
					{
						Content: []*yaml.Node{
							{
								Value: "cli",
							},
							{
								Content: []*yaml.Node{
									{Value: "discoverySources"},
									{
										Content: []*yaml.Node{
											{
												Content: []*yaml.Node{
													{Value: "gcp", Kind: yaml.ScalarNode},
													{Content: []*yaml.Node{
														{Value: "name"},
														{Value: "test"},
														{Value: "bucket"},
														{Value: "test-bucket"},
														{Value: "manifestPath"},
														{Value: "test-manifestPath"},
													}},
													{Value: "contextType", Kind: yaml.ScalarNode},
													{Value: "tmc", Kind: yaml.ScalarNode},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			out: &yaml.Node{
				Content: []*yaml.Node{
					{
						Content: []*yaml.Node{
							{Value: "gcp", Kind: yaml.ScalarNode},
							{Content: []*yaml.Node{
								{Value: "name"},
								{Value: "test"},
								{Value: "bucket"},
								{Value: "test-bucket"},
								{Value: "manifestPath"},
								{Value: "test-manifestPath"},
							}},
							{Value: "contextType", Kind: yaml.ScalarNode},
							{Value: "tmc", Kind: yaml.ScalarNode},
						},
					},
				},
			},
			opts: func(config *CfgNode) {
				config.ForceCreate = true
				config.Keys =
					[]Key{
						{Name: "clientOptions", Type: yaml.MappingNode},
						{Name: "cli", Type: yaml.MappingNode},
						{Name: "discoverySources", Type: yaml.SequenceNode},
					}
			},
		},
		{
			name: "should return node by force creating the node tree",
			in:   &yaml.Node{},
			out: &yaml.Node{
				Content: nil,
			},
			opts: func(config *CfgNode) {
				config.ForceCreate = true
				config.Keys =
					[]Key{
						{Name: "clientOptions", Type: yaml.MappingNode},
						{Name: "cli", Type: yaml.MappingNode},
						{Name: "discoverySources", Type: yaml.SequenceNode},
					}
			},
		},
		{
			name:      "should return nil when force create is false",
			in:        &yaml.Node{},
			out:       nil,
			nullValue: true,
			opts: func(config *CfgNode) {
				config.ForceCreate = false
				config.Keys =
					[]Key{
						{Name: "clientOptions", Type: yaml.MappingNode},
						{Name: "cli", Type: yaml.MappingNode},
						{Name: "discoverySources", Type: yaml.SequenceNode},
					}
			},
		},
		{
			name: "should return level one node from existing node tree",
			in: &yaml.Node{
				Content: []*yaml.Node{
					{
						Value: "clientOptions",
					},
					{
						Content: []*yaml.Node{
							{
								Value: "cli",
							},
							{
								Content: []*yaml.Node{
									{Value: "discoverySources"},
									{
										Content: []*yaml.Node{
											{
												Content: []*yaml.Node{
													{Value: "gcp", Kind: yaml.ScalarNode},
													{Content: []*yaml.Node{
														{Value: "name"},
														{Value: "test"},
														{Value: "bucket"},
														{Value: "test-bucket"},
														{Value: "manifestPath"},
														{Value: "test-manifestPath"},
													}},
													{Value: "contextType", Kind: yaml.ScalarNode},
													{Value: "tmc", Kind: yaml.ScalarNode},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			out: &yaml.Node{
				Content: []*yaml.Node{
					{
						Value: "cli",
					},
					{
						Content: []*yaml.Node{
							{Value: "discoverySources"},
							{
								Content: []*yaml.Node{
									{
										Content: []*yaml.Node{
											{Value: "gcp", Kind: yaml.ScalarNode},
											{Content: []*yaml.Node{
												{Value: "name"},
												{Value: "test"},
												{Value: "bucket"},
												{Value: "test-bucket"},
												{Value: "manifestPath"},
												{Value: "test-manifestPath"},
											}},
											{Value: "contextType", Kind: yaml.ScalarNode},
											{Value: "tmc", Kind: yaml.ScalarNode},
										},
									},
								},
							},
						},
					},
				},
			},
			opts: func(config *CfgNode) {
				config.ForceCreate = true
				config.Keys =
					[]Key{
						{Name: "clientOptions", Type: yaml.MappingNode},
					}
			},
		},
		{
			name: "should return level two node from existing node tree",
			in: &yaml.Node{
				Content: []*yaml.Node{
					{
						Value: "clientOptions",
					},
					{
						Content: []*yaml.Node{
							{
								Value: "cli",
							},
							{
								Content: []*yaml.Node{
									{Value: "discoverySources"},
									{
										Content: []*yaml.Node{
											{
												Content: []*yaml.Node{
													{Value: "gcp", Kind: yaml.ScalarNode},
													{Content: []*yaml.Node{
														{Value: "name"},
														{Value: "test"},
														{Value: "bucket"},
														{Value: "test-bucket"},
														{Value: "manifestPath"},
														{Value: "test-manifestPath"},
													}},
													{Value: "contextType", Kind: yaml.ScalarNode},
													{Value: "tmc", Kind: yaml.ScalarNode},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			out: &yaml.Node{
				Content: []*yaml.Node{
					{Value: "discoverySources"},
					{
						Content: []*yaml.Node{
							{
								Content: []*yaml.Node{
									{Value: "gcp", Kind: yaml.ScalarNode},
									{Content: []*yaml.Node{
										{Value: "name"},
										{Value: "test"},
										{Value: "bucket"},
										{Value: "test-bucket"},
										{Value: "manifestPath"},
										{Value: "test-manifestPath"},
									}},
									{Value: "contextType", Kind: yaml.ScalarNode},
									{Value: "tmc", Kind: yaml.ScalarNode},
								},
							},
						},
					},
				},
			},
			opts: func(config *CfgNode) {
				config.ForceCreate = true
				config.Keys =
					[]Key{
						{Name: "clientOptions", Type: yaml.MappingNode},
						{Name: "cli", Type: yaml.MappingNode},
					}
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			n := FindNode(spec.in, spec.opts)
			if spec.nullValue {
				assert.Nil(t, n)
			} else {
				assert.Equal(t, spec.out.Content, n.Content)
			}
		})
	}
}
