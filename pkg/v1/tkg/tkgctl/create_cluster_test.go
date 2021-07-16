// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
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
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("../fakes/config/config.yaml", "../fakes/config/config.yaml")
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
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("../fakes/config/config.yaml", "../fakes/config/config.yaml")
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
		It("InfrastructureProvider is windows-vsphere when GenerateOnly is true", func() {
			kubeConfigPath := getConfigFilePath()
			regionContext := region.RegionContext{
				ContextName:    "queen-anne-context",
				SourceFilePath: kubeConfigPath,
			}
			tkgClient = &fakes.Client{}
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			tkgClient.GetCurrentRegionContextReturns(regionContext, nil)
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("../fakes/config/config.yaml", "../fakes/config/config.yaml")
			Expect(err).NotTo(HaveOccurred())
			tkgctlClient := &tkgctl{
				configDir:              testingDir,
				tkgClient:              tkgClient,
				kubeconfig:             kubeConfigPath,
				tkgConfigReaderWriter:  tkgConfigReaderWriter,
				tkgConfigUpdaterClient: tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter),
			}

			options.InfrastructureProvider = constants.InfrastructureProviderWindowsVSphere
			options.GenerateOnly = true
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
		tkgConfigReaderWriter, err1 := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("../fakes/config/config.yaml", "../fakes/config/config.yaml")
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
