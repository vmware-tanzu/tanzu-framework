// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/managementcomponents"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
)

const (
	verStr = "v0.21.0"
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

		// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
		valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil)
	})

	Context("When provider type is AWS", func() {
		BeforeEach(func() {
			managementPackageVersion = verStr
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
			managementPackageVersion = verStr
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
			managementPackageVersion = verStr
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

	Context("When provider type is Docker", func() {
		BeforeEach(func() {
			managementPackageVersion = verStr
			providerType = "docker"
			outputFile = "test/output_docker.yaml"
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
			managementPackageVersion = verStr
			providerType = ""
		})
		It("should not return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown provider type"))
		})
	})
})

var _ = Describe("Test Set proxy settings", func() {
	var (
		managementPackageVersion string
		userProviderConfigValues map[string]interface{}
		tkgBomConfigData         string
		tkgBomConfig             *tkgconfigbom.BOMConfiguration
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
	// Configure user provider configuration
	userProviderConfigValues = map[string]interface{}{
		"TKG_HTTP_PROXY":    "http://192.168.116.1:3128",
		"TKG_HTTPS_PROXY":   "http://192.168.116.1:3128",
		"TKG_NO_PROXY":      ".svc,100.64.0.0/13,192.168.118.0/24,192.168.119.0/24,192.168.120.0/24",
		"PROVIDER_TYPE":     "vsphere",
		"TKG_PROXY_CA_CERT": "dGVzdDE=",
		"TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE": "fake-ca",
	}

	JustBeforeEach(func() {
		// Configure tkgBoMConfig
		tkgBomConfig = &tkgconfigbom.BOMConfiguration{}
		err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
		Expect(err).NotTo(HaveOccurred())

		// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
		valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil)
	})

	Context("when proxy is set", func() {
		BeforeEach(func() {
			managementPackageVersion = verStr
			outputFile = "test/output_vsphere_with_proxy.yaml"
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
})

var _ = Describe("Test Set custom ca settings", func() {
	var (
		managementPackageVersion string
		userProviderConfigValues map[string]interface{}
		tkgBomConfigData         string
		tkgBomConfig             *tkgconfigbom.BOMConfiguration
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
	// Configure user provider configuration
	userProviderConfigValues = map[string]interface{}{
		"TKG_HTTP_PROXY":              "http://192.168.116.1:3128",
		"TKG_HTTPS_PROXY":             "http://192.168.116.1:3128",
		"TKG_NO_PROXY":                ".svc,100.64.0.0/13,192.168.118.0/24,192.168.119.0/24,192.168.120.0/24",
		"PROVIDER_TYPE":               "vsphere",
		"TKG_CUSTOM_IMAGE_REPOSITORY": "fake-repo",
		"TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE": "fake-ca",
	}

	Context("when proxy is set from userconfig map", func() {
		BeforeEach(func() {
			managementPackageVersion = verStr
			outputFile = "test/output_vsphere_with_custom_repo_ca.yaml"
		})

		JustBeforeEach(func() {
			// Configure tkgBoMConfig
			tkgBomConfig = &tkgconfigbom.BOMConfiguration{}
			err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
			Expect(err).NotTo(HaveOccurred())

			// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
			valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil)
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

	Context("when proxy is set from readerWrtier", func() {
		BeforeEach(func() {
			managementPackageVersion = verStr
			outputFile = "test/output_vsphere_with_custom_repo_ca_rw.yaml"

		})

		JustBeforeEach(func() {
			// Configure tkgBoMConfig
			tkgBomConfig = &tkgconfigbom.BOMConfiguration{}
			err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
			Expect(err).NotTo(HaveOccurred())

			rw, err := tkgconfigreaderwriter.New("test/config.yaml")
			Expect(err).NotTo(HaveOccurred())

			rw.TKGConfigReaderWriter().Set("TKG_CUSTOM_IMAGE_REPOSITORY", "fake-repo-2")
			rw.TKGConfigReaderWriter().Set("TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE", "fake-ca-2")

			// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
			valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, rw.TKGConfigReaderWriter())
			Expect(err).NotTo(HaveOccurred())
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
})
