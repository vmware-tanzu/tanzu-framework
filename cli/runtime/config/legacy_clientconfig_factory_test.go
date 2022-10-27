package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestStoreClientConfig(t *testing.T) {
	// setup
	func() {
		LocalDirName = fmt.Sprintf(".tanzu-test")
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name    string
		node    *yaml.Node
		cfg     *configapi.ClientConfig
		current bool
	}{
		{
			name: "should store client config",
			node: &yaml.Node{
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
			cfg: &configapi.ClientConfig{
				KnownServers: []*configapi.Server{
					{
						Name: "test-mc",
						Type: configapi.ManagementClusterServerType,
						ManagementClusterOpts: &configapi.ManagementClusterServer{
							Endpoint: "test-endpoint",
							Context:  "test-context",
							Path:     "test-path",
						},
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "test-bucket",
									ManifestPath: "test-manifest-path",
								},
								ContextType: configapi.CtxTypeTMC,
							},
						},
					},
				},
				CurrentServer: "test-mc",
				KnownContexts: []*configapi.Context{
					{
						Name: "test-mc",
						Type: "k8s",
						ClusterOpts: &configapi.ClusterServer{
							Endpoint:            "test-context-endpoint",
							Path:                "test-context-path",
							Context:             "test-context",
							IsManagementCluster: true,
						},
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "ctx-test-bucket",
									ManifestPath: "ctx-test-manifest-path",
								},
								ContextType: configapi.CtxTypeTMC,
							},
						},
					},
				},
				CurrentContext: map[configapi.ContextType]string{
					configapi.CtxTypeK8s: "test-mc",
				},
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						Repositories: []configapi.PluginRepository{
							{
								GCPPluginRepository: &configapi.GCPPluginRepository{
									Name:       "test",
									BucketName: "bucket",
									RootPath:   "root-path",
								},
							},
						},
						DiscoverySources: []configapi.PluginDiscovery{
							{
								GCP: &configapi.GCPDiscovery{
									Name:         "test",
									Bucket:       "ctx-test-bucket",
									ManifestPath: "ctx-test-manifest-path",
								},
								ContextType: configapi.CtxTypeTMC,
							},
						},
						UnstableVersionSelector: configapi.VersionSelectorLevel("unstable-version"),
						Edition:                 configapi.EditionSelector("test=tkg"),
						BOMRepo:                 "test-bomrepo",
						CompatibilityFilePath:   "test-compatibility-file-path",
					},
					Features: map[string]configapi.FeatureMap{
						"global": {
							"tkr-version-v1alpha3-beta":     "false",
							"context-aware-cli-for-plugins": "true",
							"context-target":                "false",
						},
						"cluster": {
							"custom-nameservers":      "false",
							"dual-stack-ipv4-primary": "false",
							"dual-stack-ipv6-primary": "false",
						},
					},
					Env: map[string]string{
						"dev":  "dev",
						"user": "vmw",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := PersistNode(tc.node)
			assert.NoError(t, err)

			err = StoreClientConfig(tc.cfg)
			assert.NoError(t, err)

		})
	}

}
