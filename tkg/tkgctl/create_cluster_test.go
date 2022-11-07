// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
)

const fakeTKRVersion = "1.19.0+vmware.1-tkg.1"
const configFilePath = "../fakes/config/config.yaml"

const inputFileAws = "../fakes/config/cluster_aws.yaml"
const inputFileAwsIncorrectClass = "../fakes/config/cluster_aws_incorrectClass.yaml"
const inputFileAwsEmptyClass = "../fakes/config/cluster_aws_emptyClass.yaml"
const inputFileMultipleObjectsAws = "../fakes/config/cluster_aws_multipleObjects.yaml"
const inputFileAzure = "../fakes/config/cluster_azure.yaml"
const inputFileVsphere = "../fakes/config/cluster_vsphere.yaml"
const inputFileTKGSClusterClass = "../fakes/config/cluster_tkgs.yaml"
const inputFileTKGSTKC = "../fakes/config/cluster_tkgs_tkc.yaml"
const inputFileLegacy = "../fakes/config/cluster1_config.yaml"
const errFeatureStatus = "error while checking feature status in featuregate"

var testingDir string

var _ = Describe("Unit tests for create cluster", func() {
	var (
		options   CreateClusterOptions
		tkgClient *fakes.Client
		fg        *fakes.FakeFeatureGateHelper
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
			fg = &fakes.FakeFeatureGateHelper{}
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
				featureGateHelper:      fg,
			}
			fg.FeatureActivatedInNamespaceReturns(true, nil)

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
				featureGateHelper:      fg,
			}
			fg.FeatureActivatedInNamespaceReturns(true, nil)
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

var _ = Describe("Unit tests for (AWS)  cluster_aws.yaml as input file for 'tanzu cluster create -f cluster_aws.yaml' use case", func() {
	var (
		ctl           tkgctl
		tkgClient     = &fakes.Client{}
		bomClient     = &fakes.TKGConfigBomClient{}
		options       CreateClusterOptions
		isTKGSCluster = false
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
	Context("When input Cluster object file is valid, plan is devcc:", func() {
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            "test-cluster",
				Plan:                   "devcc",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
				ClusterConfigFile:      inputFileAws,
			}
		})
		It("Environment should be updated with legacy variables and CreateClusterOptions updated with Cluster attribute values:", func() {
			// Process input cluster yaml file, this should process input cluster yaml file
			// and update the environment with legacy name and values
			// most of cluster yaml attributes are mapped to legacy variable for more look this - constants.ClusterToLegacyVariablesMapAws
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			// process input file and get map, which has all attribute's path and its values
			inputVariablesMap := getInputAttributesMap(options.ClusterConfigFile)

			// updates lower precedence variables with higher precedence variables values.
			// In AWS use case, there were few attributes repeated in cluster yaml file, we need to take always higher precedence values.
			updateLowerPrecedenceVariablesWithHigherPrecedenceVariablesValues(inputVariablesMap)

			// validate input Cluster Object yaml file attributes values with corresponding legacy variable values in environment, both should be same, as we have already updated the environment with Cluster Object attribute values.
			validateFileInputAttributeValuesWithEnvironmentValues(&ctl, inputVariablesMap, constants.ClusterAttributesToLegacyVariablesMapAws)

			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("aws-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// checking manually for some variables mapping values
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("default").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// Check the values in environment, should be updated from the cluster class input file
			Expect("aws-workload-cluster1").To(Equal(options.ClusterName))
			Expect("default").To(Equal(options.Namespace))
			Expect("aws").To(Equal(options.InfrastructureProvider))
			Expect(1).To(Equal(options.ControlPlaneMachineCount)) // mapped to CONTROL_PLANE_MACHINE_COUNT

			// check value for "spec.clusterNetwork.services.cidrBlocks"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
			Expect("2002::1234:abcd:ffff:c0a8:101/64,100.64.0.0/18").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for  TKG_IP_FAMILY which is decided based on "spec.clusterNetwork.services.cidrBlocks" and "spec.clusterNetwork.pod.cidrBlocks"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.TKGIPFamily)
			Expect(constants.DualStackPrimaryIPv6Family).To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.proxy.httpsProxy": TKGHTTPSProxy
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
			Expect("http://10.0.200.100").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.network.subnets.0.public.cidr":  ConfigVariableAWSPublicNodeCIDR
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPublicNodeCIDR)
			Expect("10.1.1.0/24").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.network.subnets.2.private.id":   ConfigVariableAWSPrivateSubnetID2
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateSubnetID2)
			Expect("idValuePrivate2").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.network.vpc.existingID":         ConfigVariableAWSVPCID
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCID)
			Expect("vpcID11").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.network.securityGroupOverrides.node":         ConfigVariableAWSSecurityGroupNode
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSSecurityGroupNode)
			Expect("securitygroupNode").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.trust.imageRepository" : ConfigVariableCustomImageRepositoryCaCertificate
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepositoryCaCertificate)
			Expect("trust.imageRepository.val").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.controlPlane.rootVolume.sizeGiB": ConfigVariableAWSControlplaneOsDiskSizeGib
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSControlplaneOsDiskSizeGib)
			Expect("80").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for spec.topology.workers.machineDeployments.0.replicas : ConfigVariableWorkerMachineCount0 - WORKER_MACHINE_COUNT_0
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount0)
			Expect("1").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// "spec.topology.workers.machineDeployments.2.variables.overrides.worker.instanceType": ConfigVariableNodeMachineType2, // NODE_MACHINE_TYPE_2
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType2)
			Expect("worker2").To(Equal(fmt.Sprintf("%v", mappedVal)))
		})
		It("When Input file is cluster type with multiple objects, Environment should be updated with legacy variables and CreateClusterOptions also updated with Cluster attribute values:", func() {

			// Process input cluster yaml file, this should process input cluster yaml file
			// and update the environment with legacy name and values
			// most of cluster yaml attributes are mapped to legacy variable for more look this - constants.ClusterToLegacyVariablesMapAws
			options.ClusterConfigFile = inputFileMultipleObjectsAws
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			// process input file and get map, which has all attribute's path and its values
			inputVariablesMap := getInputAttributesMap(options.ClusterConfigFile)

			// updates lower precedence variables with higher precedence variables values.
			// In AWS use case, there were few attributes repeated in cluster yaml file, we need to take always higher precedence values.
			updateLowerPrecedenceVariablesWithHigherPrecedenceVariablesValues(inputVariablesMap)

			// validate input Cluster Object yaml file attributes values with corresponding legacy variable values in environment, both should be same, as we have already updated the environment with Cluster Object attribute values.
			validateFileInputAttributeValuesWithEnvironmentValues(&ctl, inputVariablesMap, constants.ClusterAttributesToLegacyVariablesMapAws)

			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("aws-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))
		})

		It("When Input file is cluster type does not have value for spec.topology.class, return an error", func() {

			// Process input cluster.yaml file, this should process input cluster.yaml file
			// and update the environment with legacy name and values
			// most of cluster.yaml attributes are mapped to legacy variable for more look this - constants.clusterToLegacyVariablesMapAws
			options.ClusterConfigFile = inputFileAwsEmptyClass
			_, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(fmt.Sprint(err)).To(Equal(constants.ClusterResourceWithoutTopologyNotSupportedErrMsg))
		})

		It("When Input file is aws clusterclass.yaml file, but in-correct spec.topology.class name:", func() {
			options.ClusterConfigFile = inputFileAwsIncorrectClass
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(fmt.Sprint(err)).To(Equal(constants.TopologyClassIncorrectValueErrMsg))
		})

		It("When Input file is config.yaml file not cluster.yaml file", func() {
			options.ClusterConfigFile = inputFileLegacy
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeFalse())
			Expect(err).To(BeNil())
		})

		It("return error when Input file is not specified", func() {
			options.ClusterConfigFile = ""
			IsInputFileClusterClassBased, _ := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeFalse())
		})

		It("return error when Input file not exists", func() {
			options.ClusterConfigFile = "NOT-EXISTS"
			_, err := ctl.processWorkloadClusterInputFile(&options, false)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Unit tests for - (Vsphere) - cluster_vsphere.yaml as input file for 'tanzu cluster create -f cluster_vsphere.yaml' use case", func() {
	var (
		ctl           tkgctl
		tkgClient     = &fakes.Client{}
		bomClient     = &fakes.TKGConfigBomClient{}
		options       CreateClusterOptions
		isTKGSCluster = false
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
	Context("When input file is valid Cluster Class, plan devcc:", func() {
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            "test-cluster",
				Plan:                   "devcc",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
				ClusterConfigFile:      inputFileVsphere,
			}
		})
		It("Environment should be updated with legacy variables with input cluster attribute values:", func() {

			// Process input cluster.yaml file, this should process input cluster.yaml file
			// and update the environment with legacy name and values
			// most of cluster.yaml attributes are mapped to legacy variable for more look this - constants.ClusterToLegacyVariablesMapVsphere
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			// process input file and get map, which has all attribute's path and its values
			inputVariablesMap := getInputAttributesMap(options.ClusterConfigFile)

			// validate input Cluster Object yaml file attributes values with corresponding legacy variable values in environment, both should be same, as we have already updated the environment with Cluster Object attribute values.
			validateFileInputAttributeValuesWithEnvironmentValues(&ctl, inputVariablesMap, constants.ClusterAttributesToLegacyVariablesMapVsphere)

			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("vsphere-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))

			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("namespace-test1").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.clusterNetwork.pods.cidrBlocks":     ConfigVariableClusterCIDR
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterCIDR)
			Expect("10.10.10.10/18").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.controlPlane.replicas": ConfigVariableControlPlaneMachineCount,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneMachineCount)
			Expect("5").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.proxy.httpsProxy": TKGHTTPSProxy
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
			Expect("http://10.0.200.100").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.imageRepository.tlsCertificateValidation": ConfigVariableCustomImageRepositorySkipTLSVerify
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepositorySkipTLSVerify)
			Expect("true").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.trust.proxy":           TKGProxyCACert  TKG_PROXY_CA_CERT
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.TKGProxyCACert)
			Expect("LS0tLS=").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.apiServerPort": ConfigVariableClusterAPIServerPort  CLUSTER_API_SERVER_PORT
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterAPIServerPort)
			Expect("443").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.variables.apiServerEndpoint":      ConfigVariableVsphereControlPlaneEndpoint, VSPHERE_CONTROL_PLANE_ENDPOINT
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
			Expect("http://10.0.200.101").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// check that cluster options also updated
			Expect("http://10.0.200.101").To(Equal(options.VsphereControlPlaneEndpoint))

			// check value for "spec.topology.variables.controlPlane.network.nameservers"
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneNodeNameservers)
			Expect("100.64.0.0").To(Equal(fmt.Sprintf("%v", mappedVal)))

			//  check value for "spec.topology.variables.controlPlane.machine.numCPUs
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPNumCpus)
			Expect("2").To(Equal(fmt.Sprintf("%v", mappedVal)))

			//  check value for "spec.topology.variables.worker.machine.memoryMiB" ConfigVariableVsphereWorkerMemMib
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerMemMib)
			Expect("16384").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.workers.machineDeployments.2.failureDomain"  - ConfigVariableVsphereAz2
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereAz2)
			Expect("us-east-1c").To(Equal(fmt.Sprintf("%v", mappedVal)))

		})
	})
})

var _ = Describe("Unit tests for - (Azure) - cluster_azure.yaml as input file for 'tanzu cluster create -f cluster_azure.yaml' use case", func() {
	var (
		ctl           tkgctl
		tkgClient     = &fakes.Client{}
		bomClient     = &fakes.TKGConfigBomClient{}
		options       CreateClusterOptions
		isTKGSCluster = false
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
	Context("When input file is valid Cluster Class, plan devcc:", func() {
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
				ClusterConfigFile:      inputFileAzure,
			}
		})
		It("Environment should be updated with legacy variables with input cluster attribute values:", func() {

			// Process input cluster yaml file, this should process input cluster yaml file
			// and update the environment with legacy name and values
			// most of cluster yaml attributes are mapped to legacy variable for more look this - constants.ClusterToLegacyVariablesMapAzure
			IsInputFileClusterClassBased, err := ctl.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			// process input file and get map, which has all attribute's path and its values
			inputVariablesMap := getInputAttributesMap(options.ClusterConfigFile)

			// validate input Cluster Object yaml file attributes values with corresponding legacy variables values in environment, both should be same, as we have already updated the environment with Cluster Object attribute values.
			validateFileInputAttributeValuesWithEnvironmentValues(&ctl, inputVariablesMap, constants.ClusterAttributesToLegacyVariablesMapAzure)

			// checking manually for some variables mapping values
			mappedVal, _ := ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("azure-workload-cluster1").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// checking manually for some variables mapping values
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("namespace-test1").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// Check that cluster options also updated
			Expect("namespace-test1").To(Equal(options.Namespace))
			Expect("azure").To(Equal(options.InfrastructureProvider))

			//  check value for "spec.clusterNetwork.services.cidrBlocks": ConfigVariableServiceCIDR,
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
			Expect("10.10.10.10/16").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// check value for "spec.topology.workers.machineDeployments.2.failureDomain"  - ConfigVariableAzureAZ2
			mappedVal, _ = ctl.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureAZ2)
			Expect("us-east-1c").To(Equal(fmt.Sprintf("%v", mappedVal)))
		})
	})
})

var _ = Describe("TKGS Cluster - cluster_tkgs.yaml as input file for 'tanzu cluster create -f cluster_tkgs.yaml' use case", func() {
	var (
		tkgctlClient  *tkgctl
		tkgClient     = &fakes.Client{}
		bomClient     = &fakes.TKGConfigBomClient{}
		options       CreateClusterOptions
		isTKGSCluster = true
		fg            *fakes.FakeFeatureGateHelper
	)
	JustBeforeEach(func() {
		fg = &fakes.FakeFeatureGateHelper{}
		tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
		tkgctlClient = &tkgctl{
			configDir:              testingDir,
			tkgClient:              tkgClient,
			tkgConfigReaderWriter:  tkgConfigReaderWriter,
			tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
			tkgBomClient:           bomClient,
			featureGateHelper:      fg,
		}
		tkgClient.IsFeatureActivatedReturns(true)
	})
	Context("When input file is valid Cluster Class", func() {
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
				ClusterConfigFile:      inputFileTKGSClusterClass,
			}
		})
		It("Environment should be updated with legacy variables with input cluster attribute values:", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			// Process input cluster yaml file, its being tkgs cluster use case, there would be no variable mapping takes place
			// only the namespace and cluster name values read from input file and set in env and options
			IsInputFileClusterClassBased, err := tkgctlClient.processWorkloadClusterInputFile(&options, isTKGSCluster)
			Expect(IsInputFileClusterClassBased).Should(BeTrue())
			Expect(err).To(BeNil())

			// checking manually for some variables mapping values
			mappedVal, _ := tkgctlClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
			Expect("cc01").To(Equal(fmt.Sprintf("%v", mappedVal)))

			// checking manually for some variables mapping values
			mappedVal, _ = tkgctlClient.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
			Expect("ns01").To(Equal(fmt.Sprintf("%v", mappedVal)))
			// Check that cluster options also updated
			Expect("ns01").To(Equal(options.Namespace))
			Expect("cc01").To(Equal(options.ClusterName))
		})
	})
})
var _ = Describe("Clusterclass FeatureGate specific use cases", func() {
	var (
		options       CreateClusterOptions
		tkgClient     *fakes.Client
		regionContext region.RegionContext
		tkgctlClient  *tkgctl
		fg            *fakes.FakeFeatureGateHelper
	)

	Context("TKGS cluster class based cluster creation", func() {
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            "test-cluster",
				Plan:                   "devcc",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
				ClusterConfigFile:      inputFileTKGSClusterClass,
			}
			fg = &fakes.FakeFeatureGateHelper{}
			kubeConfigPath := inputFileTKGSClusterClass
			regionContext = region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: inputFileTKGSClusterClass,
			}
			tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			tkgConfigReaderWriter.Set(constants.ConfigVariableClusterPlan, "dev")
			tkgClient = &fakes.Client{}
			tkgctlClient = &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
				featureGateHelper:      fg,
			}
		})
		It("When feature flag (FeatureFlagPackageBasedLCM) enabled, input Cluster file is processed:", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			_ = tkgctlClient.CreateCluster(options)
			// Make sure call completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(1).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(2).To(Equal(pc))
			// Make sure its ClusterClass use case.
			cname, _ := tkgctlClient.tkgConfigReaderWriter.Get("CLUSTER_NAME")
			Expect(cname).To(Equal("cc01"))
			ns, _ := tkgctlClient.tkgConfigReaderWriter.Get(constants.ConfigVariableNamespace)
			Expect(ns).To(Equal("ns01"))
		})
		It("Expect error when feature flag (FeatureFlagPackageBasedLCM) not enabled and CC feature is disabled but CClass input file and TKGS Cluster ", func() {
			fg.FeatureActivatedInNamespaceReturns(false, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(false)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			err := tkgctlClient.CreateCluster(options)
			Expect(err.Error()).To(ContainSubstring("vSphere with Tanzu environment detected, however, the feature 'vmware-system-tkg-clusterclass' is not activated in 'vmware-system-tkg' namespace"))
			// Make sure call not completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(0).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(1).To(Equal(pc))
		})
		It("Should be able to create cluster when feature flag (FeatureFlagPackageBasedLCM) disabled and CC feature is enabled on supervisor cluster but CClass input file and TKGS Cluster ", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(false)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			err := tkgctlClient.CreateCluster(options)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Return error when Feature constants.CCFeature is disabled in featuregate", func() {
			fg.FeatureActivatedInNamespaceReturns(false, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			// feature flag (FeatureFlagPackageBasedLCM) activated, its clusterclass input file, but "clusterclass" feature in FeatureGate is disabled, so throws error
			err := tkgctlClient.CreateCluster(options)
			expectedErrMsg := fmt.Sprintf(constants.ErrorMsgFeatureGateNotActivated, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
			Expect(err.Error()).To(ContainSubstring(expectedErrMsg))
		})

		It("Return error when featuregate api throws error", func() {
			errorMsg := errFeatureStatus
			fg.FeatureActivatedInNamespaceReturns(true, fmt.Errorf(errorMsg))
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			// but throws error for the FeatureGate api, so we expect error here.
			err := tkgctlClient.CreateCluster(options)
			errorMsg = fmt.Sprintf(constants.ErrorMsgFeatureGateStatus, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
			// as FeatureGate api throws error, we expect error.
			Expect(err.Error()).To(ContainSubstring(errorMsg))
		})
	})
	Context("TKGS TKC based cluster creation", func() {
		var (
			clustername = "tkc-01"
			namespace   = "ns01"
		)
		BeforeEach(func() {
			options = CreateClusterOptions{
				ClusterName:            clustername,
				Plan:                   "devcc",
				InfrastructureProvider: "",
				Namespace:              "",
				GenerateOnly:           false,
				TkrVersion:             fakeTKRVersion,
				SkipPrompt:             true,
				Edition:                "tkg",
				ClusterConfigFile:      inputFileTKGSTKC,
			}
			fg = &fakes.FakeFeatureGateHelper{}
			kubeConfigPath := inputFileTKGSTKC
			regionContext = region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: inputFileTKGSTKC,
			}
			tkgConfigReaderWriter, _ := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFilePath, configFilePath)
			tkgConfigReaderWriter.Set(constants.ConfigVariableClusterPlan, "dev")
			tkgClient = &fakes.Client{}
			tkgctlClient = &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
				featureGateHelper:      fg,
			}
		})
		It("When feature flag (FeatureFlagPackageBasedLCM) enabled, input TKC file is processed:", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			_ = tkgctlClient.CreateCluster(options)
			// Make sure call completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(1).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(2).To(Equal(pc))
			// Make sure its ClusterClass use case.
			cname, _ := tkgctlClient.tkgConfigReaderWriter.Get("CLUSTER_NAME")
			Expect(cname).To(Equal(clustername))
			ns, _ := tkgctlClient.tkgConfigReaderWriter.Get(constants.ConfigVariableNamespace)
			Expect(ns).To(Equal(namespace))
		})
		It("Expect to complete CreateCluster call even feature flag (FeatureFlagPackageBasedLCM) not enabled, TKC input file ", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(false)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			_ = tkgctlClient.CreateCluster(options)

			// Make sure call not completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(1).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(2).To(Equal(pc))

			cname, _ := tkgctlClient.tkgConfigReaderWriter.Get(constants.ConfigVariableClusterName)
			Expect(cname).To(Equal(clustername))
			ns, _ := tkgctlClient.tkgConfigReaderWriter.Get(constants.ConfigVariableNamespace)
			Expect(ns).To(Equal(namespace))
		})
		It("Return error when Feature constants.TKCAPIFeature is disabled in TKGS featuregate", func() {
			fg.FeatureActivatedInNamespaceReturns(false, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			// feature flag (FeatureFlagPackageBasedLCM) activated, its clusterclass input file, but "clusterclass" feature in FeatureGate is disabled, so throws error
			err := tkgctlClient.CreateCluster(options)
			expectedErrMsg := fmt.Sprintf(constants.ErrorMsgFeatureGateNotActivated, constants.TKCAPIFeature, constants.TKGSTKCAPINamespace)
			Expect(err.Error()).To(ContainSubstring(expectedErrMsg))
		})

		It("create cluster even when tkc-api featuregate api throws error", func() {
			errorMsg := errFeatureStatus
			fg.FeatureActivatedInNamespaceReturns(true, fmt.Errorf(errorMsg))
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			// but throws error for the FeatureGate api, so we expect error here.
			_ = tkgctlClient.CreateCluster(options)
			// Make sure call not completed till end
			c := tkgClient.CreateClusterCallCount()
			Expect(1).To(Equal(c))
			// Make sure its TKGs system.
			pc := tkgClient.IsPacificManagementClusterCallCount()
			Expect(2).To(Equal(pc))

			cname, _ := tkgctlClient.tkgConfigReaderWriter.Get(constants.ConfigVariableClusterName)
			Expect(cname).To(Equal(clustername))
			ns, _ := tkgctlClient.tkgConfigReaderWriter.Get(constants.ConfigVariableNamespace)
			Expect(ns).To(Equal(namespace))
		})
		It("return error when cluster already exists:", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: clustername, Namespace: namespace}, {Name: "my-cluster-2", Namespace: "my-system"}}, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)
			tkgClient.CreateClusterReturnsOnCall(0, false, nil)

			err := tkgctlClient.CreateCluster(options)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(constants.ErrorMsgClusterExistsAlready, clustername)))
		})
		It("return error when list cluster returns error:", func() {
			fg.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.IsPacificManagementClusterReturnsOnCall(0, true, nil)
			tkgClient.ListTKGClustersReturns(nil, errors.New("failed to list clusters"))
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgClient.IsFeatureActivatedReturns(true)

			err := tkgctlClient.CreateCluster(options)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(constants.ErrorMsgClusterListError))
		})
	})
})

// getInputAttributesMap process input Cluster Object yaml file and returns a map which has all attributes from the input file and its values
func getInputAttributesMap(inputFile string) map[string]interface{} {
	// to make sure input cluster attributes values are updated with legacy variables in environment (tkgctl.TKGConfigReaderWriter())
	// Processing the input cluster yaml file with the existing util api, and expecting Cluster YAML object,
	// from the Cluster YAML object, map attributes path with values, and update Map variablesMap.
	_, clusterObj, _ := CheckIfInputFileIsClusterClassBased(inputFile)
	inputVariablesMap := make(map[string]interface{})
	inputVariablesMap["metadata.name"] = clusterObj.GetName()
	inputVariablesMap["metadata.namespace"] = clusterObj.GetNamespace()
	spec := clusterObj.Object[constants.SPEC].(map[string]interface{})
	err := processYamlObjectAndAddToMap(spec, constants.SPEC, inputVariablesMap)
	Expect(err).To(BeNil())
	return inputVariablesMap
}

// validateFileInputAttributeValuesWithEnvironmentValues takes the Cluster Object yaml file input variable map and checks its values in the environment, both should be same.
func validateFileInputAttributeValuesWithEnvironmentValues(ctl *tkgctl, inputVariablesMap map[string]interface{}, clusterAttributesToLegacyVariablesMap map[string]string) {
	// take  legacy variable name for every attribute in cluster input file,
	// then check the value of legacy variable in environment, with value of the same from the input file
	// both should match.
	for key := range inputVariablesMap {
		if inputValue := inputVariablesMap[key]; inputValue != nil {
			legacyNameForClusterObjectInputVariable, ok := clusterAttributesToLegacyVariablesMap[key]
			if ok && legacyNameForClusterObjectInputVariable != "" {
				mappedVal, _ := ctl.TKGConfigReaderWriter().Get(legacyNameForClusterObjectInputVariable)
				Expect(fmt.Sprintf("%v", mappedVal)).To(Equal(fmt.Sprintf("%v", inputValue)))
			}
		}
	}
}

// updateLowerPrecedenceVariablesWithHigherPrecedenceVariablesValues, updates lower precedence variables with higher precedence variables values.
func updateLowerPrecedenceVariablesWithHigherPrecedenceVariablesValues(inputVariablesMap map[string]interface{}) {
	// below logic, takes higher precedence attribute path and its lower precedence path,
	// check if lower precedence path has value if so overrides with higher precedence path value.
	for higherPrecedenceKey := range constants.ClusterAttributesHigherPrecedenceToLowerMap {
		_, ok1 := constants.ClusterAttributesToLegacyVariablesMapAws[higherPrecedenceKey]
		value, ok2 := inputVariablesMap[higherPrecedenceKey]
		if ok1 && ok2 {
			lowerPrecedenceAttribute := constants.ClusterAttributesHigherPrecedenceToLowerMap[higherPrecedenceKey]
			// lower precedence attribute value should be the value of higher
			inputVariablesMap[lowerPrecedenceAttribute] = fmt.Sprintf("%v", value)
		}
	}
}

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
