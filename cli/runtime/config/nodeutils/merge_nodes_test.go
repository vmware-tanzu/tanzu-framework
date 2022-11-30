// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMergeNode(t *testing.T) {
	tests := []struct {
		name string
		src  *yaml.Node
		dst  *yaml.Node
		opts Options
	}{
		{
			name: "success merge src into empty dst node",
			src: &yaml.Node{
				Kind: yaml.DocumentNode,
				Content: []*yaml.Node{
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "currentContext",
							},
							{
								Kind:  yaml.MappingNode,
								Value: "",
								Content: []*yaml.Node{
									{
										Kind:  yaml.ScalarNode,
										Value: "k8s",
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "test-mc-merge",
									},
								},
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "current",
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "test-mc",
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "servers",
							},
							{
								Kind:  yaml.SequenceNode,
								Value: "",
								Content: []*yaml.Node{
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{
												Kind:  yaml.ScalarNode,
												Value: "name",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "test-mc",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "type",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "managementcluster",
											},
										},
									},
								},
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "contexts",
							},
							{
								Kind:  yaml.SequenceNode,
								Value: "",
								Content: []*yaml.Node{
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{
												Kind:  yaml.ScalarNode,
												Value: "name",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "test-mc",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "type",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "k8s",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "clusterOpts",
											},
											{
												Kind:  yaml.MappingNode,
												Value: "",
												Content: []*yaml.Node{
													{
														Kind:  yaml.ScalarNode,
														Value: "isManagementCluster",
													},
													{
														Kind:  yaml.ScalarNode,
														Tag:   "!!bool",
														Value: "true",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "annotation",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "one",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "required",
													},
													{
														Kind:  yaml.ScalarNode,
														Tag:   "!!bool",
														Value: "true",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "annotationStruct",
													},
													{
														Kind:  yaml.MappingNode,
														Value: "",
														Content: []*yaml.Node{
															{
																Kind:  yaml.ScalarNode,
																Value: "one",
															},
															{
																Kind:  yaml.ScalarNode,
																Value: "one",
															},
														},
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "endpoint",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "test-endpoint",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "path",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "test-path",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "context",
													},
													{
														Kind:  yaml.ScalarNode,
														Value: "test-context",
													},
												},
											},
											{
												Kind:  yaml.ScalarNode,
												Value: "discoverySources",
											},
											{
												Kind:  yaml.SequenceNode,
												Value: "",
												Content: []*yaml.Node{
													{
														Kind: yaml.MappingNode,
														Content: []*yaml.Node{
															{
																Kind:  yaml.ScalarNode,
																Value: "gcp",
															},
															{
																Kind: yaml.MappingNode,
																Content: []*yaml.Node{

																	{
																		Kind:  yaml.ScalarNode,
																		Value: "name",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "test",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "bucket",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "test-bucket",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "manifestPath",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "test-manifest-path",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "annotation",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "one",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Value: "required",
																	},
																	{
																		Kind:  yaml.ScalarNode,
																		Tag:   "!!bool",
																		Value: "true",
																	},
																}},
															{
																Value: "contextType",
																Kind:  yaml.ScalarNode,
															},
															{
																Value: "tmc",
																Kind:  yaml.ScalarNode,
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
					},
				},
			},
			dst: &yaml.Node{
				Kind: yaml.DocumentNode,
				Content: []*yaml.Node{
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "currentContext",
							},
							{
								Kind:  yaml.MappingNode,
								Value: "",
								Content: []*yaml.Node{
									{
										Kind:  yaml.ScalarNode,
										Value: "k8s",
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "test-mc",
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
						{Name: "currentContext", Type: yaml.MappingNode},
					}
			},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			_, err := MergeNodes(spec.src, spec.dst)
			assert.NoError(t, err)
			node := FindNode(spec.dst.Content[0], spec.opts)
			assert.NotNil(t, node)
			assert.Equal(t, node, spec.dst.Content[0].Content[1])
		})
	}
}
