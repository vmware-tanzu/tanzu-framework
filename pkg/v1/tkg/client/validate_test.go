// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
)

var _ = Describe("Validate", func() {
	Context("vCenter IP and vSphere Control Plane Endpoint", func() {
		var (
			tkgClient       *client.TkgClient
			nodeSizeOptions client.NodeSizeOptions
			err             error
		)

		BeforeEach(func() {
			tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
			Expect(err).NotTo(HaveOccurred())

			nodeSizeOptions = client.NodeSizeOptions{
				Size:             "medium",
				ControlPlaneSize: "medium",
				WorkerSize:       "medium",
			}
		})

		Context("When vCenter IP and vSphere Control Plane Endpoint are different", func() {
			It("Should validate successfully", func() {
				vip := "10.10.10.11"
				err = tkgClient.ConfigureAndValidateVsphereConfig("", nodeSizeOptions, vip, true, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When vCenter IP and vSphere Control Plane Endpoint are the same", func() {
			It("Should throw a validation error", func() {
				vip := "10.10.10.10"
				err = tkgClient.ConfigureAndValidateVsphereConfig("", nodeSizeOptions, vip, true, nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("ConfigureAndValidateManagementClusterConfiguration", func() {
		var (
			initRegionOptions     *client.InitRegionOptions
			tkgClient             *client.TkgClient
			tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
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

			options := client.Options{
				ReaderWriterConfigClient: readerWriter,
				TKGConfigUpdater:         tkgConfigUpdater,
				TKGBomClient:             tkgBomClient,
				RegionManager:            new(fakes.RegionManager),
			}
			tkgClient, err = client.New(options)
			Expect(err).NotTo(HaveOccurred())

			initRegionOptions = &client.InitRegionOptions{
				Plan:                        "dev",
				InfrastructureProvider:      "vsphere",
				VsphereControlPlaneEndpoint: "foo.bar",
			}
			tkgConfigReaderWriter.Set(constants.ConfigVariableVsphereNetwork, "foo network")
		})

		Context("IPFamily configuration and validation", func() {
			It("should allow empty IPFamily fields", func() {
				validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
				Expect(validationError).NotTo(HaveOccurred())
			})

			Context("when IPFamily is empty", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "")
				})

				Context("when SERVICE_CIDR and CLUSTER_CIDR are ipv4", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "192.168.2.1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "192.168.2.1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})

				Context("when SERVICE_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"::1/8\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})

				Context("when CLUSTER_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/16")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"::1/8\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})

				Context("HTTP(S)_PROXY variables", func() {
					BeforeEach(func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "1.2.3.4/8")
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6 with ports", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]:3128")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://[::1]:3128\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
				})
			})

			Context("when IPFamily is ipv4", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4")
				})

				Context("when SERVICE_CIDR and CLUSTER_CIDR are ipv4", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "192.168.2.1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "192.168.2.1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when SERVICE_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"::1/8\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/16")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"::1/8\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when SERVICE_CIDR is not an actual CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"1.2.3.4\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is not an actual CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "1.2.3.4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"1.2.3.4\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when SERVICE_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableServiceCIDR)
						Expect(cidr).To(Equal("100.64.0.0/13"))
					})
				})
				Context("when CLUSTER_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableClusterCIDR)
						Expect(cidr).To(Equal("100.96.0.0/11"))
					})
				})
				Context("when SERVICE_CIDR is garbage", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "klsfda")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"klsfda\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is garbage", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "aoiwnf")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"aoiwnf\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("HTTP(S)_PROXY variables", func() {
					BeforeEach(func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "1.2.3.4/8")
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are unset", func() {
						It("should pass validation", func() {
							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv4", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://1.2.3.4")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6 with ports", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]:3128")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://[::1]:3128\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv4 with ports", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://1.2.3.4:3128")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are domain names", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY is ipv6", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://[::1]\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTPS_PROXY is ipv6", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "http://[::1]")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTPS_PROXY \"http://[::1]\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
				})
			})

			Context("when IPFamily is ipv6", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
				})
				Context("when SERVICE_CIDR and CLUSTER_CIDR are ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when SERVICE_CIDR is ipv4", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/16")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"1.2.3.4/16\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is ipv4", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "1.2.3.4/16")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"1.2.3.4/16\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when SERVICE_CIDR is not an actual CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"::1\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is not an actual CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"::1\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when SERVICE_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1/8")
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableServiceCIDR)
						Expect(cidr).To(Equal("fd00:100:64::/108"))
					})
				})
				Context("when CLUSTER_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableClusterCIDR)
						Expect(cidr).To(Equal("fd00:100:96::/48"))
					})
				})
				Context("when SERVICE_CIDR is garbage", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "klsfda")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"klsfda\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is garbage", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "aoiwnf")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"aoiwnf\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("HTTP(S)_PROXY variables", func() {
					BeforeEach(func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/8")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1/8")
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are unset", func() {
						It("should pass validation", func() {
							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6 with ports", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]:3128")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv4 with ports", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://1.2.3.4:3128")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://1.2.3.4:3128\", expected to be an address of type \"ipv6\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are domain names", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY is ipv4", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://1.2.3.4\", expected to be an address of type \"ipv6\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTPS_PROXY is ipv4", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "http://1.2.3.4")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTPS_PROXY \"http://1.2.3.4\", expected to be an address of type \"ipv6\" (TKG_IP_FAMILY)"))
						})
					})
				})
			})
		})
	})
})
