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

const cclassInputFileAws = "../fakes/config/cluster_aws.yaml"
const cclusterInputFileMultipleObjectsAws = "../fakes/config/cluster_aws_multipleObjects.yaml"
const cclassInputFileAzure = "../fakes/config/cluster_azure.yaml"
const cclassInputFileVsphere = "../fakes/config/cluster_vsphere.yaml"

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

var _ = Describe("Unit tests for (AWS)  ccluster.yaml as input file for 'tanzu cluster create -f cluster_aws.yaml' use case", func() {
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
	Context("create cluster with ccluster_aws.yaml, AWS provider, plan devcc:", func() {
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
				ClusterConfigFile:      cclassInputFileAws,
			}
		})
		It("create cluster with ccluster_aws.yaml, AWS provider, plan devcc, make sure legacy configuration are updated:", func() {
			//Process input ccluster.yaml file.
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			_, cclusterObj, _ := ctl.checkIfInputFileIsClusterClassBased(options.ClusterConfigFile)
			variablesMap := make(map[string]interface{})
			variablesMap["metadata.name"] = cclusterObj.GetName()
			variablesMap["metadata.namespace"] = cclusterObj.GetNamespace()
			spec := cclusterObj.Object[constants.SPEC].(map[string]interface{})
			err = processYamlObjectAndAddToMap(spec, constants.SPEC, variablesMap)
			Expect(err).To(BeNil())

			// override legacy variables with higher precedence values if exists
			for higherPrecedenceKey := range constants.CclusterVariablesWithHigherPrecedence {
				_, ok1 := constants.CclusterToLegacyVariablesMapAws[higherPrecedenceKey]
				value, ok2 := variablesMap[higherPrecedenceKey]
				if ok1 && ok2 {
					lowerPrecedenceAttribute := constants.CclusterVariablesWithHigherPrecedence[higherPrecedenceKey]
					// lower precedence attribute value should be the value of higher
					variablesMap[lowerPrecedenceAttribute] = fmt.Sprintf("%v", value)
				}
			}

			for key := range variablesMap {
				if inputValue := variablesMap[key]; inputValue != nil {
					var legacyNameForClusterObjectInputVariable string
					var ok bool
					if _, ok = constants.CclusterToLegacyVariablesMapCommon[key]; ok {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapCommon[key]
					} else {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapAws[key]
					}
					if ok && legacyNameForClusterObjectInputVariable != "" {
						mappedVal, _ := ctl.TKGConfigReaderWriter().Get(legacyNameForClusterObjectInputVariable)
						Expect(fmt.Sprintf("%v", mappedVal)).To(Equal(fmt.Sprintf("%v", inputValue)))
					}
				}
			}
			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("aws-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// checking manually for some variables mapping values
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("namespace-test1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.variables.subnets.2.public.cidr"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPublicNodeCIDR2)
			Expect("10.0.5.0/24").To(Equal(fmt.Sprintf("%v", mappedVal)))

			//check value for "spec.topology.variables.nodes.0.osDisk.sizeGiB"
			//mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeOsDiskSizeGib)
			//Expect("80").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.workers.machineDeployments.2.replicas"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount2)
			Expect("3").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.workers.machineDeployments.1.variables.overrides.NODE_MACHINE_TYPE"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType1)
			Expect("m6.xlarge").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.proxy" is set
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.TKGHTTPProxyEnabled)
			Expect("true").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.variables.proxy" is set
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy)
			Expect("http://10.0.200.100").To(Equal(fmt.Sprintf("%v", mappedVal)))
		})

		It("Input file is ccluster type with multiple objects, AWS infra and plan is prodcc, make sure Cluster object processed and legacy configs are updated:", func() {

			// Process input cluster class file with multiple objects
			options.ClusterConfigFile = cclusterInputFileMultipleObjectsAws
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			_, cclusterObj, _ := ctl.checkIfInputFileIsClusterClassBased(options.ClusterConfigFile)
			variablesMap := make(map[string]interface{})
			variablesMap["metadata.name"] = cclusterObj.GetName()
			variablesMap["metadata.namespace"] = cclusterObj.GetNamespace()
			spec := cclusterObj.Object[constants.SPEC].(map[string]interface{})
			err = processYamlObjectAndAddToMap(spec, constants.SPEC, variablesMap)
			Expect(err).To(BeNil())

			// override legacy variables with higher precedence values if exists
			for higherPrecedenceKey := range constants.CclusterVariablesWithHigherPrecedence {
				_, ok1 := constants.CclusterToLegacyVariablesMapAws[higherPrecedenceKey]
				value, ok2 := variablesMap[higherPrecedenceKey]
				if ok1 && ok2 {
					lowerPrecedenceAttribute := constants.CclusterVariablesWithHigherPrecedence[higherPrecedenceKey]
					// lower precedence attribute value should be the value of higher
					variablesMap[lowerPrecedenceAttribute] = fmt.Sprintf("%v", value)
				}
			}

			for key := range variablesMap {
				if inputValue := variablesMap[key]; inputValue != nil {
					var legacyNameForClusterObjectInputVariable string
					var ok bool
					if _, ok = constants.CclusterToLegacyVariablesMapCommon[key]; ok {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapCommon[key]
					} else {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapAws[key]
					}
					if ok && legacyNameForClusterObjectInputVariable != "" {
						mappedVal, _ := ctl.TKGConfigReaderWriter().Get(legacyNameForClusterObjectInputVariable)
						Expect(fmt.Sprintf("%v", mappedVal)).To(Equal(fmt.Sprintf("%v", inputValue)))
					}
				}
			}
			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("aws-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// checking manually for some variables mapping values
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("namespace-test1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.variables.subnets.2.public.cidr"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPublicNodeCIDR2)
			Expect("10.0.5.0/24").To(Equal(fmt.Sprintf("%v", mappedVal)))

			//check value for "spec.topology.variables.nodes.0.osDisk.sizeGiB"
			//mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeOsDiskSizeGib)
			//Expect("80").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.workers.machineDeployments.2.replicas"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount2)
			Expect("3").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check value for "spec.topology.workers.machineDeployments.1.variables.overrides.NODE_MACHINE_TYPE"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType1)
			Expect("m6.xlarge").To(Equal(fmt.Sprintf("%v", mappedVal)))
		})

		It("Input file is config.yaml file not ccluster.yaml file", func() {
			options.ClusterConfigFile = "../fakes/config/ccluster1_config.yaml"
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeFalse())
			Expect(err).To(BeNil())
		})

		It("Input file is aws clusterclass.yaml file, but in-correct topology.class name:", func() {
			options.ClusterConfigFile = "../fakes/config/cluster_aws_incorrectClass.yaml"
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(fmt.Sprint(err)).To(Equal(constants.TopologyClassIncorrectValueErrMsg))
		})

		It("Input file is not specified", func() {
			options.ClusterConfigFile = ""
			IsInputFileClusterClassBased, _ := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeFalse())
		})

		It("Input file not exists", func() {
			options.ClusterConfigFile = "NOT-EXISTS"
			_, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Unit tests for - (Vsphere)- cluster_vsphere.yaml as input file for 'tanzu cluster create -f cluster_vsphere.yaml' use case", func() {
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
	Context("create cluster with clusterclass_vsphere.yaml, vsphere provider, plan devcc:", func() {
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
				ClusterConfigFile:      cclassInputFileVsphere,
			}
		})
		It("Input file is clusterclass_vsphere.yaml type,  vsphere provider, plan devcc, make sure legacy configurations are updated..", func() {
			//Process input cluster class file.
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			_, cclusterObj, _ := ctl.checkIfInputFileIsClusterClassBased(options.ClusterConfigFile)
			variablesMap := make(map[string]interface{})
			variablesMap["metadata.name"] = cclusterObj.GetName()
			variablesMap["metadata.namespace"] = cclusterObj.GetNamespace()
			spec := cclusterObj.Object[constants.SPEC].(map[string]interface{})
			err = processYamlObjectAndAddToMap(spec, constants.SPEC, variablesMap)
			Expect(err).To(BeNil())

			for key := range variablesMap {
				if inputValue := variablesMap[key]; inputValue != nil {
					var legacyNameForClusterObjectInputVariable string
					var ok bool
					if _, ok = constants.CclusterToLegacyVariablesMapCommon[key]; ok {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapCommon[key]
					} else {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapVsphere[key]
					}
					if ok && legacyNameForClusterObjectInputVariable != "" {
						mappedVal, _ := ctl.TKGConfigReaderWriter().Get(legacyNameForClusterObjectInputVariable)
						Expect(fmt.Sprintf("%v", mappedVal)).To(Equal(fmt.Sprintf("%v", inputValue)))
					}
				}
			}
			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("vsphere-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// checking manually for some variables mapping values
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("namespace-test1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// "spec.clusterNetwork.services.cidrBlocks": ConfigVariableServiceCIDR,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
			Expect("10.10.10.10/16").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// "spec.topology.variables.controlPlane.network.nameservers":       ConfigVariableControlPlaneNodeNameservers,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneNodeNameservers)
			Expect("3").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// "spec.topology.variables.controlPlane.machine.numCPUs : ConfigVariableVsphereCPNumCpus
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPNumCpus)
			Expect("2").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// "spec.topology.variables.node.machine.memoryMiB": ConfigVariableVsphereWorkerMemMib,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerMemMib)
			Expect("16384").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// "spec.topology.workers.machineDeployments.0.replicas": ConfigVariableWorkerMachineCount
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount)
			Expect("4").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// TopologyWorkersMachineDeploymentsFailureDomain2: ConfigVariableVsphereAz2
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereAz2)
			Expect("3").To(Equal(fmt.Sprintf("%v", mappedVal)))

		})
	})
})

var _ = Describe(" 'tanzu cluster create -f cluster_azure.yaml' use case - vsphere provider - devcc plan:", func() {
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
	Context("'tanzu cluster create -f cluster_azure.yaml' use case - vsphere provider - devcc plan:", func() {
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
				ClusterConfigFile:      cclassInputFileAzure,
			}
		})
		It("'tanzu cluster create -f cluster_azure.yaml' use case - vsphere provider - devcc plan:, make sure legacy configurations are updated..", func() {
			//Process input cluster class file.
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			_, cclusterObj, _ := ctl.checkIfInputFileIsClusterClassBased(options.ClusterConfigFile)
			variablesMap := make(map[string]interface{})
			variablesMap["metadata.name"] = cclusterObj.GetName()
			variablesMap["metadata.namespace"] = cclusterObj.GetNamespace()
			spec := cclusterObj.Object[constants.SPEC].(map[string]interface{})
			err = processYamlObjectAndAddToMap(spec, constants.SPEC, variablesMap)
			Expect(err).To(BeNil())

			for key := range variablesMap {
				if inputValue := variablesMap[key]; inputValue != nil {
					var legacyNameForClusterObjectInputVariable string
					var ok bool
					if _, ok = constants.CclusterToLegacyVariablesMapCommon[key]; ok {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapCommon[key]
					} else {
						legacyNameForClusterObjectInputVariable, ok = constants.CclusterToLegacyVariablesMapAzure[key]
					}
					if ok && legacyNameForClusterObjectInputVariable != "" {
						mappedVal, _ := ctl.TKGConfigReaderWriter().Get(legacyNameForClusterObjectInputVariable)
						Expect(fmt.Sprintf("%v", mappedVal)).To(Equal(fmt.Sprintf("%v", inputValue)))
					}
				}
			}

			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("azure-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// checking manually for some variables mapping values
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("namespace-test1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// "spec.clusterNetwork.services.cidrBlocks": ConfigVariableServiceCIDR,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
			Expect("10.10.10.10/16").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// "spec.topology.variables.controlPlane.subnet.securityGroup":       ConfigVariableAzureControlPlaneSubnetSecurityGroup,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureControlPlaneSubnetSecurityGroup)
			Expect("SecurityGroupCP").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// "spec.topology.variables.node.outboundLB.idleTimeoutInMinutes": ConfigVariableAzureNodeOutboundLbIdleTimeoutInMinutes,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureNodeOutboundLbIdleTimeoutInMinutes)
			Expect("8").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// TopologyWorkersMachineDeploymentsFailureDomain2: ConfigVariableAzureAZ2
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureAZ2)
			Expect("3").To(Equal(fmt.Sprintf("%v", mappedVal)))

		})
	})
})

var _ = Describe("Unit tests for feature flag (config.FeatureFlagPackageBasedLCM) and featureGate for clusterclass - TKGS : vSphere provider, devcc:", func() {
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
		It("positive case, feature flag (config.FeatureFlagPackageBasedLCM) enabled, clusterclass featuregate enabled, and its TKGS cluster : vSphere provider, devcc:", func() {
			kubeConfigPath := cclassInputFileVsphere
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: cclassInputFileVsphere,
			}

			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)
			options.ClusterConfigFile = cclassInputFileVsphere
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
			Expect(cname).To(Equal("vsphere-workload-cluster1"))
		})
		It("clusterclass feature flag 'config.FeatureFlagPackageBasedLCM' not enabled : vSphere provider, devcc:", func() {
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
			options.ClusterConfigFile = cclassInputFileAzure
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
			options.ClusterConfigFile = cclassInputFileAzure
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
			options.ClusterConfigFile = cclassInputFileAzure
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
