// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/matchers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/ytt"
)

const (
	capvVersion = "v1.0.2"
	yamlRoot    = "../../"
)

var _ = Describe("TKG_IP_FAMILY Ytt Templating", func() {
	Describe("IP family ytt validations", func() {
		var paths []string
		BeforeEach(func() {
			paths = []string{
				// assumes that ../../ is where the yaml templates live
				filepath.Join(yamlRoot, "config_default.yaml"),
				filepath.Join(yamlRoot, "ytt", "03_customizations", "ip_family.yaml"),
			}
		})

		It("allows undefined", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "",
			})
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows ipv4", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "ipv4",
			})
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
		})

		When("an unsupported ip family is set", func() {
			It("does not allow dual", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "dual",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does not allow garbage", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "garbage",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
		})

		When("provider type is vsphere", func() {
			It("allows ipv6", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv6",
					"PROVIDER_TYPE": "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows ipv4,ipv6", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv4,ipv6",
					"PROVIDER_TYPE": "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows ipv6,ipv4", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv6,ipv4",
					"PROVIDER_TYPE": "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("provider type is not vsphere", func() {
			It("does not allow ipv6", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv6",
					"PROVIDER_TYPE": "aws",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does not allow ipv4,ipv6", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv4,ipv6",
					"PROVIDER_TYPE": "azure",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does not allow ipv6,ipv4", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv6,ipv4",
					"PROVIDER_TYPE": "azure",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
		})

		When("workload cluster is windows on vsphere", func() {
			It("does not allow ipv6", func() {
				values := createDataValues(map[string]string{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv6",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(MatchError(ContainSubstring("IS_WINDOWS_WORKLOAD_CLUSTER is not compatible with TKG_IP_FAMLY values of \"ipv6\", \"ipv4,ipv6\" or \"ipv6,ipv4\"")))
			})
			It("allows ipv4", func() {
				values := createDataValues(map[string]string{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv4",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
			It("does not allow ipv4,ipv6", func() {
				values := createDataValues(map[string]string{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv4,ipv6",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(MatchError(ContainSubstring("IS_WINDOWS_WORKLOAD_CLUSTER is not compatible with TKG_IP_FAMLY values of \"ipv6\", \"ipv4,ipv6\" or \"ipv6,ipv4\"")))
			})
			It("does not allow ipv6,ipv4", func() {
				values := createDataValues(map[string]string{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv4,ipv6",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(MatchError(ContainSubstring("IS_WINDOWS_WORKLOAD_CLUSTER is not compatible with TKG_IP_FAMLY values of \"ipv6\", \"ipv4,ipv6\" or \"ipv6,ipv4\"")))
			})
		})
	})
	Describe("antrea_addon_data.lib.yaml", func() {
		var paths []string
		BeforeEach(func() {
			paths = []string{
				filepath.Join("fixtures", "yttmocks"),
				filepath.Join("fixtures", "antrea_data_values.yaml"),
				filepath.Join("..", "..", "ytt", "02_addons", "cni", "antrea", "antrea_addon_data.lib.yaml"),
				filepath.Join("..", "..", "config_default.yaml"),
			}
		})

		It("renders antrea yaml with ipv4,ipv6 dual stack settings", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "ipv4,ipv6",
				"SERVICE_CIDR":  "1.2.3.4/16,fd00::/48",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDRv6", "fd00::/48"))
		})

		It("renders antrea yaml with ipv6,ipv4 dual stack settings", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "ipv6,ipv4",
				"SERVICE_CIDR":  "fd00::/48,1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDRv6", "fd00::/48"))
		})

		It("renders antrea yaml with ipv4 single stack settings", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "ipv4",
				"SERVICE_CIDR":  "1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})

		It("renders antrea yaml with ipv6 single stack settings", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "ipv6",
				"SERVICE_CIDR":  "fd00::/48",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "fd00::/48"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})

		It("renders antrea yaml with ipv4 single stack settings with undefined TKG_IP_FAMILY", func() {
			values := createDataValues(map[string]string{
				"SERVICE_CIDR": "1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})

		It("renders antrea yaml with ipv4 single stack settings with an empty TKG_IP_FAMILY", func() {
			values := createDataValues(map[string]string{
				"TKG_IP_FAMILY": "",
				"SERVICE_CIDR":  "1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})
	})

	Describe("infrastructure-vsphere overlay.yaml", func() {
		var paths []string
		BeforeEach(func() {
			paths = []string{
				filepath.Join("fixtures", "yttmocks"),
				filepath.Join("..", "..", "infrastructure-vsphere", capvVersion, "ytt", "overlay.yaml"),
				filepath.Join("..", "..", "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
				filepath.Join("..", "..", "config_default.yaml"),
			}
		})

		Describe("cluster cidr blocks", func() {
			var values string
			When("cluster cidr and service cidr have multiple values", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":     "foo",
						"TKG_CLUSTER_ROLE": "workload",
						"TKG_IP_FAMILY":    "ipv4,ipv6",
						"CLUSTER_CIDR":     "100.96.0.0/11,fd00:100:96::/48",
						"SERVICE_CIDR":     "100.64.0.0/18,fd00:100:64::/108",
					})
				})

				It("renders the cluster with the pod and service cidrs with dual stack settings", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					clusterDoc, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "Cluster",
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(clusterDoc).To(HaveLen(1))
					Expect(clusterDoc[0]).To(HaveYAMLPathWithValue("$.spec.clusterNetwork.pods.cidrBlocks[0]", "100.96.0.0/11"))
					Expect(clusterDoc[0]).To(HaveYAMLPathWithValue("$.spec.clusterNetwork.pods.cidrBlocks[1]", "fd00:100:96::/48"))
					Expect(clusterDoc[0]).To(HaveYAMLPathWithValue("$.spec.clusterNetwork.services.cidrBlocks[0]", "100.64.0.0/18"))
					Expect(clusterDoc[0]).To(HaveYAMLPathWithValue("$.spec.clusterNetwork.services.cidrBlocks[1]", "fd00:100:64::/108"))
				})
			})
			When("cluster cidr and service cidr have a single value", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":     "foo",
						"TKG_CLUSTER_ROLE": "workload",
						"TKG_IP_FAMILY":    "ipv4",
						"CLUSTER_CIDR":     "100.96.0.0/11",
						"SERVICE_CIDR":     "100.64.0.0/18",
					})
				})

				It("renders the cluster with the pod and service cidrs with single stack settings", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					clusterDoc, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "Cluster",
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(clusterDoc).To(HaveLen(1))
					Expect(clusterDoc[0]).To(HaveYAMLPathWithValue("$.spec.clusterNetwork.pods.cidrBlocks[0]", "100.96.0.0/11"))
					Expect(clusterDoc[0]).NotTo(HaveYAMLPath("$.spec.clusterNetwork.pods.cidrBlocks[1]"))
					Expect(clusterDoc[0]).To(HaveYAMLPathWithValue("$.spec.clusterNetwork.services.cidrBlocks[0]", "100.64.0.0/18"))
					Expect(clusterDoc[0]).NotTo(HaveYAMLPath("$.spec.clusterNetwork.services.cidrBlocks[1]"))
				})
			})
		})

		Describe("vsphere machine templates", func() {
			var values string
			When("data values are set to single stack IPv4 settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":            "foo",
						"TKG_CLUSTER_ROLE":        "workload",
						"TKG_IP_FAMILY":           "ipv4",
						"CLUSTER_CIDR":            "100.96.0.0/11",
						"SERVICE_CIDR":            "100.64.0.0/18",
						"CLUSTER_API_SERVER_PORT": "443",
					})
				})
				It("renders control plane and worker templates each with an ipv4 single stack network device", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					vsphereMachineTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "VSphereMachineTemplate",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(vsphereMachineTemplateDocs).To(HaveLen(2))
					for _, machineDoc := range vsphereMachineTemplateDocs {
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].dhcp4", "true"))
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].networkName", "VM Network"))
						Expect(machineDoc).NotTo(HaveYAMLPath("$.spec.template.spec.network.devices[0].dhcp6"))
						Expect(machineDoc).NotTo(HaveYAMLPath("$.spec.template.spec.network.devices[1]"))
					}
				})
				It("renders control plane template to bind the local apiServer endpoint to '0.0.0.0' as the node IP and port to custom api server port configured", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.advertiseAddress", "0.0.0.0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.bindPort", "443"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.advertiseAddress", "0.0.0.0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.bindPort", "443"))
				})

			})

			When("data values are set to single stack IPv6 settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":            "foo",
						"TKG_CLUSTER_ROLE":        "workload",
						"TKG_IP_FAMILY":           "ipv6",
						"CLUSTER_CIDR":            "fd00:100:96::/48",
						"SERVICE_CIDR":            "fd00:100:64::/108",
						"CLUSTER_API_SERVER_PORT": "443",
					})
				})
				It("renders control plane and worker templates each with an ipv6 single stack network device", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					vsphereMachineTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "VSphereMachineTemplate",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(vsphereMachineTemplateDocs).To(HaveLen(2))
					for _, machineDoc := range vsphereMachineTemplateDocs {
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].dhcp6", "true"))
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].networkName", "VM Network"))
						Expect(machineDoc).NotTo(HaveYAMLPath("$.spec.template.spec.network.devices[0].dhcp4"))
						Expect(machineDoc).NotTo(HaveYAMLPath("$.spec.template.spec.network.devices[1]"))
					}
				})
				It("renders control plane template to bind the local apiServer endpoint to '::/0' as the node IP and port to custom api server port configured", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.advertiseAddress", "::/0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.bindPort", "443"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.advertiseAddress", "::/0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.bindPort", "443"))
				})
			})

			When("data values are set to ipv4,ipv6 dual stack settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":            "foo",
						"TKG_CLUSTER_ROLE":        "workload",
						"TKG_IP_FAMILY":           "ipv4,ipv6",
						"CLUSTER_CIDR":            "100.96.0.0/11,fd00:100:96::/48",
						"SERVICE_CIDR":            "100.64.0.0/18,fd00:100:64::/108",
						"CLUSTER_API_SERVER_PORT": "443",
					})
				})
				It("renders control plane and worker templates each with a dual stack network device", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					vsphereMachineTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "VSphereMachineTemplate",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(vsphereMachineTemplateDocs).To(HaveLen(2))
					for _, machineDoc := range vsphereMachineTemplateDocs {
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].dhcp4", "true"))
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].dhcp6", "true"))
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].networkName", "VM Network"))
						Expect(machineDoc).NotTo(HaveYAMLPath("$.spec.template.spec.network.devices[1]"))
					}
				})
				It("renders control plane template to bind the local apiServer endpoint to '0.0.0.0' as the node IP and port to custom api server port configured", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.advertiseAddress", "0.0.0.0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.bindPort", "443"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.advertiseAddress", "0.0.0.0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.bindPort", "443"))
				})
				It("does not render bind-address field", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraArgs.bind-address"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.clusterConfiguration.controllerManager.extraArgs.bind-address"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.clusterConfiguration.scheduler.extraArgs.bind-address"))
				})
				It("does not render advertise-address field", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraArgs.advertise-address"))
				})
				It("does not render node-ip field", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.initConfiguration.nodeRegistration.kubeletExtraArgs.node-ip"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.joinConfiguration.nodeRegistration.kubeletExtraArgs.node-ip"))

					kubeadmConfigTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmConfigTemplate",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmConfigTemplateDocs).To(HaveLen(1))
					Expect(kubeadmConfigTemplateDocs[0]).NotTo(HaveYAMLPath("$.spec.template.spec.joinConfiguration.nodeRegistration.kubeletExtraArgs.node-ip"))
				})
			})

			When("data values are set to ipv6,ipv4 dual stack settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":                   "foo",
						"TKG_CLUSTER_ROLE":               "workload",
						"TKG_IP_FAMILY":                  "ipv6,ipv4",
						"CLUSTER_CIDR":                   "fd00:100:96::/48,100.96.0.0/11",
						"SERVICE_CIDR":                   "fd00:100:64::/108,100.64.0.0/18",
						"VSPHERE_CONTROL_PLANE_ENDPOINT": "2001:db8::2",
						"CLUSTER_API_SERVER_PORT":        "443",
					})
				})
				It("renders a control plane and worker template each with a dual stack network device", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					vsphereMachineTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "VSphereMachineTemplate",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(vsphereMachineTemplateDocs).To(HaveLen(2))
					for _, machineDoc := range vsphereMachineTemplateDocs {
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].dhcp4", "true"))
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].dhcp6", "true"))
						Expect(machineDoc).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].networkName", "VM Network"))
						Expect(machineDoc).NotTo(HaveYAMLPath("$.spec.template.spec.network.devices[1]"))
					}
				})
				It("renders control plane template to bind the apiServer, controllerManager, and scheduler to all ipv6 interfaces", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraArgs.bind-address", "::"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.clusterConfiguration.controllerManager.extraArgs.bind-address", "::"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.clusterConfiguration.scheduler.extraArgs.bind-address", "::"))
				})
				It("renders control plane template to bind the local apiServer endpoint to '::/0' as the node IP and port to custom api server port configured", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.advertiseAddress", "::/0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.bindPort", "443"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.advertiseAddress", "::/0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.bindPort", "443"))
				})
				It("renders control plane template to advertise control plane endpoint on the apiServer", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraArgs.advertise-address", "2001:db8::2"))
				})
				It("renders control plane template to configure kubelet with '::' as the node IP", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.nodeRegistration.kubeletExtraArgs.node-ip", "::"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.nodeRegistration.kubeletExtraArgs.node-ip", "::"))

					kubeadmConfigTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmConfigTemplate",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmConfigTemplateDocs).To(HaveLen(1))
					Expect(kubeadmConfigTemplateDocs[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.joinConfiguration.nodeRegistration.kubeletExtraArgs.node-ip", "::"))
				})
			})
		})

	})
	Describe("infrastructure-vsphere windows overlay.yaml", func() {
		var paths []string
		BeforeEach(func() {
			paths = []string{
				filepath.Join("fixtures", "yttmocks"),
				filepath.Join("..", "..", "infrastructure-vsphere", capvVersion, "ytt", "overlay-windows.yaml"),
				filepath.Join("..", "..", "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
				filepath.Join("..", "..", "config_default.yaml"),
			}
		})

		Describe("cluster api server port configuration", func() {
			var values string
			When("ip family is configured to ipv4", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]string{
						"CLUSTER_NAME":                "foo",
						"TKG_CLUSTER_ROLE":            "workload",
						"TKG_IP_FAMILY":               "ipv4",
						"CLUSTER_CIDR":                "10.0.0.0/16",
						"SERVICE_CIDR":                "10.0.0.0/18",
						"CLUSTER_API_SERVER_PORT":     "443",
						"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					})
				})

				It("renders control plane template to bind the local apiServer endpoint to '0.0.0.0' as the node IP and port to custom api server port configured", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.advertiseAddress", "0.0.0.0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.initConfiguration.localAPIEndpoint.bindPort", "443"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.advertiseAddress", "0.0.0.0"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.joinConfiguration.controlPlane.localAPIEndpoint.bindPort", "443"))
				})
			})
		})
	})

	Describe("vsphere cpi", func() {
		var paths []string

		Describe("vsphere cpi data values", func() {
			var ipFamilyPath = "$.data.vsphereCPI.ipFamily"
			BeforeEach(func() {
				paths = []string{
					filepath.Join("fixtures", "yttmocks"),
					filepath.Join("fixtures", "vsphere_cpi_ip_family.yaml"),
					filepath.Join("..", "..", "ytt", "02_addons", "cpi", "cpi_addon_data.lib.yaml"),
					filepath.Join("..", "..", "config_default.yaml"),
				}
			})
			When("TKG_IP_FAMILY is unset", func() {
				It("does not configure the CPI ip family", func() {
					values := createDataValues(map[string]string{
						"PROVIDER_TYPE": "vsphere",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).NotTo(HaveYAMLPath(ipFamilyPath))
				})
			})
			When("TKG_IP_FAMILY is ipv4", func() {
				It("configure the CPI for ipv4 only", func() {
					values := createDataValues(map[string]string{
						"PROVIDER_TYPE": "vsphere",
						"TKG_IP_FAMILY": "ipv4",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).To(HaveYAMLPathWithValue(ipFamilyPath, "ipv4"))
				})
			})
			When("TKG_IP_FAMILY is ipv6", func() {
				It("configure the CPI for ipv6 only", func() {
					values := createDataValues(map[string]string{
						"PROVIDER_TYPE": "vsphere",
						"TKG_IP_FAMILY": "ipv6",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).To(HaveYAMLPathWithValue(ipFamilyPath, "ipv6"))
				})
			})
			When("TKG_IP_FAMILY is ipv4,ipv6", func() {
				It("configure the CPI for ipv4 only", func() {
					values := createDataValues(map[string]string{
						"PROVIDER_TYPE": "vsphere",
						"TKG_IP_FAMILY": "ipv4,ipv6",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					// TODO: change this to ipv4,ipv6 once this issue is resolved:
					// https://github.com/kubernetes/cloud-provider-vsphere/issues/302
					Expect(output).To(HaveYAMLPathWithValue(ipFamilyPath, "ipv4"))
				})
			})
		})
	})
})

func createDataValues(values map[string]string) string {
	dataValues := "#@data/values\n---\n"
	for k, v := range values {
		dataValues += fmt.Sprintf("%s: %s\n", k, v)
	}
	return dataValues
}
