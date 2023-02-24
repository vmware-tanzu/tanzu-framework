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
		valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil, true)
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
		valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil, true)
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
		})

		It("should not return error", func() {
			// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
			valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil, true)

			Expect(err).NotTo(HaveOccurred())
			f1, err := os.ReadFile(valuesFile)
			Expect(err).NotTo(HaveOccurred())
			f2, err := os.ReadFile(outputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(f1)).To(Equal(string(f2)))
		})

		When("skipVerifyCert is set in user config", func() {

			It("skipVerify should be correctly parsed", func() {
				userProviderConfigValues["TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"] = 1
			})
			It("skipVerify should be correctly parsed", func() {
				userProviderConfigValues["TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"] = "1"
			})
			It("skipVerify should be correctly parsed", func() {
				userProviderConfigValues["TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"] = true
			})
			It("skipVerify should be correctly parsed", func() {
				userProviderConfigValues["TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"] = "true"
			})

			AfterEach(func() {
				tkgPackageConfig, err := GetTKGPackageConfigFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(tkgPackageConfig.TKRSourceControllerPackage.TKRSourceControllerPackageValues.SkipVerifyCert).To(BeTrue())
			})
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
			valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, rw.TKGConfigReaderWriter(), true)
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

var _ = Describe("Test AVI related settings", func() {
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

	Context("On bootstrap cluster", func() {
		JustBeforeEach(func() {
			// Configure tkgBoMConfig
			tkgBomConfig = &tkgconfigbom.BOMConfiguration{}
			err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
			Expect(err).NotTo(HaveOccurred())

			// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
			valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil, true)
		})

		Context("when AVI_ENABLE is set to true", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_bootstrap_cluster.yaml"
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_INGRESS_NODE_NETWORK_LIST": `- networkName: node-network-name
  cidrs:
    - 10.191.176.0/20
`,
					"AVI_PASSWORD":             "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP": "Default-Group",
					"AVI_USERNAME":             "admin",
					"PROVIDER_TYPE":            "vsphere",
				}
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

		Context("when AVI_ENABLE is set to true, AVI_INGRESS_NODE_NETWORK_LIST is not set", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_no_node_network_list.yaml"
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_PASSWORD":                  "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":      "Default-Group",
					"AVI_USERNAME":                  "admin",
					"PROVIDER_TYPE":                 "vsphere",
					"VSPHERE_NETWORK":               "VM Network",
				}
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

		Context("when AVI_ENABLE is set to true, AVI_INGRESS_NODE_NETWORK_LIST is empty", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_empty_node_network_list.yaml"
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_PASSWORD":                  "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":      "Default-Group",
					"AVI_USERNAME":                  "admin",
					"PROVIDER_TYPE":                 "vsphere",
					"VSPHERE_NETWORK":               "VM Network",
					"AVI_INGRESS_NODE_NETWORK_LIST": `""`,
				}
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

		Context("when AVI_ENABLE is set to true, AVI_INGRESS_NODE_NETWORK_LIST is invalid", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_INGRESS_NODE_NETWORK_LIST": "VM Network",
					"AVI_PASSWORD":                  "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":      "Default-Group",
					"AVI_USERNAME":                  "admin",
					"PROVIDER_TYPE":                 "vsphere",
				}
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error convert node network list"))
			})
		})

		Context("when AVI_ENABLE is set to true, VIP network is fully customized", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_custom_vip_network.yaml"
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_INGRESS_NODE_NETWORK_LIST": `- networkName: node-network-name
  cidrs:
    - 10.191.176.0/20
`,
					"AVI_PASSWORD":                                          "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":                              "Default-Group",
					"AVI_USERNAME":                                          "admin",
					"PROVIDER_TYPE":                                         "vsphere",
					"AVI_CONTROL_PLANE_NETWORK":                             "avi-control-plane-network",
					"AVI_CONTROL_PLANE_NETWORK_CIDR":                        "10.10.93.25/20",
					"AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_NAME":               "avi-management-cluster-vip-network",
					"AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_CIDR":               "10.94.13.45/20",
					"AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_NAME": "avi-management-cluster-control-plane-vip-network",
					"AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_CIDR": "10.48.99.33/20",
				}
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

		Context("when AVI_ENABLE is set to false", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_disabled.yaml"
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":    false,
					"PROVIDER_TYPE": "vsphere",
				}
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

	Context("On management cluster", func() {
		JustBeforeEach(func() {
			// Configure tkgBoMConfig
			tkgBomConfig = &tkgconfigbom.BOMConfiguration{}
			err = yaml.Unmarshal([]byte(tkgBomConfigData), tkgBomConfig)
			Expect(err).NotTo(HaveOccurred())

			// invoke GetTKGPackageConfigValuesFileFromUserConfig for testing using addonsManagerPackageVersion = managementPackageVersion
			valuesFile, err = GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, managementPackageVersion, userProviderConfigValues, tkgBomConfig, nil, false)
		})

		Context("when AVI_ENABLE is set to true", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_management_cluster.yaml"
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_INGRESS_NODE_NETWORK_LIST": `- networkName: node-network-name
  cidrs:
    - 10.191.176.0/20
`,
					"AVI_PASSWORD":             "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP": "Default-Group",
					"AVI_USERNAME":             "admin",
					"PROVIDER_TYPE":            "vsphere",
				}
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

		Context("when AVI_ENABLE is set to true", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_with_nsxt_cloud_management_cluster.yaml"
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_PASSWORD":                  "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":      "Default-Group",
					"AVI_USERNAME":                  "admin",
					"AVI_NSXT_T1LR":                 "/infra/test_t1",
					"PROVIDER_TYPE":                 "vsphere",
					"VSPHERE_NETWORK":               "VM Network",
				}
			})
			It("when set NSX-T T1 router, it should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
				f1, err := os.ReadFile(valuesFile)
				Expect(err).NotTo(HaveOccurred())
				f2, err := os.ReadFile(outputFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(f1)).To(Equal(string(f2)))
			})
		})

		Context("when AVI_ENABLE is set to true", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_with_avi_labels_0_management_cluster.yaml"
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_PASSWORD":                  "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":      "Default-Group",
					"AVI_USERNAME":                  "admin",
					"AVI_LABELS":                    `{"foo":"bar"}`,
					"PROVIDER_TYPE":                 "vsphere",
					"VSPHERE_NETWORK":               "VM Network",
				}
			})
			It("set AVI_LABELS, it should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
				f1, err := os.ReadFile(valuesFile)
				Expect(err).NotTo(HaveOccurred())
				f2, err := os.ReadFile(outputFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(f1)).To(Equal(string(f2)))
			})
		})

		Context("when AVI_ENABLE is set to true", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_enabled_with_avi_labels_1_management_cluster.yaml"
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":                    true,
					"AVI_CLOUD_NAME":                "Default-Cloud",
					"AVI_CONTROL_PLANE_HA_PROVIDER": true,
					"AVI_CONTROLLER":                "10.191.186.55",
					"AVI_DATA_NETWORK":              "VM Network",
					"AVI_DATA_NETWORK_CIDR":         "10.191.176.0/20",
					"AVI_PASSWORD":                  "Admin!23",
					"AVI_SERVICE_ENGINE_GROUP":      "Default-Group",
					"AVI_USERNAME":                  "admin",
					"AVI_LABELS":                    map[string]string{"foo": "bar"},
					"PROVIDER_TYPE":                 "vsphere",
					"VSPHERE_NETWORK":               "VM Network",
				}
			})
			It("set AVI_LABELS, it should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
				f1, err := os.ReadFile(valuesFile)
				Expect(err).NotTo(HaveOccurred())
				f2, err := os.ReadFile(outputFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(f1)).To(Equal(string(f2)))
			})
		})

		Context("when AVI_ENABLE is set to false", func() {
			BeforeEach(func() {
				managementPackageVersion = verStr
				outputFile = "test/output_vsphere_with_avi_disabled.yaml"
				// Configure user provider configuration
				userProviderConfigValues = map[string]interface{}{
					"AVI_ENABLE":    false,
					"PROVIDER_TYPE": "vsphere",
				}
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
})
