// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"

	"os"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
)

var _ = Describe("Validate", func() {
	var (
		tkgClient             *client.TkgClient
		tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		featureFlagClient     *fakes.FeatureFlagClient
	)
	BeforeEach(func() {
		tkgBomClient := new(fakes.TKGConfigBomClient)
		tkgBomClient.GetDefaultTkrBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{
			Release: &tkgconfigbom.ReleaseInfo{Version: "v1.3"},
			Components: map[string][]*tkgconfigbom.ComponentInfo{
				"kubernetes": {{Version: "v1.20"}},
			},
		}, nil)
		tkgBomClient.GetDefaultTkgBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{
			Release: &tkgconfigbom.ReleaseInfo{Version: "v1.23"},
		}, nil)

		configDir := os.TempDir()

		configFile, err := os.CreateTemp(configDir, "cluster-config-*.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(configFile.Sync()).To(Succeed())
		Expect(configFile.Close()).To(Succeed())

		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configFile.Name(), configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		readerWriter, err := tkgconfigreaderwriter.NewWithReaderWriter(tkgConfigReaderWriter)
		Expect(err).NotTo(HaveOccurred())

		tkgConfigUpdater := new(fakes.TKGConfigUpdaterClient)
		tkgConfigUpdater.CheckInfrastructureVersionStub = func(providerName string) (string, error) {
			return providerName, nil
		}

		featureFlagClient = &fakes.FeatureFlagClient{}
		featureFlagClient.IsConfigFeatureActivatedReturns(true, nil)

		options := client.Options{
			ReaderWriterConfigClient: readerWriter,
			TKGConfigUpdater:         tkgConfigUpdater,
			TKGBomClient:             tkgBomClient,
			RegionManager:            new(fakes.RegionManager),
			FeatureFlagClient:        featureFlagClient,
		}
		tkgClient, err = client.New(options)
		Expect(err).NotTo(HaveOccurred())
	})
	Context("Validate presence of Azure env variables during upgrade", func() {
		It("Azure client secret has not been set", func() {
			err := os.Unsetenv(constants.ConfigVariableAzureClientSecret)
			Expect(err).ToNot(HaveOccurred())

			err = tkgClient.ValidateEnvVariables(client.AzureProviderName)
			Expect(err).To(HaveOccurred())
		})
		It("Azure client secret has been set", func() {
			tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureClientSecret, "foo-bar")
			err := tkgClient.ValidateEnvVariables(client.AzureProviderName)
			Expect(err).NotTo(HaveOccurred())
		})
		It("IaaS is AWS. This is a no-op currently", func() {
			err := tkgClient.ValidateEnvVariables(client.AWSProviderName)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
