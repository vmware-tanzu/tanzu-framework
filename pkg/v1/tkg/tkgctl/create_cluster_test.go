// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/region"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigupdater"
)

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
				TkrVersion:             "1.19.0+vmware.1-tkg.1",
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
	})
})

func getConfigFilePath() string {
	filename := "config1.yaml"
	filePath := "../fakes/config/kubeconfig/" + filename
	f, _ := ioutil.TempFile(testingDir, "kube")
	copyFile(filePath, f.Name())
	return f.Name()
}

func copyFile(sourceFile, destFile string) {
	input, _ := ioutil.ReadFile(sourceFile)
	_ = ioutil.WriteFile(destFile, input, constants.ConfigFilePermissions)
}

func createTempDirectory() {
	testingDir, _ = ioutil.TempDir("", "cluster_client_test")
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}
