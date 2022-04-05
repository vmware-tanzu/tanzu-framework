// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigupdater"
)

const fakeTKRVersion = "1.19.0+vmware.1-tkg.1"
const configFilePath = "../fakes/config/config.yaml"
const ccConfigFilePath = "../fakes/config/ccluster1_clusterOnly.yaml"

var testingDir string

var _ = Describe("Unit tests for create cluster", func() {
	var (
		options   CreateClusterOptions
		tkgClient *fakes.Client
	)

	BeforeSuite(createTempDirectory)
	AfterSuite(deleteTempDirectory)

	Context("Creating clusters for TKGs", func() {
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            "test-cluster",
				Plan:                   "dev",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
			}
		})
		It("Namespace is taken from the context when no -n flag is specified", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}

			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			Expect(err).NotTo(HaveOccurred())
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
			}

			err = tkgctlClient.CreateCluster(options)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Namespace is taken from the flag when -n is specified", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}
			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			Expect(err).NotTo(HaveOccurred())
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
			}

			options.Namespace = "custom-namespace"
			err = tkgctlClient.CreateCluster(options)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Unit tests for getAndDownloadTkrIfNeeded", func() {
	var (
		tkrVersion       string
		ctl              tkgctl
		tkgClient        = &fakes.Client{}
		bomClient        = &fakes.TKGConfigBomClient{}
		resultTKRVersion string
		resultK8SVersion string
		err              error
	)
	JustBeforeEach(func() {
		tkgConfigReaderWriter, err1 := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
		Expect(err1).NotTo(HaveOccurred())
		ctl = tkgctl{
			configDir:              testingDir,
			tkgClient:              tkgClient,
			tkgConfigReaderWriter:  tkgConfigReaderWriter,
			tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
			tkgBomClient:           bomClient,
		}
		resultTKRVersion, resultK8SVersion, err = ctl.getAndDownloadTkrIfNeeded(tkrVersion)
	})
	Context("When tkrVersion is not provided, and default tkr does not exist", func() {
		BeforeEach(func() {
			tkrVersion = ""
			bomClient.GetDefaultTkrBOMConfigurationReturns(nil, errors.New("bom not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(resultTKRVersion).To(Equal(""))
			Expect(resultK8SVersion).To(Equal(""))
		})
	})

	Context("When tkrVersion is not provided, and cannot get default k8s version", func() {
		BeforeEach(func() {
			tkrVersion = ""
			bomClient.GetDefaultTkrBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{Release: &tkgconfigbom.ReleaseInfo{Version: fakeTKRVersion}}, nil)
			bomClient.GetDefaultK8sVersionReturns("", errors.New("cannot get default k8s version"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(resultTKRVersion).To(Equal(""))
			Expect(resultK8SVersion).To(Equal(""))
		})
	})
	Context("When tkrVersion is not provided, and default tkr and k8s version can be found", func() {
		BeforeEach(func() {
			tkrVersion = ""
			bomClient.GetDefaultTkrBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{Release: &tkgconfigbom.ReleaseInfo{Version: fakeTKRVersion}}, nil)
			bomClient.GetDefaultK8sVersionReturns("1.19.0", nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resultTKRVersion).To(Equal(fakeTKRVersion))
			Expect(resultK8SVersion).To(Equal("1.19.0"))
		})
	})

	Context("When tkrVersion is provided and bom presents locally", func() {
		BeforeEach(func() {
			tkrVersion = fakeTKRVersion
			bomClient.GetBOMConfigurationFromTkrVersionReturns(nil, nil)
			bomClient.GetK8sVersionFromTkrVersionReturns("1.19.0", nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resultTKRVersion).To(Equal(fakeTKRVersion))
			Expect(resultK8SVersion).To(Equal("1.19.0"))
		})
	})

	Context("When tkrVersion is provided but local bom is mal-formated", func() {
		BeforeEach(func() {
			tkrVersion = fakeTKRVersion
			bomClient.GetBOMConfigurationFromTkrVersionReturns(nil, nil)
			bomClient.GetK8sVersionFromTkrVersionReturns("", errors.New("failed to get k8s version"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When valid tkrVersion is provided, but bom presents locally", func() {
		BeforeEach(func() {
			tkrVersion = fakeTKRVersion
			bomClient.GetBOMConfigurationFromTkrVersionReturns(nil, tkgconfigbom.BomNotPresent{})
			tkgClient.DownloadBomFileReturns(nil)
			bomClient.GetK8sVersionFromTkrVersionReturns("1.19.0", nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resultTKRVersion).To(Equal(fakeTKRVersion))
			Expect(resultK8SVersion).To(Equal("1.19.0"))
		})
	})
	Context("When invalid tkrVersion is provided, and bom presents locally", func() {
		BeforeEach(func() {
			tkrVersion = fakeTKRVersion
			bomClient.GetBOMConfigurationFromTkrVersionReturns(nil, tkgconfigbom.BomNotPresent{})
			tkgClient.DownloadBomFileReturns(errors.New("configmap not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Unit tests for - ccluster.yaml as input file for 'tanzu cluster create -f ccluster' use case", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		bomClient = &fakes.TKGConfigBomClient{}
		options   CreateClusterOptions
	)
	JustBeforeEach(func() {
		tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
		ctl = tkgctl{
			configDir:              testingDir,
			tkgClient:              tkgClient,
			tkgConfigReaderWriter:  tkgConfigReaderWriter,
			tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
			tkgBomClient:           bomClient,
		}
		tkgClient.IsFeatureActivatedReturns(true)
	})
	Context("create cluster with ccluster.yaml : use cases", func() {
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            "test-cluster",
				Plan:                   "dev",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
				ClusterConfigFile:      ccConfigFilePath,
			}
		})
		It("Input file is ccluster type, make sure configurations are updated..", func() {
			//Set some values before processing the input ccluster.yaml file.
			ctl.TKGConfigReaderWriter().Set("CLUSTER_NAME", "BeforeCheckingInputFile")
			vpcID := "VPC_ID_BeforeCheckingInputFile_11"
			ctl.TKGConfigReaderWriter().Set("AWS_VPC_ID", vpcID)

			//Process input ccluster.yaml file.
			IsInputFileHasCClass, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileHasCClass).Should(BeTrue())
			Expect(err).To(BeNil())

			cname, _ := ctl.TKGConfigReaderWriter().Get("CLUSTER_NAME")
			//cname should be same as CLUSTER_NAME value from the input config file options.ClusterConfigFile
			// though we set before checking but it should override by the input ccluster.yaml file
			Expect(cname).To(Equal("wcc2"))
			vpcCIDR, _ := ctl.TKGConfigReaderWriter().Get("AWS_VPC_CIDR")
			Expect(vpcCIDR).To(Equal("10.0.0.0/16"))
			nodeType, _ := ctl.TKGConfigReaderWriter().Get("NODE_MACHINE_TYPE")
			Expect(nodeType).To(Equal("m3.xlarge"))
			tls, _ := ctl.TKGConfigReaderWriter().Get("TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY")
			Expect(tls).To(Equal("true"))
			//AWS_VPC_ID should be the same value what we set before processing, as this value is empty in ccluster.yaml file
			vpcIDAfterProcess, _ := ctl.TKGConfigReaderWriter().Get("AWS_VPC_ID")
			Expect(vpcIDAfterProcess).To(Equal(vpcID))
		})
		It("Input file is ccluster type with multiple objects.", func() {
			//Set some values before processing the input ccluster.yaml file.
			ctl.TKGConfigReaderWriter().Set("CLUSTER_NAME", "BeforeCheckingInputFile")
			vpcID := "VPC_ID_BeforeCheckingInputFile_22"
			ctl.TKGConfigReaderWriter().Set("AWS_VPC_ID", vpcID)

			//Process input ccluster.yaml file.
			options.ClusterConfigFile = "../fakes/config/ccluster2_multipleObjects.yaml"
			IsInputFileHasCClass, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileHasCClass).Should(BeTrue())
			Expect(err).To(BeNil())

			cname, _ := ctl.TKGConfigReaderWriter().Get("CLUSTER_NAME")
			//cname should be same as CLUSTER_NAME value from the input config file options.ClusterConfigFile
			// though we set before checking but it should override by the input ccluster.yaml file
			Expect(cname).To(Equal("wcc11"))
			vpcCIDR, _ := ctl.TKGConfigReaderWriter().Get("AWS_VPC_CIDR")
			Expect(vpcCIDR).To(Equal("10.0.0.0/16"))
			nodeType, _ := ctl.TKGConfigReaderWriter().Get("NODE_MACHINE_TYPE")
			Expect(nodeType).To(Equal("m3.xlarge"))
			tls, _ := ctl.TKGConfigReaderWriter().Get("TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY")
			Expect(tls).To(Equal("false"))
			repository, _ := ctl.TKGConfigReaderWriter().Get("TKG_CUSTOM_IMAGE_REPOSITORY")
			Expect(repository).To(BeEmpty())
			//AWS_VPC_ID should be the same value what we set before processing, as this value is empty in ccluster.yaml file
			vpcIDAfterProcess, _ := ctl.TKGConfigReaderWriter().Get("AWS_VPC_ID")
			Expect(vpcIDAfterProcess).To(Equal(vpcID))
		})

		It("make sure CreateClusterOptions values are updated.", func() {
			options.ClusterName = "BeforeProcess"
			options.Plan = "Plan"
			//Process input ccluster.yaml file.
			Expect(options.ClusterName).To(Equal("BeforeProcess"))
			Expect(options.Plan).To(Equal("Plan"))

			options.ClusterConfigFile = ccConfigFilePath
			_, _ = ctl.processWorkloadClusterInputFile(&options)

			cname, _ := ctl.TKGConfigReaderWriter().Get("CLUSTER_NAME")
			Expect(cname).To(Equal("wcc2"))
			Expect(options.ClusterName).To(Equal("wcc2"))
			Expect(options.Plan).To(Equal("devcc"))
		})

		It("Input file is config.yaml file not ccluster.yaml file", func() {
			options.ClusterConfigFile = "../fakes/config/ccluster1_config.yaml"
			IsInputFileHasCClass, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileHasCClass).Should(BeFalse())
			Expect(err).To(BeNil())
		})

		It("Input file is not specified", func() {
			options.ClusterConfigFile = ""
			IsInputFileHasCClass, _ := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileHasCClass).Should(BeFalse())
		})

		It("Input file not exists", func() {
			options.ClusterConfigFile = "NOT-EXISTS"
			_, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Unit tests for feature flag (config.FeatureFlagPackageBasedLCM) and featureGate for clusterclass - TKGS ", func() {
	var (
		options   CreateClusterOptions
		tkgClient *fakes.Client
	)

	Context("Creating clusters for TKGs", func() {
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            "test-cluster",
				Plan:                   "dev",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
			}
		})
		It("positive case, feature flag (config.FeatureFlagPackageBasedLCM) enabled, clusterclass featuregate enabled, and its TKGS cluster", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}

			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)
			options.ClusterConfigFile = ccConfigFilePath
			tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			fg := &fakes.FakeFeatureGateHelper{}
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
				featureGateHelper:      fg,
			}
			_ = tkgctlClient.CreateCluster(options)
			// Make sure call completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(1).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(1).To(Equal(pc))
			// Make sure its ClusterClass use case.
			cname, _ := tkgctlClient.tkgConfigReaderWriter.Get("CLUSTER_NAME")
			Expect(cname).To(Equal("wcc2"))
		})
		It("clusterclass feature flag 'config.FeatureFlagPackageBasedLCM' not enabled", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}

			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(false)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)
			options.ClusterConfigFile = ccConfigFilePath
			tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			fg := &fakes.FakeFeatureGateHelper{}
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
				featureGateHelper:      fg,
			}
			_ = tkgctlClient.CreateCluster(options)
			// Make sure call completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(1).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(1).To(Equal(pc))
			// As feature flag (config.FeatureFlagPackageBasedLCM) is not enabled, the input ccluster1_clusterOnly.yaml file not processed,
			// so cname is empty only.
			cname, _ := tkgctlClient.tkgConfigReaderWriter.Get("CLUSTER_NAME")
			Expect(cname).To(Equal(""))
		})
		It("Feature 'clusterclass' is disabled in featuregate", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}

			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)
			options.ClusterConfigFile = ccConfigFilePath
			tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			fg := &fakes.FakeFeatureGateHelper{}
			fg.FeatureActivatedInNamespaceReturns(false, nil)
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
				featureGateHelper:      fg,
			}
			// feature flag (config.FeatureFlagPackageBasedLCM) activated, its clusterclass input file, but "clusterclass" feature in FeatureGate is disabled, so throws error
			err := tkgctlClient.CreateCluster(options)
			expectedErrMsg := "vSphere with Tanzu environment detected, however, the feature 'clusterclass' is not activated in 'vmware-system-capw' namespace "
			Expect(err.Error()).To(ContainSubstring(expectedErrMsg))
		})

		It("featuregate api throws error", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}

			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)
			options.ClusterConfigFile = ccConfigFilePath
			tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			fg := &fakes.FakeFeatureGateHelper{}
			errorMsg := "error while feature status in featuregate"
			fg.FeatureActivatedInNamespaceReturns(true, fmt.Errorf(errorMsg))
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
				featureGateHelper:      fg,
			}
			// feature flag (config.FeatureFlagPackageBasedLCM) activated, its clusterclass config input file, but "clusterclass" feature in FeatureGate is enabled,
			// but throws errro for the FeatureGate api, so we expecte error here.
			err := tkgctlClient.CreateCluster(options)
			fmt.Println(err.Error())
			// as FeatureGate api throws error, we expecte error.
			Expect(err.Error()).To(ContainSubstring(errorMsg))
		})
	})
})

func getConfigFilePath() string {
	filename := "config1.yaml"
	filePath := "../fakes/config/kubeconfig/" + filename
	f, _ := os.CreateTemp(testingDir, "kube")
	copyFile(filePath, f.Name())
	return f.Name()
}

func copyFile(sourceFile, destFile string) {
	input, _ := os.ReadFile(sourceFile)
	_ = os.WriteFile(destFile, input, constants.ConfigFilePermissions)
}

func createTempDirectory() {
	testingDir, _ = os.MkdirTemp("", "cluster_client_test")
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}
