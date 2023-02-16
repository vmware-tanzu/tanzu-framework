// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

type MockClusterctlClient struct {
	clusterctl.Client
}

func (m *MockClusterctlClient) GetClusterTemplate(options clusterctl.GetClusterTemplateOptions) (clusterctl.Template, error) {
	templateMapYaml := []byte("cluster.x-k8s.io/v1alpha3\n" +
		"kind: Cluster\n" +
		"metadata:\n" +
		fmt.Sprintf("  name: ${%s}\n", "foobar") +
		"spec:\n" +
		"  clusterNetwork:\n" +
		"    pods:\n" +
		"      cidrBlocks:\n" +
		"        - 192.168.0.0/16\n" +
		"  infrastructureRef:\n" +
		"    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3\n" +
		"    kind: vSphereCluster\n" +
		"    name: tkg-region-vSphere-11111111111111")

	targetNamespace := options.TargetNamespace
	return repository.NewTemplate(repository.TemplateInput{
		RawArtifact:           templateMapYaml,
		ConfigVariablesClient: nil,
		Processor:             yamlprocessor.NewSimpleProcessor(),
		TargetNamespace:       targetNamespace,
		SkipTemplateProcess:   true,
	})
}

var _ = Describe("ValidateManagementClusterVersionWithCLI", func() {
	const (
		clusterName = "test-cluster"
		v140        = "v1.4.0"
		v141        = "v1.4.1"
		v150        = "v1.5.0"
	)
	var (
		regionalClient fakes.ClusterClient
		tkgBomClient   fakes.TKGConfigBomClient
		regionManager  fakes.RegionManager
		c              *TkgClient
		err            error
	)
	JustBeforeEach(func() {
		err = c.ValidateManagementClusterVersionWithCLI(&regionalClient)
	})
	BeforeEach(func() {
		regionManager = fakes.RegionManager{}
		regionManager.GetCurrentContextReturns(region.RegionContext{
			ClusterName: clusterName,
			Status:      region.Success,
		}, nil)

		regionalClient = fakes.ClusterClient{}
		regionalClient.ListResourcesStub = func(i interface{}, lo ...client.ListOption) error {
			list := i.(*v1alpha3.ClusterList)
			*list = v1alpha3.ClusterList{
				Items: []v1alpha3.Cluster{
					{
						ObjectMeta: v1.ObjectMeta{
							Name:      clusterName,
							Namespace: "default",
						},
					},
				},
			}
			return nil
		}

		c, err = New(Options{
			TKGConfigUpdater: &fakes.TKGConfigUpdaterClient{},
			TKGBomClient:     &tkgBomClient,
			RegionManager:    &regionManager,
		})
	})
	Context("v1.4.0 management cluster", func() {
		BeforeEach(func() {
			regionalClient.GetManagementClusterTKGVersionReturns(v140, nil)
		})

		When("management cluster version matches cli version", func() {
			BeforeEach(func() {
				tkgBomClient = fakes.TKGConfigBomClient{}
				tkgBomClient.GetDefaultTKGReleaseVersionReturns(v140, nil)
			})
			It("should validate without error", func() {
				Expect(err).To(BeNil())
			})
		})

		When("cli version is a patch version ahead of management cluster", func() {
			BeforeEach(func() {
				tkgBomClient = fakes.TKGConfigBomClient{}
				tkgBomClient.GetDefaultTKGReleaseVersionReturns(v141, nil)
			})
			It("should validate without error", func() {
				Expect(err).To(BeNil())
			})
		})

		When("cli version is a minor version ahead of management cluster", func() {
			BeforeEach(func() {
				tkgBomClient = fakes.TKGConfigBomClient{}
				tkgBomClient.GetDefaultTKGReleaseVersionReturns(v150, nil)
			})
			It("should return an error", func() {
				Expect(err).Should(MatchError("version mismatch between management cluster and cli version. Please upgrade your management cluster to the latest to continue"))
			})
		})
	})

})

var _ = Describe("CreateCluster", func() {
	const (
		clusterName = "regional-cluster-2"
	)
	var (
		tkgClient                   *TkgClient
		clusterClientFactory        *fakes.ClusterClientFactory
		clusterClient               *fakes.ClusterClient
		featureFlagClient           *fakes.FeatureFlagClient
		tkgBomClient                *fakes.TKGConfigBomClient
		tkgConfigUpdaterClient      *fakes.TKGConfigUpdaterClient
		tkgConfigReaderWriter       *fakes.TKGConfigReaderWriter
		tkgConfigReaderWriterClient *fakes.TKGConfigReaderWriterClient
		vcClientFactory             *fakes.VcClientFactory
		vcClient                    *fakes.VCClient
		options                     CreateClusterOptions
		clusterCtlClient            *MockClusterctlClient
	)

	BeforeEach(func() {
		clusterClientFactory = &fakes.ClusterClientFactory{}
		clusterClient = &fakes.ClusterClient{}
		clusterClientFactory.NewClientReturns(clusterClient, nil)
		featureFlagClient = &fakes.FeatureFlagClient{}
		tkgBomClient = &fakes.TKGConfigBomClient{}
		tkgConfigUpdaterClient = &fakes.TKGConfigUpdaterClient{}
		tkgConfigReaderWriterClient = &fakes.TKGConfigReaderWriterClient{}
		tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
		vcClientFactory = &fakes.VcClientFactory{}
		vcClient = &fakes.VCClient{}
		clusterCtlClient = &MockClusterctlClient{}

		tkgConfigReaderWriterClient.TKGConfigReaderWriterReturns(tkgConfigReaderWriter)
		vcClientFactory.NewClientReturns(vcClient, nil)

		tkgClient, err = CreateTKGClientOptsMutator(configFile2, testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Second, func(o Options) Options {
			o.ClusterClientFactory = clusterClientFactory
			o.FeatureFlagClient = featureFlagClient
			o.TKGBomClient = tkgBomClient
			o.TKGConfigUpdater = tkgConfigUpdaterClient
			o.ReaderWriterConfigClient = tkgConfigReaderWriterClient
			o.VcClientFactory = vcClientFactory
			o.ClusterCtlClient = clusterCtlClient
			return o
		})
		Expect(err).NotTo(HaveOccurred())

		tkgBomConfigData := `
ova: []
`
		tkgBomConfig := &tkgconfigbom.BOMConfiguration{}
		err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
		Expect(err).NotTo(HaveOccurred())
		tkgBomClient.GetBOMConfigurationFromTkrVersionReturns(tkgBomConfig, nil)
		tkgBomClient.GetDefaultTkgBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{
			Release: &tkgconfigbom.ReleaseInfo{Version: "v1.23"},
		}, nil)

		clusterClient.GetManagementClusterTKGVersionReturns("v1.2.1-rc.1", nil)
		clusterClient.GetRegionalClusterDefaultProviderNameReturns(VSphereProviderName, nil)
		tkgBomClient.GetDefaultTKGReleaseVersionReturns("v1.2.1-rc.1", nil)
		tkgBomClient.GetDefaultTkrBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{
			Release: &tkgconfigbom.ReleaseInfo{Version: "v1.3"},
			Components: map[string][]*tkgconfigbom.ComponentInfo{
				"kubernetes": {{Version: "v1.18.0+vmware.2"}},
			},
		}, nil)
		clusterClient.ListResourcesCalls(func(clusterList interface{}, options ...client.ListOption) error {
			if clusterList, ok := clusterList.(*v1alpha3.ClusterList); ok {
				clusterList.Items = []v1alpha3.Cluster{
					{
						ObjectMeta: v1.ObjectMeta{
							Name:      clusterName,
							Namespace: constants.DefaultNamespace,
						},
					},
				}
				return nil
			}
			return nil
		})
		clusterClient.IsPacificRegionalClusterReturns(false, nil)

		tkgConfigReaderWriter.GetCalls(func(key string) (string, error) {
			configMap := populateConfigMap()
			if val, ok := configMap[key]; ok {
				return val, nil
			}
			return "192.168.2.1/16", nil
		})
	})

	Context("ValidateConfigForSingleNodeCluster", func() {
		When("Feature gate is enabled", func() {
			BeforeEach(func() {
				featureFlagClient.IsConfigFeatureActivatedStub = func(featureFlagName string) (bool, error) {
					if featureFlagName == constants.FeatureFlagSingleNodeClusters {
						return true, nil
					}
					return true, nil
				}
			})

			It("Should fail reading the cluster yaml", func() {
				options = createClusterOptions(clusterName, "../fakes/config/invalid_config.yaml")
				_, err := tkgClient.CreateCluster(&options, false)
				Expect(err.Error()).To(ContainSubstring("unable to read cluster yaml"))
			})

			It("Should fail if cluster is single node and controlPlaneTaint exists", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere_snc_cp_taint_true.yaml")
				_, err := tkgClient.CreateCluster(&options, false)

				Expect(err).To(MatchError(fmt.Sprintf("unable to create single node cluster %s as control plane node has taint", clusterName)))
				Expect(clusterClient.ApplyCallCount()).To(BeZero())
			})

			It("Should fail validation if control plane taint is invalid", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere_snc_invalid_cp_taint.yaml")
				_, err := tkgClient.CreateCluster(&options, false)

				Expect(err).To(MatchError("failed to get CC variable controlPlaneTaint: unmarshalling from JSON into value: json: cannot unmarshal string into Go value of type bool"))
				Expect(clusterClient.ApplyCallCount()).To(BeZero())
			})

			It("Should fail if cluster is single node with workers nil and controlPlaneTaint are set", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere_snc_omit_workers.yaml")
				_, err := tkgClient.CreateCluster(&options, false)
				Expect(err).To(MatchError(fmt.Sprintf("unable to create single node cluster %s as control plane node has taint", clusterName)))
				Expect(clusterClient.ApplyCallCount()).To(BeZero())
			})

			It("Should successfully create a single node cluster", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere_snc.yaml")
				_, err := tkgClient.CreateCluster(&options, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterClient.ApplyCallCount()).To(Equal(1))
			})

			It("Should successfully create a multi node cluster", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere.yaml")
				_, err := tkgClient.CreateCluster(&options, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterClient.ApplyCallCount()).To(Equal(1))
			})
		})
		When("Feature gate is disabled", func() {
			BeforeEach(func() {
				featureFlagClient.IsConfigFeatureActivatedStub = func(featureFlagName string) (bool, error) {
					if featureFlagName == constants.FeatureFlagSingleNodeClusters {
						return false, nil
					}
					return true, nil
				}
			})
			It("Should fail if cluster is single node", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere_snc.yaml")
				_, err := tkgClient.CreateCluster(&options, false)

				Expect(err).To(MatchError("Worker count cannot be 0, minimum worker count required is 1"))
				Expect(clusterClient.ApplyCallCount()).To(BeZero())
			})
			It("Should successfully create a multi node cluster", func() {
				options = createClusterOptions(clusterName, "../fakes/config/cluster_vsphere.yaml")
				_, err := tkgClient.CreateCluster(&options, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterClient.ApplyCallCount()).To(Equal(1))
			})

		})
	})

	Context("ValidateFeatureFlagAllowLegacyClusterWhenCreateCluster", func() {
		When("Feature gate is enabled", func() {
			BeforeEach(func() {
				featureFlagClient.IsConfigFeatureActivatedStub = func(featureFlagName string) (bool, error) {
					if featureFlagName == constants.FeatureFlagForceDeployClusterWithClusterClass {
						return false, nil
					}
					return true, nil
				}
				tkgConfigUpdaterClient.GetProvidersChecksumStub = func() (string, error) {
					return "FakeFileShaIsSame", nil
				}
				tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFileStub = func() (string, error) {
					return "FakeFileShaIsSame", nil
				}
				tkgConfigReaderWriter.GetCalls(func(key string) (string, error) {
					configMap := populateConfigMap()
					if val, ok := configMap[key]; ok {
						return val, nil
					}
					if key == constants.ConfigVariableAllowLegacyCluster {
						return "true", nil
					}
					return "192.168.2.1/16", nil
				})
			})

			It("Should go to create a legacy based workload cluster directly", func() {
				options = createClusterOptions(clusterName, "../fakes/config/config.yaml")
				options.IsInputFileClusterClassBased = false
				Expect(err).ToNot(HaveOccurred())

				result := captureOutput(*tkgClient, options, false)
				Expect(result).To(ContainSubstring(constants.YTTBasedClusterWarning))
				Expect(result).To(ContainSubstring("creating workload cluster"))
			})

		})

		When("Feature gate is disabled", func() {
			BeforeEach(func() {
				tkgConfigUpdaterClient.GetProvidersChecksumStub = func() (string, error) {
					return "FakeFileShaIsSame", nil
				}
				tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFileStub = func() (string, error) {
					return "FakeFileShaIsSame", nil
				}
				tkgConfigReaderWriter.GetCalls(func(key string) (string, error) {
					configMap := populateConfigMap()
					if val, ok := configMap[key]; ok {
						return val, nil
					}
					if key == constants.ConfigVariableAllowLegacyCluster {
						return "false", nil
					}
					return "192.168.2.1/16", nil
				})
			})

			It("Should go to create a classy based workload cluster directly", func() {
				featureFlagClient.IsConfigFeatureActivatedStub = func(featureFlagName string) (bool, error) {
					if featureFlagName == constants.FeatureFlagAllowLegacyCluster {
						return false, nil
					}
					return true, nil
				}
				options = createClusterOptions(clusterName, "../fakes/config/config.yaml")
				options.IsInputFileClusterClassBased = false
				Expect(err).ToNot(HaveOccurred())

				result := captureOutput(*tkgClient, options, false)
				Expect(result).To(ContainSubstring("Legacy configuration file detected. The inputs from said file have been converted into the new Cluster configuration"))
				Expect(result).To(ContainSubstring(fmt.Sprintf("Using this new Cluster configuration '%v", testingDir)))
				Expect(result).To(ContainSubstring("creating workload cluster"))
			})

			It("Should return create a classy based workload cluster command", func() {
				featureFlagClient.IsConfigFeatureActivatedStub = func(featureFlagName string) (bool, error) {
					if featureFlagName == constants.FeatureFlagAllowLegacyCluster || featureFlagName == constants.FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration {
						return false, nil
					}
					return true, nil
				}
				options = createClusterOptions(clusterName, "../fakes/config/config.yaml")
				options.IsInputFileClusterClassBased = false
				Expect(err).ToNot(HaveOccurred())

				result := captureOutput(*tkgClient, options, false)
				Expect(result).To(ContainSubstring("Legacy configuration file detected. The inputs from said file have been converted into the new Cluster configuration"))
				Expect(result).To(ContainSubstring("To create a cluster with it, use"))
			})

		})

		When("ALLOW_LEGACY_CLUSTER is failed to set", func() {
			BeforeEach(func() {
				tkgConfigReaderWriter.GetCalls(func(key string) (string, error) {
					configMap := populateConfigMap()
					if val, ok := configMap[key]; ok {
						return val, nil
					}
					if key == constants.ConfigVariableAllowLegacyCluster {
						return "false", errors.Errorf("fake error")
					}
					return "192.168.2.1/16", nil
				})

				featureFlagClient.IsConfigFeatureActivatedStub = func(featureFlagName string) (bool, error) {
					if featureFlagName == constants.FeatureFlagAllowLegacyCluster {
						return false, nil
					}
					return true, nil
				}

				tkgConfigUpdaterClient.GetProvidersChecksumStub = func() (string, error) {
					return "FakeFileShaIsSame", nil
				}
				tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFileStub = func() (string, error) {
					return "FakeFileShaIsSame", nil
				}
			})

			It("Should return an error", func() {
				options = createClusterOptions(clusterName, "../fakes/config/config.yaml")
				options.IsInputFileClusterClassBased = false
				Expect(err).ToNot(HaveOccurred())
				result := captureOutput(*tkgClient, options, false)
				Expect(result).To(ContainSubstring("failed getting variable ALLOW_LEGACY_CLUSTER"))
			})
		})
	})

})

// captureOutput will capture logs and warnings
func captureOutput(tkgClient TkgClient, options CreateClusterOptions, isManagementCluster bool) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// stdout := os.Stdout
	// os.Stdout = w
	// defer func() {
	// 	os.Stdout = stdout
	// }()

	stderr := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = stderr
	}()

	_, _ = tkgClient.CreateCluster(&options, isManagementCluster)
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func populateConfigMap() map[string]string {
	configMap := make(map[string]string, 0)
	configMap[constants.ConfigVariableCNI] = "antrea"
	configMap[constants.ConfigVariableControlPlaneNodeNameservers] = "8.8.8.8"
	configMap[constants.ConfigVariableWorkerNodeNameservers] = "8.8.8.8"
	configMap[VsphereNodeCPUVarName[0]] = "2"
	configMap[VsphereNodeCPUVarName[1]] = "2"
	configMap[VsphereNodeMemVarName[0]] = "4098"
	configMap[VsphereNodeMemVarName[1]] = "4098"
	configMap[VsphereNodeDiskVarName[0]] = "20"
	configMap[VsphereNodeDiskVarName[1]] = "20"
	configMap[constants.ConfigVariableVsphereServer] = "10.0.0.1"
	configMap[constants.ConfigVariableWorkerMachineCount0] = "0"
	configMap[constants.ConfigVariableWorkerMachineCount1] = "0"
	configMap[constants.ConfigVariableWorkerMachineCount2] = "0"
	configMap[constants.ConfigVariableEtcdExtraArgs] = ""
	configMap[constants.ConfigVariableAPIServerExtraArgs] = ""
	configMap[constants.ConfigVariableKubeSchedulerExtraArgs] = ""
	configMap[constants.ConfigVariableKubeControllerManagerExtraArgs] = ""
	configMap[constants.ConfigVariableControlPlaneKubeletExtraArgs] = ""
	configMap[constants.ConfigVariableWorkerKubeletExtraArgs] = ""
	return configMap
}

func createClusterOptions(clusterName, configFile string) CreateClusterOptions {
	options := CreateClusterOptions{
		ClusterConfigOptions: ClusterConfigOptions{
			KubernetesVersion: "v1.18.0+vmware.2",
			ClusterName:       clusterName,
			TargetNamespace:   constants.DefaultNamespace,
			ProviderRepositorySource: &clusterctl.ProviderRepositorySourceOptions{
				InfrastructureProvider: VSphereProviderName,
				Flavor:                 "dev",
			},
			WorkerMachineCount: pointer.Int64Ptr(0),
		},
		IsInputFileClusterClassBased: true,
		ClusterConfigFile:            configFile,
	}
	options.Edition = "some edition"
	return options
}
