// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	aviMock "github.com/vmware-tanzu/tanzu-framework/tkg/avi/mocks"
	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"

	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
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
	Context("vCenter IP and vSphere Control Plane Endpoint", func() {
		var (
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
			initRegionOptions *client.InitRegionOptions
		)

		BeforeEach(func() {
			initRegionOptions = &client.InitRegionOptions{
				Plan:                        "dev",
				InfrastructureProvider:      "vsphere",
				VsphereControlPlaneEndpoint: "foo.bar",
				Edition:                     "tkg",
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
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "192.168.2.1/12")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "192.168.2.1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})

				Context("when SERVICE_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"::1/108\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})

				Context("when CLUSTER_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"::1/8\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})

				Context("HTTP(S)_PROXY variables", func() {
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6 with ports", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://[::1]:3128\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are present without TKG_HTTP_PROXY_ENABLED set to true", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://1.2.3.4:3128")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("cannot get TKG_HTTP_PROXY_ENABLED: Failed to get value for variable \"TKG_HTTP_PROXY_ENABLED\". Please set the variable value using os env variables or using the config file"))
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
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "192.168.2.1/12")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "192.168.2.1/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when SERVICE_CIDR is ipv6", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"::1/108\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when CLUSTER_CIDR is ipv6", func() {
					It("should fail validation", func() {
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
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "aoiwnf")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"aoiwnf\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when multiple CIDRs are provided to SERVICE_CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/12,1.2.3.5/12")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"1.2.3.4/12,1.2.3.5/12\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when multiple CIDRs are provided to CLUSTER_CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "1.2.3.5/8,1.2.3.6/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"1.2.3.5/8,1.2.3.6/8\", expected to be a CIDR of type \"ipv4\" (TKG_IP_FAMILY)"))
					})
				})
				Context("HTTP(S)_PROXY variables", func() {
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
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY set but not TKG_HTTP_PROXY_ENABLE is unset", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "http://1.2.3.4")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("cannot get TKG_HTTP_PROXY_ENABLED: Failed to get value for variable \"TKG_HTTP_PROXY_ENABLED\". Please set the variable value using os env variables or using the config file"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6 with ports", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://[::1]:3128\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv4 with ports", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are domain names", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY is ipv6", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://[::1]\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTPS_PROXY is ipv6", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "http://[::1]")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTPS_PROXY \"http://[::1]\", expected to be an address of type \"ipv4\" (TKG_IP_FAMILY)"))
						})
					})
					DescribeTable("NO_PROXY validate", func(httpProxy, httpsProxy, noProxy string, hasError bool) {
						tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, httpProxy)
						tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, httpsProxy)
						tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")
						tkgConfigReaderWriter.Set(constants.TKGNoProxy, noProxy)

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						if hasError {
							Expect(validationError).To(HaveOccurred())
							return
						}

						Expect(validationError).NotTo(HaveOccurred())
						v, err := tkgConfigReaderWriter.Get(constants.TKGNoProxy)
						Expect(err).NotTo(HaveOccurred())
						Expect(v).NotTo(ContainSubstring(" "))
						Expect(v).NotTo(ContainSubstring("	"))
						Expect(v).NotTo(ContainSubstring(`
						`))
					},
						Entry("No proxy has new line, trim new line", "http://1.2.3.4", "http://1.2.3.4", `10.2.1.3/23,
                			10.1.3.3`, false),
						Entry("No Proxy has space, trim space", "http://1.2.3.4", "http://1.2.3.4", "example.com, svc.c", false),
						Entry("No Proxy has *", "http://1.2.3.4", "http://1.2.3.4", "example.com, svc.c,*.vmware.com", true),
						Entry("No Proxy", "http://1.2.3.4", "http://1.2.3.4", "10.0.0.0/24", false),
					)
				})
			})

			Context("when IPFamily is ipv6", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
				})
				Context("when SERVICE_CIDR and CLUSTER_CIDR are ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108")
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
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::1")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"::1\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when SERVICE_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
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
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "aoiwnf")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"aoiwnf\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when multiple CIDRs are provided to SERVICE_CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108,::2/108")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid SERVICE_CIDR \"::1/108,::2/108\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("when multiple CIDRs are provided to CLUSTER_CIDR", func() {
					It("should fail validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::3/8,::4/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CLUSTER_CIDR \"::3/8,::4/8\", expected to be a CIDR of type \"ipv6\" (TKG_IP_FAMILY)"))
					})
				})
				Context("HTTP(S)_PROXY variables", func() {
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
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY set but not TKG_HTTP_PROXY_ENABLE is unset", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "http://[::1]")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("cannot get TKG_HTTP_PROXY_ENABLED: Failed to get value for variable \"TKG_HTTP_PROXY_ENABLED\". Please set the variable value using os env variables or using the config file"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv6 with ports", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://[::1]:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are ipv4 with ports", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://1.2.3.4:3128")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://1.2.3.4:3128\", expected to be an address of type \"ipv6\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTP_PROXY and HTTPS_PROXY are domain names", func() {
						It("should pass validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).NotTo(HaveOccurred())
						})
					})
					Context("when HTTP_PROXY is ipv4", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "http://1.2.3.4")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTP_PROXY \"http://1.2.3.4\", expected to be an address of type \"ipv6\" (TKG_IP_FAMILY)"))
						})
					})
					Context("when HTTPS_PROXY is ipv4", func() {
						It("should fail validation", func() {
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, "https://foo.bar.com")
							tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, "http://1.2.3.4")
							tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

							validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
							Expect(validationError).To(HaveOccurred())
							Expect(validationError.Error()).To(ContainSubstring("invalid TKG_HTTPS_PROXY \"http://1.2.3.4\", expected to be an address of type \"ipv6\" (TKG_IP_FAMILY)"))
						})
					})
				})
			})

			Context("when IPFamily is ipv4,ipv6 i.e Dual-stack Primary IPv4", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
				})

				Context("when dual-stack-ipv4-primary feature gate is false", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(false, nil)
					})
					It("returns an error", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("option TKG_IP_FAMILY is set to \"ipv4,ipv6\", but dualstack support is not enabled (because it is under development). To enable dualstack, set features.management-cluster.dual-stack-ipv4-primary to \"true\""))
					})
				})
				Context("when SERVICE_CIDR and CLUSTER_CIDR are ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/16,::1/108")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "1.2.3.5/16,::3/8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})

				Context("when SERVICE_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableServiceCIDR)
						Expect(cidr).To(Equal("100.64.0.0/13,fd00:100:64::/108"))
					})
				})

				Context("when CLUSTER_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableClusterCIDR)
						Expect(cidr).To(Equal("100.96.0.0/11,fd00:100:96::/48"))
					})
				})

				DescribeTable("HTTP(S)_PROXY variables", func(httpProxy, httpsProxy string) {
					tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, httpProxy)
					tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, httpsProxy)
					tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())
				},
					Entry("IPv6 Address", "http://[::1]", "https://[::1]"),
					Entry("IPv6 Address with Ports", "http://[::1]:3128", "https://[::1]:3128"),
					Entry("IPv4 Address", "http://1.2.3.4", "https://1.2.3.4"),
					Entry("IPv4 Address with Ports", "http://1.2.3.4:3128", "https://1.2.3.4:3128"),
					Entry("Domain Name", "http://foo.bar.com", "https://foo.bar.com"),
				)

				DescribeTable("HTTP(S)_PROXY variables without TKGHTTPProxyEnabled set to true", func(httpProxy, httpsProxy string) {
					tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, httpProxy)
					tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, httpsProxy)

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).To(HaveOccurred())
					Expect(validationError.Error()).To(ContainSubstring("cannot get TKG_HTTP_PROXY_ENABLED: Failed to get value for variable \"TKG_HTTP_PROXY_ENABLED\". Please set the variable value using os env variables or using the config file"))
				},
					Entry("IPv6 Address", "http://[::1]", "https://[::1]"),
					Entry("IPv6 Address with Ports", "http://[::1]:3128", "https://[::1]:3128"),
					Entry("IPv4 Address", "http://1.2.3.4", "https://1.2.3.4"),
					Entry("IPv4 Address with Ports", "http://1.2.3.4:3128", "https://1.2.3.4:3128"),
					Entry("Domain Name", "http://foo.bar.com", "https://foo.bar.com"),
				)
			})

			Context("when IPFamily is ipv6,ipv4 i.e Dual-stack Primary IPv6", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6,ipv4")
				})

				Context("when dual-stack-ipv6-primary feature gate is false", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(false, nil)
					})
					It("returns an error", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("option TKG_IP_FAMILY is set to \"ipv6,ipv4\", but dualstack support is not enabled (because it is under development). To enable dualstack, set features.management-cluster.dual-stack-ipv6-primary to \"true\""))
					})
				})

				Context("when SERVICE_CIDR and CLUSTER_CIDR are ipv6,ipv4", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108,1.2.3.4/16")
						tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, "::3/8,1.2.3.5/16")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})

				Context("when SERVICE_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableServiceCIDR)
						Expect(cidr).To(Equal("fd00:100:64::/108,100.64.0.0/13"))
					})
				})

				Context("when CLUSTER_CIDR is undefined", func() {
					It("should set the default CIDR", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
						cidr, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableClusterCIDR)
						Expect(cidr).To(Equal("fd00:100:96::/48,100.96.0.0/11"))
					})
				})

				DescribeTable("HTTP(S)_PROXY variables", func(httpProxy, httpsProxy string) {
					tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, httpProxy)
					tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, httpsProxy)
					tkgConfigReaderWriter.Set(constants.TKGHTTPProxyEnabled, "true")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())
				},
					Entry("IPv6 Address", "http://[::1]", "https://[::1]"),
					Entry("IPv6 Address with Ports", "http://[::1]:3128", "https://[::1]:3128"),
					Entry("IPv4 Address", "http://1.2.3.4", "https://1.2.3.4"),
					Entry("IPv4 Address with Ports", "http://1.2.3.4:3128", "https://1.2.3.4:3128"),
					Entry("Domain Name", "http://foo.bar.com", "https://foo.bar.com"),
				)

				DescribeTable("HTTP(S)_PROXY variables without TKGHTTPProxyEnabled set to true", func(httpProxy, httpsProxy string) {
					tkgConfigReaderWriter.Set(constants.TKGHTTPProxy, httpProxy)
					tkgConfigReaderWriter.Set(constants.TKGHTTPSProxy, httpsProxy)

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).To(HaveOccurred())
					Expect(validationError.Error()).To(ContainSubstring("cannot get TKG_HTTP_PROXY_ENABLED: Failed to get value for variable \"TKG_HTTP_PROXY_ENABLED\". Please set the variable value using os env variables or using the config file"))
				},
					Entry("IPv6 Address", "http://[::1]", "https://[::1]"),
					Entry("IPv6 Address with Ports", "http://[::1]:3128", "https://[::1]:3128"),
					Entry("IPv4 Address", "http://1.2.3.4", "https://1.2.3.4"),
					Entry("IPv4 Address with Ports", "http://1.2.3.4:3128", "https://1.2.3.4:3128"),
					Entry("Domain Name", "http://foo.bar.com", "https://foo.bar.com"),
				)
			})

			const dualStackIPv4Primary = "ipv4,ipv6"
			const dualStackIPv6Primary = "ipv6,ipv4"
			const dualStackIPv4PrimaryFormatted = "<IPv4 CIDR>,<IPv6 CIDR>"
			const dualStackIPv6PrimaryFormatted = "<IPv6 CIDR>,<IPv4 CIDR>"

			DescribeTable("Dual-stack ServiceCIDR failure cases", func(ipFamily, serviceCIDRs string) {
				tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, ipFamily)
				tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, serviceCIDRs)
				var expectedFormat string
				switch ipFamily {
				case dualStackIPv4Primary:
					expectedFormat = dualStackIPv4PrimaryFormatted
				case dualStackIPv6Primary:
					expectedFormat = dualStackIPv6PrimaryFormatted
				}
				validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
				Expect(validationError).To(HaveOccurred())
				Expect(validationError).To(MatchError(
					fmt.Sprintf(`invalid SERVICE_CIDR %q, expected to have %q for TKG_IP_FAMILY %q`,
						serviceCIDRs,
						expectedFormat,
						ipFamily,
					),
				))
			},
				// Primary IPv4:
				Entry("Primary IPv4: IPv4 CIDR only", dualStackIPv4Primary, "1.2.3.4/16"),
				Entry("Primary IPv4: IPv6 CIDR only", dualStackIPv4Primary, "::1/108"),
				Entry("Primary IPv4: IPv4 Address", dualStackIPv4Primary, "1.2.3.4,::1/108"),
				Entry("Primary IPv4: IPv6 Address", dualStackIPv4Primary, "1.2.3.4/16,::1"),
				Entry("Primary IPv4: Too many CIDRs", dualStackIPv4Primary, "1.2.3.4/16,::1/108,::2/108"),
				Entry("Primary IPv4: Two IPv4 CIDRs", dualStackIPv4Primary, "1.2.3.4/16,2.3.4.5/16"),
				Entry("Primary IPv4: Two IPv6 CIDRs", dualStackIPv4Primary, "::1/108,::2/108"),
				Entry("Primary IPv4: Out of order", dualStackIPv4Primary, "::1/108,1.2.3.4/16"),
				Entry("Primary IPv4: Garbage", dualStackIPv4Primary, "asdf,fasd"),
				// Primary Ipv6:
				Entry("Primary IPv6: IPv4 CIDR only", dualStackIPv6Primary, "1.2.3.4/16"),
				Entry("Primary IPv6: IPv6 CIDR only", dualStackIPv6Primary, "::1/108"),
				Entry("Primary IPv6: IPv4 Address", dualStackIPv6Primary, "::1/108,1.2.3.4"),
				Entry("Primary IPv6: IPv6 Address", dualStackIPv6Primary, "::1,1.2.3.4/16"),
				Entry("Primary IPv6: Too many CIDRs", dualStackIPv6Primary, "::1/108,::2/108,1.2.3.4/16"),
				Entry("Primary IPv6: Two IPv4 CIDRs", dualStackIPv6Primary, "1.2.3.4/16,2.3.4.5/16"),
				Entry("Primary IPv6: Two IPv6 CIDRs", dualStackIPv6Primary, "::1/108,::2/108"),
				Entry("Primary IPv6: Out of order", dualStackIPv6Primary, "1.2.3.4/16,::1/108"),
				Entry("Primary IPv6: Garbage", dualStackIPv6Primary, "asdf,fasd"),
			)

			DescribeTable("Dual Stack ClusterCIDR failure cases", func(ipFamily, clusterCIDRs string) {
				tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, ipFamily)
				tkgConfigReaderWriter.Set(constants.ConfigVariableClusterCIDR, clusterCIDRs)
				var expectedFormat string
				switch ipFamily {
				case dualStackIPv4Primary:
					expectedFormat = dualStackIPv4PrimaryFormatted
				case dualStackIPv6Primary:
					expectedFormat = dualStackIPv6PrimaryFormatted
				}
				validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
				Expect(validationError).To(HaveOccurred())
				Expect(validationError).To(MatchError(
					fmt.Sprintf(`invalid CLUSTER_CIDR %q, expected to have %q for TKG_IP_FAMILY %q`,
						clusterCIDRs,
						expectedFormat,
						ipFamily,
					),
				))
			},
				// Primary IPv4:
				Entry("Primary IPv4: IPv4 CIDR only", dualStackIPv4Primary, "1.2.3.4/16"),
				Entry("Primary IPv4: IPv6 CIDR only", dualStackIPv4Primary, "::1/8"),
				Entry("Primary IPv4: IPv4 Address", dualStackIPv4Primary, "1.2.3.4,::1/8"),
				Entry("Primary IPv4: IPv6 Address", dualStackIPv4Primary, "1.2.3.4/16,::1"),
				Entry("Primary IPv4: Too many CIDRs", dualStackIPv4Primary, "1.2.3.4/16,::1/8,::2/8"),
				Entry("Primary IPv4: Two IPv4 CIDRs", dualStackIPv4Primary, "1.2.3.4/16,2.3.4.5/16"),
				Entry("Primary IPv4: Two IPv6 CIDRs", dualStackIPv4Primary, "::1/8,::2/8"),
				Entry("Primary IPv4: Out of order", dualStackIPv4Primary, "::1/8,1.2.3.4/16"),
				Entry("Primary IPv4: Garbage", dualStackIPv4Primary, "asdf,fasd"),
				// Primary Ipv6:
				Entry("Primary IPv6: IPv4 CIDR only", dualStackIPv6Primary, "1.2.3.4/16"),
				Entry("Primary IPv6: IPv6 CIDR only", dualStackIPv6Primary, "::1/8"),
				Entry("Primary IPv6: IPv4 Address", dualStackIPv6Primary, "::1/8,1.2.3.4"),
				Entry("Primary IPv6: IPv6 Address", dualStackIPv6Primary, "::1,1.2.3.4/16"),
				Entry("Primary IPv6: Too many CIDRs", dualStackIPv6Primary, "::1/8,::2/8,1.2.3.4/16"),
				Entry("Primary IPv6: Two IPv4 CIDRs", dualStackIPv6Primary, "1.2.3.4/16,2.3.4.5/16"),
				Entry("Primary IPv6: Two IPv6 CIDRs", dualStackIPv6Primary, "::1/8,::2/8"),
				Entry("Primary IPv6: Out of order", dualStackIPv6Primary, "1.2.3.4/16,::1/8"),
				Entry("Primary IPv6: Garbage", dualStackIPv6Primary, "asdf,fasd"),
			)
		})

		DescribeTable("SERVICE_CIDR size validation - invalid cases", func(ipFamily, serviceCIDR, problematicCIDR, netmaskSizeConstraint string) {
			tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, ipFamily)
			tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, serviceCIDR)

			validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
			Expect(validationError).To(HaveOccurred())
			expectedErrorMsg := fmt.Sprintf(`invalid SERVICE_CIDR "%s", expected netmask to be "%s" or greater`, problematicCIDR, netmaskSizeConstraint)
			Expect(validationError.Error()).To(ContainSubstring(expectedErrorMsg))
		},
			Entry("ipv4 cidr too large", "ipv4", "192.168.2.1/11", "192.168.2.1/11", "/12"),
			Entry("ipv6 cidr too large", "ipv6", "::1/107", "::1/107", "/108"),
			Entry("ipv4-primary dualstack: ipv4 cidr too large", "ipv4,ipv6", "1.2.3.4/11,::1/108", "1.2.3.4/11", "/12"),
			Entry("ipv4-primary dualstack: ipv6 cidr too large", "ipv4,ipv6", "1.2.3.4/12,::1/107", "::1/107", "/108"),
			Entry("ipv6-primary dualstack: ipv6 cidr too large", "ipv6,ipv4", "::1/107,1.2.3.4/12", "::1/107", "/108"),
			Entry("ipv6-primary dualstack: ipv4 cidr too large", "ipv6,ipv4", "::1/108,1.2.3.4/11", "1.2.3.4/11", "/12"),
		)

		DescribeTable("SERVICE_CIDR size validation - valid cases", func(ipFamily, serviceCIDR string) {
			tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, ipFamily)
			tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, serviceCIDR)

			validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
			Expect(validationError).NotTo(HaveOccurred())
		},
			Entry("ipv4 cidr at max size", "ipv4", "192.168.2.1/12"),
			Entry("ipv4 cidr at slightly smaller size", "ipv4", "192.168.2.1/13"),
			Entry("ipv6 cidr at max size", "ipv6", "::1/108"),
			Entry("ipv6 cidr at slightly smaller size", "ipv6", "::1/108"),
			Entry("ipv4-primary dualstack: cidrs at max size", "ipv4,ipv6", "1.2.3.4/12,::1/108"),
			Entry("ipv6-primary dualstack: cidrs at max size", "ipv6,ipv4", "::1/108,1.2.3.4/12"),
		)

		Context("Nameserver configuration and validation", func() {
			It("should allow empty nameserver configurations", func() {
				validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
				Expect(validationError).NotTo(HaveOccurred())
			})

			Context("Control Plane Node Nameservers", func() {
				Context("Custom Nameserver feature gate is false", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(false, nil)
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "8.8.8.8")
					})

					It("should return an error", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("option CONTROL_PLANE_NODE_NAMESERVERS is set to \"8.8.8.8\", but custom nameserver support is not enabled (because it is not fully functional). To enable custom nameservers, run the command: tanzu config set features.management-cluster.custom-nameservers true"))
					})
				})

				Context("when CONTROL_PLANE_NODE_NAMESERVERS is a valid set of IPv4 address", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "8.8.8.8,8.8.4.4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS is a valid set of IPv6 address and TKG_IP_FAMILY is ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "2001:DB8::1, 2001:DB8::2")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS only contains IPv6 addresses and TKG_IP_FAMILY is ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "2001:DB8::1,2001:DB8::2")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS only contains IPv4 addresses and TKG_IP_FAMILY is ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "8.8.8.8,8.8.4.4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS is a valid set of IPv4,IPv6 address and TKG_IP_FAMILY is ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "8.8.8.8,2001:DB8::2,8.8.4.4,2001:DB8::4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS contains multiple invalid entries", func() {
					It("should return an error", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "google.dns,1.2.3.4,foo.bar")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CONTROL_PLANE_NODE_NAMESERVERS \"google.dns,foo.bar\", expected to be IP addresses that match TKG_IP_FAMILY \"ipv4\""))
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS is a IPv6, but the TKG_IP_FAMILY is ipv4", func() {
					It("should return an error", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "2001:DB8::1")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CONTROL_PLANE_NODE_NAMESERVERS \"2001:DB8::1\", expected to be IP addresses that match TKG_IP_FAMILY \"ipv4\""))
					})
				})
				Context("when CONTROL_PLANE_NODE_NAMESERVERS is a IPv4, but the TKG_IP_FAMILY is ipv6", func() {
					It("should return an error", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableControlPlaneNodeNameservers, "8.8.8.8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid CONTROL_PLANE_NODE_NAMESERVERS \"8.8.8.8\", expected to be IP addresses that match TKG_IP_FAMILY \"ipv6\""))
					})
				})
			})
			Context("Worker Node Nameservers", func() {
				Context("Custom Nameserver feature gate is false", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(false, nil)
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "8.8.8.8")
					})

					It("should return an error", func() {
						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("option WORKER_NODE_NAMESERVERS is set to \"8.8.8.8\", but custom nameserver support is not enabled (because it is not fully functional). To enable custom nameservers, run the command: tanzu config set features.management-cluster.custom-nameservers true"))
					})
				})

				Context("when WORKER_NODE_NAMESERVERS is a valid set of IPv4 address", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "8.8.8.8,8.8.4.4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when WORKER_NODE_NAMESERVERS is a valid set of IPv6 address and TKG_IP_FAMILY is ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "2001:DB8::1, 2001:DB8::2")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when WORKER_NODE_NAMESERVERS only contains IPv6 addresses and TKG_IP_FAMILY is ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "2001:DB8::1,2001:DB8::2")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when WORKER_NODE_NAMESERVERS only contains IPv4 addresses and TKG_IP_FAMILY is ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "8.8.8.8,8.8.4.4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when WORKER_NODE_NAMESERVERS is a valid set of IPv4,IPv6 address and TKG_IP_FAMILY is ipv4,ipv6", func() {
					It("should pass validation", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "8.8.8.8,2001:DB8::2,8.8.4.4,2001:DB8::4")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
				Context("when WORKER_NODE_NAMESERVERS contains multiple invalid entries", func() {
					It("should return an error", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "google.dns,1.2.3.4,foo.bar")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid WORKER_NODE_NAMESERVERS \"google.dns,foo.bar\", expected to be IP addresses that match TKG_IP_FAMILY \"ipv4\""))
					})
				})
				Context("when WORKER_NODE_NAMESERVERS is a IPv6, but the TKG_IP_FAMILY is ipv4", func() {
					It("should return an error", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "2001:DB8::1")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid WORKER_NODE_NAMESERVERS \"2001:DB8::1\", expected to be IP addresses that match TKG_IP_FAMILY \"ipv4\""))
					})
				})
				Context("when WORKER_NODE_NAMESERVERS is a IPv4, but the TKG_IP_FAMILY is ipv6", func() {
					It("should return an error", func() {
						tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
						tkgConfigReaderWriter.Set(constants.ConfigVariableWorkerNodeNameservers, "8.8.8.8")

						validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("invalid WORKER_NODE_NAMESERVERS \"8.8.8.8\", expected to be IP addresses that match TKG_IP_FAMILY \"ipv6\""))
					})
				})
			})
		})

		Context("CoreDNSIP configuration and validation", func() {
			Context("when SERVICE_CIDR is ipv4", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4")
				})

				It("should have correct ipv4 coreDNSIP configured", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "10.64.0.1/13")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())

					coreDNSIP, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableCoreDNSIP)
					Expect(coreDNSIP).Should(Equal("10.64.0.10"))
				})

				It("should have correct ipv4 coreDNSIP configured", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "192.168.2.1/12")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())

					coreDNSIP, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableCoreDNSIP)
					Expect(coreDNSIP).Should(Equal("192.160.0.10"))
				})
			})

			Context("when SERVICE_CIDR is ipv6", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6")
				})

				It("should have correct ipv6 coreDNSIP configured", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "fd00:100:64::/108")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())

					coreDNSIP, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableCoreDNSIP)
					Expect(coreDNSIP).Should(Equal("fd00:100:64::a"))
				})

				It("should have correct ipv6 coreDNSIP configured", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())

					coreDNSIP, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableCoreDNSIP)
					Expect(coreDNSIP).Should(Equal("::a"))
				})
			})

			Context("when IPFamily is ipv4,ipv6 i.e Dual-stack Primary IPv4", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
				})

				It("should have correct ipv4 coreDNSIP configured", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "1.2.3.4/12,::1/108")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())

					coreDNSIP, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableCoreDNSIP)
					Expect(coreDNSIP).Should(Equal("1.0.0.10"))
				})
			})

			Context("when IPFamily is ipv6,ipv4 i.e Dual-stack Primary IPv6", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6,ipv4")
				})

				It("should have correct ipv6 coreDNSIP configured", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, "::1/108,1.2.3.4/12")

					validationError := tkgClient.ConfigureAndValidateManagementClusterConfiguration(initRegionOptions, true)
					Expect(validationError).NotTo(HaveOccurred())

					coreDNSIP, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableCoreDNSIP)
					Expect(coreDNSIP).Should(Equal("::a"))
				})
			})
		})

		Context("ConfigureAndValidateAviConfiguration", func() {
			var (
				mockCtrl   *gomock.Controller
				mockClient *aviMock.MockClient
			)
			BeforeEach(func() {
				mockCtrl = gomock.NewController(GinkgoT())
				mockClient = aviMock.NewMockClient(mockCtrl)
			})
			AfterEach(func() {
				mockCtrl.Finish()
			})
			When("avi is not enabled", func() {
				It("should skip avi validation", func() {
					err := tkgClient.ConfigureAndValidateAviConfiguration()
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should skip avi validation", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviEnable, "false")
					err := tkgClient.ConfigureAndValidateAviConfiguration()
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			When("validate avi account", func() {
				It("should throw error if not set AVI_CONTROLLER", func() {
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if not set AVI_USERNAME", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if not set AVI_PASSWORD", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerUsername, "test-user")
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if not set AVI_CA_DATA_B64", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerUsername, "test-user")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerPassword, "test-password")
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if AVI_CA_DATA_B64 format error", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerUsername, "test-user")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerPassword, "test-password")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerCA, "adacad")
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if call AVI controller API error", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerUsername, "test-user")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerPassword, "test-password")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerCA, "dGVzdC1jYQ==")
					aviControllerParams := &models.AviControllerParams{
						Username: "test-user",
						Password: "test-password",
						Host:     "10.10.10.1",
						Tenant:   "admin",
						CAData:   string("test-ca"),
					}
					mockClient.EXPECT().VerifyAccount(aviControllerParams).Return(false, errors.New("call avi controller api issue")).Times(1)
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if using wrong credentials", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerUsername, "test-user")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerPassword, "test-wrong-password")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerCA, "dGVzdC1jYQ==")
					aviControllerParams := &models.AviControllerParams{
						Username: "test-user",
						Password: "test-wrong-password",
						Host:     "10.10.10.1",
						Tenant:   "admin",
						CAData:   string("test-ca"),
					}
					mockClient.EXPECT().VerifyAccount(aviControllerParams).Return(false, nil).Times(1)
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should pass if provide correct configurations", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerAddress, "10.10.10.1")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerUsername, "test-user")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerPassword, "test-password")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControllerCA, "dGVzdC1jYQ==")
					aviControllerParams := &models.AviControllerParams{
						Username: "test-user",
						Password: "test-password",
						Host:     "10.10.10.1",
						Tenant:   "admin",
						CAData:   string("test-ca"),
					}
					mockClient.EXPECT().VerifyAccount(aviControllerParams).Return(true, nil).Times(1)
					err := tkgClient.ValidateAviControllerAccount(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
			})
			When("validate avi cloud", func() {
				It("should throw error if not set AVI_CLOUD_NAME", func() {
					err := tkgClient.ValidateAviCloud(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should pass if cloud exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviCloudName, "test-cloud")
					mockClient.EXPECT().GetCloudByName("test-cloud").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviCloud(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if cloud not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviCloudName, "test-cloud")
					mockClient.EXPECT().GetCloudByName("test-cloud").Return(nil, errors.New("test-cloud is not found")).Times(1)
					err := tkgClient.ValidateAviCloud(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
			When("validate avi service engine group", func() {
				It("should throw error if not set AVI_SERVICE_ENGINE_GROUP", func() {
					err := tkgClient.ValidateAviServiceEngineGroup(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should pass if service engine group exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviServiceEngineGroup, "test-seg")
					mockClient.EXPECT().GetServiceEngineGroupByName("test-seg").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviServiceEngineGroup(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if service engine group not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviServiceEngineGroup, "test-seg")
					mockClient.EXPECT().GetServiceEngineGroupByName("test-seg").Return(nil, errors.New("test-seg is not found")).Times(1)
					err := tkgClient.ValidateAviServiceEngineGroup(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
			When("validate avi management cluster service engine group", func() {
				It("should just pass if not set AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP", func() {
					err := tkgClient.ValidateAviManagementClusterServiceEngineGroup(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should pass if service engine group exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterServiceEngineGroup, "test-mc-seg")
					mockClient.EXPECT().GetServiceEngineGroupByName("test-mc-seg").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviManagementClusterServiceEngineGroup(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if service engine group not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterServiceEngineGroup, "test-mc-seg")
					mockClient.EXPECT().GetServiceEngineGroupByName("test-mc-seg").Return(nil, errors.New("test-mc-seg is not found")).Times(1)
					err := tkgClient.ValidateAviManagementClusterServiceEngineGroup(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
			When("validate avi data plane network", func() {
				It("should throw error if not set AVI_DATA_NETWORK", func() {
					err := tkgClient.ValidateAviDataPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if not set AVI_DATA_NETWORK_CIDR", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkName, "test-data-net")
					err := tkgClient.ValidateAviDataPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should pass if data plane network exists in avi controller and cidr format is valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkName, "test-data-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-data-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviDataPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if data plane network not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkName, "test-data-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-data-net").Return(nil, errors.New("test-data-net is not found")).Times(1)
					err := tkgClient.ValidateAviDataPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if data plane network CIDR format is not valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkName, "test-data-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviDataPlaneNetworkCIDR, "10.10.10/test")
					mockClient.EXPECT().GetVipNetworkByName("test-data-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviDataPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
			When("validate avi control plane network", func() {
				It("should pass if not set AVI_CONTROL_PLANE_NETWORK", func() {
					err := tkgClient.ValidateAviControlPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should pass if control plane network exists in avi controller and cidr format is valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControlPlaneNetworkName, "test-cp-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControlPlaneNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-cp-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviControlPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if control plane network not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControlPlaneNetworkName, "test-cp-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControlPlaneNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-cp-net").Return(nil, errors.New("test-cp-net is not found")).Times(1)
					err := tkgClient.ValidateAviControlPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if control plane network CIDR format is not valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControlPlaneNetworkName, "test-cp-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviControlPlaneNetworkCIDR, "10.10.10/test")
					mockClient.EXPECT().GetVipNetworkByName("test-cp-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviControlPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
			When("validate avi management cluster control plane network", func() {
				It("should pass if not set AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_NAME", func() {
					err := tkgClient.ValidateAviManagementClusterControlPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should pass if management cluster control plane network exists in avi controller and cidr format is valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName, "test-mc-cp-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-mc-cp-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviManagementClusterControlPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if management cluster control plane network not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName, "test-mc-cp-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-mc-cp-net").Return(nil, errors.New("test-mc-cp-net is not found")).Times(1)
					err := tkgClient.ValidateAviManagementClusterControlPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if management cluster control plane network CIDR format is not valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName, "test-mc-cp-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR, "10.10.10/test")
					mockClient.EXPECT().GetVipNetworkByName("test-mc-cp-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviManagementClusterControlPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
			When("validate avi management cluster data plane network", func() {
				It("should pass if not set AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_NAME", func() {
					err := tkgClient.ValidateAviManagementClusterDataPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should pass if management cluster data plane network exists in avi controller and cidr format is valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterDataPlaneNetworkName, "test-mc-data-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-mc-data-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviManagementClusterDataPlaneNetwork(mockClient)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should throw error if management cluster data plane network not exists in avi controller", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterDataPlaneNetworkName, "test-mc-data-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR, "10.10.10.1/24")
					mockClient.EXPECT().GetVipNetworkByName("test-mc-data-net").Return(nil, errors.New("test-mc-data-net is not found")).Times(1)
					err := tkgClient.ValidateAviManagementClusterDataPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
				It("should throw error if management cluster control plane network CIDR format is not valid", func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterDataPlaneNetworkName, "test-mc-data-net")
					tkgConfigReaderWriter.Set(constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR, "10.10.10/test")
					mockClient.EXPECT().GetVipNetworkByName("test-mc-data-net").Return(nil, nil).Times(1)
					err := tkgClient.ValidateAviManagementClusterDataPlaneNetwork(mockClient)
					Expect(err).Should(HaveOccurred())
				})
			})
		})
	})

	Context("ConfigureAndValidateWorkloadClusterConfiguration", func() {
		var (
			createClusterOptions *client.CreateClusterOptions
			clusterClient        *fakes.ClusterClient
		)
		BeforeEach(func() {
			workerMachineCount := int64(3)
			createClusterOptions = &client.CreateClusterOptions{
				VsphereControlPlaneEndpoint: "foo.bar",
				Edition:                     "tkg",
				ClusterConfigOptions: client.ClusterConfigOptions{
					WorkerMachineCount: &workerMachineCount,
				},
			}
			createClusterOptions.ProviderRepositorySource = &clusterctl.ProviderRepositorySourceOptions{
				InfrastructureProvider: "vsphere",
			}

			clusterClient = &fakes.ClusterClient{}
		})
		Context("IPFamily configuration and validation", func() {
			Context("when IPFamily is ipv4,ipv6 i.e Dual-stack Primary IPv4", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv4,ipv6")
				})

				Context("when dual-stack-ipv4-primary feature gate is false", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(false, nil)
					})
					It("returns an error", func() {
						validationError := tkgClient.ConfigureAndValidateWorkloadClusterConfiguration(createClusterOptions, clusterClient, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("option TKG_IP_FAMILY is set to \"ipv4,ipv6\", but dualstack support is not enabled (because it is under development). To enable dualstack, set features.cluster.dual-stack-ipv4-primary to \"true\""))
					})
				})
				Context("when dual-stack-ipv4-primary feature gate is true", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(true, nil)
					})
					It("passes validation", func() {
						validationError := tkgClient.ConfigureAndValidateWorkloadClusterConfiguration(createClusterOptions, clusterClient, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
			})
			Context("when IPFamily is ipv6,ipv4 i.e Dual-stack Primary IPv6", func() {
				BeforeEach(func() {
					tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, "ipv6,ipv4")
				})

				Context("when dual-stack-ipv6-primary feature gate is false", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(false, nil)
					})
					It("returns an error", func() {
						validationError := tkgClient.ConfigureAndValidateWorkloadClusterConfiguration(createClusterOptions, clusterClient, true)
						Expect(validationError).To(HaveOccurred())
						Expect(validationError.Error()).To(ContainSubstring("option TKG_IP_FAMILY is set to \"ipv6,ipv4\", but dualstack support is not enabled (because it is under development). To enable dualstack, set features.cluster.dual-stack-ipv6-primary to \"true\""))
					})
				})
				Context("when dual-stack-ipv6-primary feature gate is true", func() {
					BeforeEach(func() {
						featureFlagClient.IsConfigFeatureActivatedReturns(true, nil)
					})
					It("passes validation", func() {
						validationError := tkgClient.ConfigureAndValidateWorkloadClusterConfiguration(createClusterOptions, clusterClient, true)
						Expect(validationError).NotTo(HaveOccurred())
					})
				})
			})
		})

		DescribeTable("SERVICE_CIDR size validation - invalid cases", func(ipFamily, serviceCIDR, problematicCIDR, netmaskSizeConstraint string) {
			tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, ipFamily)
			tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, serviceCIDR)

			validationError := tkgClient.ConfigureAndValidateWorkloadClusterConfiguration(createClusterOptions, clusterClient, true)
			Expect(validationError).To(HaveOccurred())
			expectedErrorMsg := fmt.Sprintf(`invalid SERVICE_CIDR "%s", expected netmask to be "%s" or greater`, problematicCIDR, netmaskSizeConstraint)
			Expect(validationError.Error()).To(ContainSubstring(expectedErrorMsg))
		},
			Entry("ipv4 cidr too large", "ipv4", "192.168.2.1/11", "192.168.2.1/11", "/12"),
			Entry("ipv6 cidr too large", "ipv6", "::1/107", "::1/107", "/108"),
			Entry("ipv4-primary dualstack: ipv4 cidr too large", "ipv4,ipv6", "1.2.3.4/11,::1/108", "1.2.3.4/11", "/12"),
			Entry("ipv4-primary dualstack: ipv6 cidr too large", "ipv4,ipv6", "1.2.3.4/12,::1/107", "::1/107", "/108"),
			Entry("ipv6-primary dualstack: ipv6 cidr too large", "ipv6,ipv4", "::1/107,1.2.3.4/12", "::1/107", "/108"),
			Entry("ipv6-primary dualstack: ipv4 cidr too large", "ipv6,ipv4", "::1/108,1.2.3.4/11", "1.2.3.4/11", "/12"),
		)

		DescribeTable("SERVICE_CIDR size validation - valid cases", func(ipFamily, serviceCIDR string) {
			tkgConfigReaderWriter.Set(constants.ConfigVariableIPFamily, ipFamily)
			tkgConfigReaderWriter.Set(constants.ConfigVariableServiceCIDR, serviceCIDR)

			validationError := tkgClient.ConfigureAndValidateWorkloadClusterConfiguration(createClusterOptions, clusterClient, true)
			Expect(validationError).NotTo(HaveOccurred())
		},
			Entry("ipv4 cidr at max size", "ipv4", "192.168.2.1/12"),
			Entry("ipv4 cidr at slightly smaller size", "ipv4", "192.168.2.1/13"),
			Entry("ipv6 cidr at max size", "ipv6", "::1/108"),
			Entry("ipv6 cidr at slightly smaller size", "ipv6", "::1/108"),
			Entry("ipv4-primary dualstack: cidrs at max size", "ipv4,ipv6", "1.2.3.4/12,::1/108"),
			Entry("ipv6-primary dualstack: cidrs at max size", "ipv6,ipv4", "::1/108,1.2.3.4/12"),
		)
	})
})

var _ = Describe("Cluster Name Validation", func() {
	var (
		infrastructureProvider string
		err                    error
	)
	Context("Azure Cluster", func() {
		BeforeEach(func() {
			infrastructureProvider = "azure"
		})
		When("cluster name starts with lowercase alpha and is less than 45 characters", func() {
			It("should validate successfully", func() {
				err = client.CheckClusterNameFormat("azure-test-cluster", infrastructureProvider)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("cluster name does not start with lowercase alpha and is more than 44 characters", func() {
			It("should throw an error", func() {
				err = client.CheckClusterNameFormat("1-azure-test-cluster-with-a-really-long-name-that-is-excessive", infrastructureProvider)
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("AWS/vSphere Cluster", func() {
		BeforeEach(func() {
			infrastructureProvider = "aws"
		})
		When("cluster name starts with lowercase alphanumeric and is less than 64 characters", func() {
			It("should validate successfully", func() {
				err = client.CheckClusterNameFormat("1aws-test-cluster", infrastructureProvider)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("cluster name does not start with lowercase alphanumeric and is more than 63 characters", func() {
			It("should throw an error", func() {
				err = client.CheckClusterNameFormat("-aws-test-cluster-with-a-really-long-name-that-is-excessive-and-still-needs-to-be-longer", infrastructureProvider)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
