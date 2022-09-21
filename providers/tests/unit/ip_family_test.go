// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	. "github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

const (
	capvVersion = "v1.3.1"
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
			values := createDataValues(map[string]interface{}{
				"TKG_IP_FAMILY": "",
			})
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows ipv4", func() {
			values := createDataValues(map[string]interface{}{
				"TKG_IP_FAMILY": "ipv4",
			})
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
		})

		When("an unsupported ip family is set", func() {
			It("does not allow dual", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "dual",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does not allow garbage", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "garbage",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
		})

		When("provider type is vsphere", func() {
			It("allows ipv6", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "ipv6",
					"PROVIDER_TYPE": "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows ipv4,ipv6", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "ipv4,ipv6",
					"PROVIDER_TYPE": "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows ipv6,ipv4", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "ipv6,ipv4",
					"PROVIDER_TYPE": "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("provider type is not vsphere", func() {
			It("does not allow ipv6", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "ipv6",
					"PROVIDER_TYPE": "aws",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does not allow ipv4,ipv6", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "ipv4,ipv6",
					"PROVIDER_TYPE": "azure",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does not allow ipv6,ipv4", func() {
				values := createDataValues(map[string]interface{}{
					"TKG_IP_FAMILY": "ipv6,ipv4",
					"PROVIDER_TYPE": "azure",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
		})

		When("workload cluster is windows on vsphere", func() {
			It("does not allow ipv6", func() {
				values := createDataValues(map[string]interface{}{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv6",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(MatchError(ContainSubstring("IS_WINDOWS_WORKLOAD_CLUSTER is not compatible with TKG_IP_FAMLY values of \"ipv6\", \"ipv4,ipv6\" or \"ipv6,ipv4\"")))
			})
			It("allows ipv4", func() {
				values := createDataValues(map[string]interface{}{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv4",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
			It("does not allow ipv4,ipv6", func() {
				values := createDataValues(map[string]interface{}{
					"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
					"TKG_IP_FAMILY":               "ipv4,ipv6",
					"PROVIDER_TYPE":               "vsphere",
				})
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(MatchError(ContainSubstring("IS_WINDOWS_WORKLOAD_CLUSTER is not compatible with TKG_IP_FAMLY values of \"ipv6\", \"ipv4,ipv6\" or \"ipv6,ipv4\"")))
			})
			It("does not allow ipv6,ipv4", func() {
				values := createDataValues(map[string]interface{}{
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
			values := createDataValues(map[string]interface{}{
				"TKG_IP_FAMILY": "ipv4,ipv6",
				"SERVICE_CIDR":  "1.2.3.4/16,fd00::/48",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDRv6", "fd00::/48"))
		})

		It("renders antrea yaml with ipv6,ipv4 dual stack settings", func() {
			values := createDataValues(map[string]interface{}{
				"TKG_IP_FAMILY": "ipv6,ipv4",
				"SERVICE_CIDR":  "fd00::/48,1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDRv6", "fd00::/48"))
		})

		It("renders antrea yaml with ipv4 single stack settings", func() {
			values := createDataValues(map[string]interface{}{
				"TKG_IP_FAMILY": "ipv4",
				"SERVICE_CIDR":  "1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})

		It("renders antrea yaml with ipv6 single stack settings", func() {
			values := createDataValues(map[string]interface{}{
				"TKG_IP_FAMILY": "ipv6",
				"SERVICE_CIDR":  "fd00::/48",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "fd00::/48"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})

		It("renders antrea yaml with ipv4 single stack settings with undefined TKG_IP_FAMILY", func() {
			values := createDataValues(map[string]interface{}{
				"SERVICE_CIDR": "1.2.3.4/16",
			})
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveYAMLPathWithValue("$.data.antrea.config.serviceCIDR", "1.2.3.4/16"))
			Expect(output).NotTo(HaveYAMLPath("$.data.antrea.config.serviceCIDRv6"))
		})

		It("renders antrea yaml with ipv4 single stack settings with an empty TKG_IP_FAMILY", func() {
			values := createDataValues(map[string]interface{}{
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
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":       "foo",
						"KUBERNETES_RELEASE": "v1.22.11---vmware.1-tkg.1",
						"TKG_CLUSTER_ROLE":   "workload",
						"TKG_IP_FAMILY":      "ipv4,ipv6",
						"CLUSTER_CIDR":       "100.96.0.0/11,fd00:100:96::/48",
						"SERVICE_CIDR":       "100.64.0.0/18,fd00:100:64::/108",
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
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":       "foo",
						"KUBERNETES_RELEASE": "v1.22.11---vmware.1-tkg.1",
						"TKG_CLUSTER_ROLE":   "workload",
						"TKG_IP_FAMILY":      "ipv4",
						"CLUSTER_CIDR":       "100.96.0.0/11",
						"SERVICE_CIDR":       "100.64.0.0/18",
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
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":            "foo",
						"KUBERNETES_RELEASE":      "v1.22.11---vmware.1-tkg.1",
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
				It("does not configure node ip in KUBELET_EXTRA_ARGS in /etc/sysconfig/kubelet", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.files[1]"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.preKubeadmCommands[5]"))
				})
			})

			When("data values are set to single stack IPv6 settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":            "foo",
						"KUBERNETES_RELEASE":      "v1.22.11---vmware.1-tkg.1",
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
				DescribeTable("configures node-ip on the control plane nodes by echoing the detected node ip into KUBELET_EXTRA_ARGS in /etc/sysconfig/kubelet when the tkr is >= 1.22.8", func(kubernetesRelease string) {
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":            "foo",
						"KUBERNETES_RELEASE":      kubernetesRelease,
						"TKG_CLUSTER_ROLE":        "workload",
						"TKG_IP_FAMILY":           "ipv6",
						"CLUSTER_CIDR":            "fd00:100:96::/48",
						"SERVICE_CIDR":            "fd00:100:64::/108",
						"CLUSTER_API_SERVER_PORT": "443",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].content", ""))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].owner", "root:root"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].path", "/etc/sysconfig/kubelet"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].permissions", "0640"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.preKubeadmCommands[5]", "echo \"KUBELET_EXTRA_ARGS=--node-ip=$(ip -6 -json addr show dev eth0 scope global | jq -r .[0].addr_info[0].local)\" >> /etc/sysconfig/kubelet"))
				},
					Entry("when the tkr is 1.22.8", "v1.22.8---vmware.1-tkg.1"),
					Entry("when the tkr is 1.22.11", "v1.22.11---vmware.1-tkg.1"),
					Entry("when the tkr is 1.22.12", "v1.22.12---vmware.1-tkg.1"),
					Entry("when the tkr is 1.22.40", "v1.22.40---vmware.1-tkg.1"),
					Entry("when the tkr is 1.23.7", "v1.23.7---vmware.1-tkg.1"),
					Entry("when the tkr is 1.23.10", "v1.23.10---vmware.1-tkg.1"),
				)
				DescribeTable("does not configure node-ip on the control plane into KUBELET_EXTRA_ARGS in /etc/sysconfig/kubelet when the tkr is < 1.22.8", func(kubernetesRelease string) {
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":            "foo",
						"KUBERNETES_RELEASE":      kubernetesRelease,
						"TKG_CLUSTER_ROLE":        "workload",
						"TKG_IP_FAMILY":           "ipv6",
						"CLUSTER_CIDR":            "fd00:100:96::/48",
						"SERVICE_CIDR":            "fd00:100:64::/108",
						"CLUSTER_API_SERVER_PORT": "443",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.files[1]"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(ContainSubstring("KUBELET_EXTRA_ARGS=--node-ip="))
				},
					Entry("when the tkr is 1.22.7", "v1.22.7---vmware.1-tkg.1"),
					Entry("when the tkr is 1.21.20", "v1.21.20---vmware.1-tkg.1"),
					Entry("when the tkr is 1.21.8", "v1.21.8---vmware.1-tkg.1"),
					Entry("when the tkr is 1.20.15", "v1.20.15---vmware.1-tkg.1"),
					Entry("when the tkr is 1.20.3", "v1.20.3---vmware.1-tkg.1"),
					Entry("when the tkr is 1.6.0", "v1.6.0---vmware.1-tkg.1"),
				)
			})

			When("data values are set to ipv4,ipv6 dual stack settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":            "foo",
						"KUBERNETES_RELEASE":      "v1.22.11---vmware.1-tkg.1",
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
				It("does not configure node ip in KUBELET_EXTRA_ARGS in /etc/sysconfig/kubelet", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.files[1]"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.preKubeadmCommands[5]"))
				})
			})

			When("data values are set to ipv6,ipv4 dual stack settings", func() {
				BeforeEach(func() {
					values = createDataValues(map[string]interface{}{
						"CLUSTER_NAME":                   "foo",
						"KUBERNETES_RELEASE":             "v1.22.11---vmware.1-tkg.1",
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
				It("does not render node-ip field for control plane nodes", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.initConfiguration.nodeRegistration.kubeletExtraArgs.node-ip"))
					Expect(kubeadmControlPlaneDocs[0]).NotTo(HaveYAMLPath("$.spec.kubeadmConfigSpec.joinConfiguration.nodeRegistration.kubeletExtraArgs.node-ip"))
				})
				It("configures kubelet on the worker nodes with '::' as the node IP", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmConfigTemplateDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmConfigTemplate",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmConfigTemplateDocs).To(HaveLen(1))
					Expect(kubeadmConfigTemplateDocs[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.joinConfiguration.nodeRegistration.kubeletExtraArgs.node-ip", "::"))
				})
				It("configures node-ip on the control plane nodes by echoing the detected node ip into KUBELET_EXTRA_ARGS in /etc/sysconfig/kubelet", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					kubeadmControlPlaneDocs, err := FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind": "KubeadmControlPlane",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(kubeadmControlPlaneDocs).To(HaveLen(1))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].content", ""))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].owner", "root:root"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].path", "/etc/sysconfig/kubelet"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.files[1].permissions", "0640"))
					Expect(kubeadmControlPlaneDocs[0]).To(HaveYAMLPathWithValue("$.spec.kubeadmConfigSpec.preKubeadmCommands[5]", "echo \"KUBELET_EXTRA_ARGS=--node-ip=$(ip -6 -json addr show dev eth0 scope global | jq -r .[0].addr_info[0].local)\" >> /etc/sysconfig/kubelet"))
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
					values = createDataValues(map[string]interface{}{
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
					values := createDataValues(map[string]interface{}{
						"PROVIDER_TYPE": "vsphere",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).NotTo(HaveYAMLPath(ipFamilyPath))
				})
			})
			When("TKG_IP_FAMILY is ipv4", func() {
				It("configure the CPI for ipv4 only", func() {
					values := createDataValues(map[string]interface{}{
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
					values := createDataValues(map[string]interface{}{
						"PROVIDER_TYPE": "vsphere",
						"TKG_IP_FAMILY": "ipv6",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).To(HaveYAMLPathWithValue(ipFamilyPath, "ipv6"))
				})
			})
			When("TKG_IP_FAMILY is ipv4,ipv6", func() {
				It("configure the CPI for ipv4 and ipv6", func() {
					values := createDataValues(map[string]interface{}{
						"PROVIDER_TYPE": "vsphere",
						"TKG_IP_FAMILY": "ipv4,ipv6",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).To(HaveYAMLPathWithValue(ipFamilyPath, "ipv4,ipv6"))
				})
			})
			When("TKG_IP_FAMILY is ipv6,ipv4", func() {
				It("configure the CPI for ipv6 and ipv4", func() {
					values := createDataValues(map[string]interface{}{
						"PROVIDER_TYPE": "vsphere",
						"TKG_IP_FAMILY": "ipv6,ipv4",
					})
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
					Expect(err).NotTo(HaveOccurred())

					Expect(output).To(HaveYAMLPathWithValue(ipFamilyPath, "ipv6,ipv4"))
				})
			})
			Context("exclude the vsphere control plane endpoint from node ip selection", func() {
				var excludeInternalNetworkSubnetCidr = "$.data.vsphereCPI.vmExcludeInternalNetworkSubnetCidr"
				var excludeExternalNetworkSubnetCidr = "$.data.vsphereCPI.vmExcludeExternalNetworkSubnetCidr"
				When("VSPHERE_CONTROL_PLANE_ENDPOINT is ipv4", func() {
					It("excludes it as a CIDR from both external and internal node ip selection", func() {
						values := createDataValues(map[string]interface{}{
							"PROVIDER_TYPE":                  "vsphere",
							"TKG_IP_FAMILY":                  "ipv4",
							"VSPHERE_CONTROL_PLANE_ENDPOINT": "192.168.0.1",
						})

						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
						Expect(err).NotTo(HaveOccurred())

						Expect(output).To(HaveYAMLPathWithValue(excludeInternalNetworkSubnetCidr, "192.168.0.1/32"))
						Expect(output).To(HaveYAMLPathWithValue(excludeExternalNetworkSubnetCidr, "192.168.0.1/32"))
					})
				})
				When("VSPHERE_CONTROL_PLANE_ENDPOINT is ipv6", func() {
					It("excludes it as a CIDR from both external and internal node ip selection", func() {
						values := createDataValues(map[string]interface{}{
							"PROVIDER_TYPE":                  "vsphere",
							"TKG_IP_FAMILY":                  "ipv6",
							"VSPHERE_CONTROL_PLANE_ENDPOINT": "fd00:100:64::1",
						})

						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
						Expect(err).NotTo(HaveOccurred())

						Expect(output).To(HaveYAMLPathWithValue(excludeInternalNetworkSubnetCidr, "fd00:100:64::1/128"))
						Expect(output).To(HaveYAMLPathWithValue(excludeExternalNetworkSubnetCidr, "fd00:100:64::1/128"))
					})
				})
				When("VSPHERE_CONTROL_PLANE_ENDPOINT is a hostname", func() {
					It("excludes no ips from internal and external node ip selection", func() {
						values := createDataValues(map[string]interface{}{
							"PROVIDER_TYPE":                  "vsphere",
							"TKG_IP_FAMILY":                  "ipv6",
							"VSPHERE_CONTROL_PLANE_ENDPOINT": "cluster.local",
						})

						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
						Expect(err).NotTo(HaveOccurred())

						Expect(output).To(HaveYAMLPathWithValue(excludeInternalNetworkSubnetCidr, ""))
						Expect(output).To(HaveYAMLPathWithValue(excludeExternalNetworkSubnetCidr, ""))
					})
				})
			})
		})
	})
})

func createDataValues(values map[string]interface{}) string {
	dataValues := "#@data/values\n---\n"
	bytes, err := yaml.Marshal(values)
	if err != nil {
		return ""
	}
	valuesStr := string(bytes)
	valuesStr = strings.ReplaceAll(valuesStr, "\"true\"", "true")
	valuesStr = strings.ReplaceAll(valuesStr, "\"false\"", "false")
	return dataValues + valuesStr
}
