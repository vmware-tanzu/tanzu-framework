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
		name string
		in   *yaml.Node
		opts Options
		out  *yaml.Node
	}{
		{
			name: "success find nodeutils with existing nodes",
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
			opts: func(config *Config) {
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
			name: "success find nodeutils no  nodes",
			in:   &yaml.Node{},
			opts: func(config *Config) {
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
			name: "success find nodeutils level one",
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
			opts: func(config *Config) {
				config.ForceCreate = true
				config.Keys =
					[]Key{
						{Name: "clientOptions", Type: yaml.MappingNode},
					}
			},
		},
		{
			name: "success find nodeutils level two",
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
			opts: func(config *Config) {
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
			_, err := FindNode(spec.in, spec.opts)
			assert.NoError(t, err)
		})
	}
}
