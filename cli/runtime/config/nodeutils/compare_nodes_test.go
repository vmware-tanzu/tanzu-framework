// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	cliapi "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

func TestCompareNodes(t *testing.T) {
	tests := []struct {
		name  string
		node1 *yaml.Node
		node2 *yaml.Node
		out   bool
	}{
		{
			name: "should return false when Equal method is called",
			node1: &yaml.Node{
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
										Value: string(cliapi.TargetK8s),
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "test-mc-merge",
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
												Value: "target",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: string(cliapi.TargetK8s),
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
			node2: &yaml.Node{
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
										Value: string(cliapi.TargetK8s),
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "test-mc-merge",
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
												Value: "target",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: string(cliapi.TargetK8s),
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
			out: false,
		},
		{
			name: "should return true when NotEqual method is called",
			node1: &yaml.Node{
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
										Value: string(cliapi.TargetK8s),
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "test-mc-merge",
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
												Value: "target",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: string(cliapi.TargetK8s),
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
			node2: &yaml.Node{
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
										Value: string(cliapi.TargetK8s),
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "test-mc",
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
												Value: "target",
											},
											{
												Kind:  yaml.ScalarNode,
												Value: string(cliapi.TargetK8s),
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
			out: true,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			ok, err := NotEqual(spec.node1, spec.node2)
			assert.NoError(t, err)
			assert.Equal(t, ok, spec.out)
		})
	}
}
