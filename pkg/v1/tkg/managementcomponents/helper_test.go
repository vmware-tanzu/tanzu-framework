// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/managementcomponents"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
)

var _ = Describe("Test GetTKGPackageConfigValuesFileFromUserConfig", func() {
	var (
		managementPackageVersion string
		userProviderConfigValues map[string]interface{}
		tkgBomConfigData         string
		tkgBomConfig             *tkgconfigbom.BOMConfiguration
		providerType             string
		valuesFile               string
		outputFile               string
		err                      error
	)

	tkgBomConfigData = `apiVersion: run.tanzu.vmware.com/v1alpha2
default:
  k8sVersion: v1.23.5+vmware.1-tkg.1-fake
release:
  version: v1.6.0-fake
imageConfig:
  imageRepository: fake.custom.repo
tkr-bom:
  imagePath: tkr-bom
tkr-compatibility:
  imagePath: fake-path/tkr-compatibility
tkr-package-repo:
  aws: tkr-repository-aws
  azure: tkr-repository-azure
  vsphere-nonparavirt: tkr-repository-vsphere-nonparavirt
tkr-package:
  aws: tkr-aws
  azure: tkr-azure
  vsphere-nonparavirt: tkr-vsphere-nonparavirt
`

	JustBeforeEach(func() {
		// Configure tkgBoMConfig
		tkgBomConfig = &tkgconfigbom.BOMConfiguration{}
		err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
		Expect(err).NotTo(HaveOccurred())

		// Configure user provider configuration
		userProviderConfigValues = map[string]interface{}{
			"PROVIDER_TYPE": providerType,
		}

		// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing
		valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, userProviderConfigValues, tkgBomConfig)
	})

	Context("When provider type is AWS", func() {
		BeforeEach(func() {
			managementPackageVersion = "v0.21.0"
			providerType = "aws"
			outputFile = "test/output_aws.yaml"
		})
		It("should not return error", func() {
			Expect(err).NotTo(HaveOccurred())
			f1, err := os.ReadFile(valuesFile)
			Expect(err).NotTo(HaveOccurred())
			f2, err := os.ReadFile(outputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(f1)).To(Equal(string(f2)))
		})
	})

	Context("When provider type is vSphere", func() {
		BeforeEach(func() {
			managementPackageVersion = "v0.21.0"
			providerType = "vsphere"
			outputFile = "test/output_vsphere.yaml"
		})
		It("should not return error", func() {
			Expect(err).NotTo(HaveOccurred())
			f1, err := os.ReadFile(valuesFile)
			Expect(err).NotTo(HaveOccurred())
			f2, err := os.ReadFile(outputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(f1)).To(Equal(string(f2)))
		})
	})

	Context("When provider type is Azure", func() {
		BeforeEach(func() {
			managementPackageVersion = "v0.21.0"
			providerType = "azure"
			outputFile = "test/output_azure.yaml"
		})
		It("should not return error", func() {
			Expect(err).NotTo(HaveOccurred())
			f1, err := os.ReadFile(valuesFile)
			Expect(err).NotTo(HaveOccurred())
			f2, err := os.ReadFile(outputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(f1)).To(Equal(string(f2)))
		})
	})

	Context("When provider type is not provided", func() {
		BeforeEach(func() {
			managementPackageVersion = "v0.21.0"
			providerType = ""
		})
		It("should not return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown provider type"))
		})
	})
})
