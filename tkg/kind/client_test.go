// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kind_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kindv1 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/kind"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

var (
	testingDir                            string
	defaultBoMFileForTesting              = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
	configPath                            = "../fakes/config/config6.yaml"
	configPathCustomRegistrySkipTLSVerify = "../fakes/config/config_custom_registry_skip_tls_verify.yaml"
	configPathCustomRegistryCaCert        = "../fakes/config/config_custom_registry_ca_cert.yaml"
	configPathIPv6                        = "../fakes/config/config_ipv6.yaml"
	configPathIPv4                        = "../fakes/config/config_ipv4.yaml"
	configPathIPv4IPv6                    = "../fakes/config/config_ipv4_ipv6.yaml"
	configPathIPv6IPv4                    = "../fakes/config/config_ipv6_ipv4.yaml"
	configPathIPv6IPv4WithCIDRS           = "../fakes/config/config_ipv6_ipv4_with_cidrs.yaml"
	configPathCIDR                        = "../fakes/config/config_cluster_service_cidr.yaml"
	registryHostname                      = "registry.mydomain.com"
)

func TestKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kind Suite")
}

var (
	kindClient   kind.Client
	kindProvider *fakes.KindProvider
	err          error
	clusterName  string
	kindConfig   *kindv1.Cluster
)

var _ = Describe("Kind Client", func() {
	BeforeSuite(func() {
		testingDir = fakehelper.CreateTempTestingDirectory()
	})

	AfterSuite(func() {
		fakehelper.DeleteTempTestingDirectory(testingDir)
	})

	Context("When TKG_CUSTOM_IMAGE_REPOSITORY is not set", func() {
		BeforeEach(func() {
			setupTestingFiles(configPath, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
		})

		Describe("Create bootstrap kind cluster", func() {
			JustBeforeEach(func() {
				clusterName, err = kindClient.CreateKindCluster()
			})

			Context("When kind provider fails to create kind cluster", func() {
				BeforeEach(func() {
					kindProvider.CreateReturns(errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("failed to create kind cluster"))
				})
			})

			Context("When kind provider create kind cluster but unable to retrieve kubeconfig", func() {
				BeforeEach(func() {
					kindProvider.CreateReturns(nil)
					kindProvider.KubeConfigReturns("", errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("unable to retrieve kubeconfig for created kind cluster"))
				})
			})

			Context("When kind provider create kind cluster and able to retrieve kubeconfig successfully", func() {
				BeforeEach(func() {
					kindProvider.CreateReturns(nil)
					kindProvider.KubeConfigReturns("fake-kube-config", nil)
				})
				It("does not return error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(clusterName).To(Equal("clusterName"))
				})
			})
		})

		Describe("Delete bootstrap kind cluster", func() {
			JustBeforeEach(func() {
				err = kindClient.DeleteKindCluster()
			})

			Context("When kind provider fails to delete kind cluster", func() {
				BeforeEach(func() {
					kindProvider.DeleteReturns(errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("failed to delete kind cluster"))
				})
			})

			Context("When kind provider deletes kind cluster successfully", func() {
				BeforeEach(func() {
					kindProvider.DeleteReturns(nil)
				})
				It("returns an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Context("When TKG_CUSTOM_IMAGE_REPOSITORY is set", func() {
		Context("When TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY is set to true", func() {
			BeforeEach(func() {
				setupTestingFiles(configPathCustomRegistrySkipTLSVerify, testingDir, defaultBoMFileForTesting)
				kindClient = buildKindClient()
				_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("Generate kind cluster config", func() {
				It("generates 'insecure_skip_verify = true' in containerdConfigPatches", func() {
					Expect(kindConfig.ContainerdConfigPatches[0]).Should(ContainSubstring(fmt.Sprintf("plugins.'io.containerd.grpc.v1.cri'.registry.configs.'%s'.tls", registryHostname)))
					Expect(kindConfig.ContainerdConfigPatches[0]).Should(ContainSubstring("insecure_skip_verify = true"))
					Expect(kindConfig.ContainerdConfigPatches[0]).ShouldNot(ContainSubstring("ca_file = '/etc/containerd/tkg-registry-ca.crt'"))
					Expect(len(kindConfig.Nodes[0].ExtraMounts)).To(Equal(1))
				})
			})
		})

		Context("When TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE is set to non-empty string", func() {
			BeforeEach(func() {
				setupTestingFiles(configPathCustomRegistryCaCert, testingDir, defaultBoMFileForTesting)
				kindClient = buildKindClient()
				_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("Generate kind cluster config", func() {
				It("generates ca_file config in containerdConfigPatches", func() {
					Expect(kindConfig.ContainerdConfigPatches[0]).Should(ContainSubstring(fmt.Sprintf("plugins.'io.containerd.grpc.v1.cri'.registry.configs.'%s'.tls", registryHostname)))
					Expect(kindConfig.ContainerdConfigPatches[0]).Should(ContainSubstring("insecure_skip_verify = false"))
					Expect(kindConfig.ContainerdConfigPatches[0]).Should(ContainSubstring("ca_file = '/etc/containerd/tkg-registry-ca.crt'"))
					Expect(kindConfig.Nodes[0].ExtraMounts[1].ContainerPath).Should(ContainSubstring("/etc/containerd/tkg-registry-ca.crt"))
				})
			})
		})
	})

	Context("When TKG_IP_FAMILY is unset", func() {
		BeforeEach(func() {
			setupTestingFiles(configPath, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with ipfamily omitted", func() {
			Expect(string(kindConfig.Networking.IPFamily)).To(Equal(""))
		})

		Context("When CLUSTER_CIDR and SERVICE_CIDR are not set", func() {
			It("generates a config with default pod and service subnet", func() {
				Expect(kindConfig.Networking.PodSubnet).To(Equal("100.96.0.0/11"))
				Expect(kindConfig.Networking.ServiceSubnet).To(Equal("100.64.0.0/13"))
			})
		})
	})

	Context("When TKG_IP_FAMILY is ipv4", func() {
		BeforeEach(func() {
			setupTestingFiles(configPathIPv4, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with ipfamily set to ipv4", func() {
			Expect(kindConfig.Networking.IPFamily).To(Equal(kindv1.IPv4Family))
		})
	})

	Context("When TKG_IP_FAMILY is ipv6", func() {
		BeforeEach(func() {
			setupTestingFiles(configPathIPv6, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with ipfamily set to ipv6", func() {
			Expect(kindConfig.Networking.IPFamily).To(Equal(kindv1.IPv6Family))
		})

		Context("When CLUSTER_CIDR and SERVICE_CIDR are not set", func() {
			It("generates a config with default pod and service subnet", func() {
				Expect(kindConfig.Networking.PodSubnet).To(Equal("fd00:100:96::/48"))
				Expect(kindConfig.Networking.ServiceSubnet).To(Equal("fd00:100:64::/108"))
			})
		})
	})

	Context("When TKG_IP_FAMILY is ipv4,ipv6", func() {
		BeforeEach(func() {
			setupTestingFiles(configPathIPv4IPv6, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with ipfamily set to dual", func() {
			Expect(kindConfig.Networking.IPFamily).To(Equal(kindv1.DualStackFamily))
		})

		Context("When CLUSTER_CIDR and SERVICE_CIDR are not set", func() {
			It("generates a config with default pod and service subnet", func() {
				Expect(kindConfig.Networking.PodSubnet).To(Equal("100.96.0.0/11,fd00:100:96::/48"))
				Expect(kindConfig.Networking.ServiceSubnet).To(Equal("100.64.0.0/13,fd00:100:64::/108"))
			})
		})
	})

	Context("When TKG_IP_FAMILY is ipv6,ipv4 and config vars are unset", func() {
		// This context should never happen, the initRegion code will always
		// ensure there's a default value set if the user has not provided one.
		BeforeEach(func() {
			setupTestingFiles(configPathIPv6IPv4, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with ipfamily set to dual", func() {
			Expect(kindConfig.Networking.IPFamily).To(Equal(kindv1.DualStackFamily))
		})

		It("generates a config with default pod and service subnet", func() {
			Expect(kindConfig.Networking.PodSubnet).To(Equal("100.96.0.0/11,fd00:100:96::/48"))
			Expect(kindConfig.Networking.ServiceSubnet).To(Equal("100.64.0.0/13,fd00:100:64::/108"))
		})
	})

	Context("When TKG_IP_FAMILY is ipv6,ipv4 and SERVICE_CIDR and CLUSTER_CIDR are set", func() {
		BeforeEach(func() {
			setupTestingFiles(configPathIPv6IPv4WithCIDRS, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with ipfamily set to dual", func() {
			Expect(kindConfig.Networking.IPFamily).To(Equal(kindv1.DualStackFamily))
		})

		It("generates a config with default pod and service subnet in ipv4,ipv6 order because Kind's dualstack family is ipv4,ipv6", func() {
			Expect(kindConfig.Networking.PodSubnet).To(Equal("100.96.0.0/11,fd00:100:96::/48"))
			Expect(kindConfig.Networking.ServiceSubnet).To(Equal("1.2.3.4/16,fd00::/48"))
		})
	})

	Context("When CLUSTER_CIDR and SERVICE_CIDR are explicitly set", func() {
		BeforeEach(func() {
			setupTestingFiles(configPathCIDR, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
			_, kindConfig, err = kindClient.GetKindNodeImageAndConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates a config with specified pod and service subnet", func() {
			Expect(kindConfig.Networking.PodSubnet).To(Equal("200.200.200.0/24"))
			Expect(kindConfig.Networking.ServiceSubnet).To(Equal("250.250.250.0/24"))
		})
	})
})

func buildKindClient() kind.Client {
	tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
	Expect(err).NotTo(HaveOccurred())

	kindProvider = &fakes.KindProvider{}
	options := kind.KindClusterOptions{
		Provider:       kindProvider,
		ClusterName:    "clusterName",
		NodeImage:      "nodeImage",
		KubeConfigPath: "kubeConfigPath",
		TKGConfigDir:   testingDir,
		Readerwriter:   tkgConfigReaderWriter,
	}
	return kind.New(&options)
}

var testTKGCompatibilityFileFmt = `
version: v1
managementClusterPluginVersions:
- version: %s
  supportedTKGBomVersions:
  - imagePath: tkg-bom
    tag: %s
`

func setupTestingFiles(clusterConfigFile string, configDir string, defaultBomFile string) {
	testClusterConfigFile := filepath.Join(configDir, "config.yaml")
	err := utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())

	compatibilityDir, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityConfigPath()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	Expect(err).ToNot(HaveOccurred())
}
