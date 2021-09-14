// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/ytt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/matchers"
)

var _ = Describe("TKG_IP_FAMILY Ytt Templating", func() {
	Describe("IP family ytt validations", func() {
		var paths []string
		BeforeEach(func() {
			paths = []string{
				filepath.Join("..", "..", "config_default.yaml"),
				filepath.Join("..", "..", "ytt", "03_customizations", "ip_family.yaml"),
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

		When("a not yet implemented ip family is set", func() {
			It("does not allow ipv6,ipv4", func() {
				values := createDataValues(map[string]string{
					"TKG_IP_FAMILY": "ipv6,ipv4",
					"PROVIDER_TYPE": "vsphere",
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
				filepath.Join("..", "..", "infrastructure-vsphere", "v0.7.10", "ytt", "overlay.yaml"),
				filepath.Join("..", "..", "infrastructure-vsphere", "v0.7.10", "ytt", "base-template.yaml"),
				filepath.Join("..", "..", "config_default.yaml"),
			}
		})

		It("renders control plane and worker VSphereMachineTemplates with ipv4 single stack settings", func() {
			values := createDataValues(map[string]string{
				"CLUSTER_NAME":     "foo",
				"TKG_CLUSTER_ROLE": "workload",
				"TKG_IP_FAMILY":    "ipv4",
			})
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

		It("renders control plane and worker VSphereMachineTemplates with ipv6 single stack settings", func() {
			values := createDataValues(map[string]string{
				"CLUSTER_NAME":     "foo",
				"TKG_CLUSTER_ROLE": "workload",
				"TKG_IP_FAMILY":    "ipv6",
			})
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

		It("renders control plane and worker VSphereMachineTemplates with ipv4,ipv6 dual stack settings", func() {
			values := createDataValues(map[string]string{
				"CLUSTER_NAME":     "foo",
				"TKG_CLUSTER_ROLE": "workload",
				"TKG_IP_FAMILY":    "ipv4,ipv6",
			})
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

		It("renders control plane and worker VSphereMachineTemplates with ipv6,ipv4 dual stack settings", func() {
			values := createDataValues(map[string]string{
				"CLUSTER_NAME":     "foo",
				"TKG_CLUSTER_ROLE": "workload",
				"TKG_IP_FAMILY":    "ipv6,ipv4",
			})
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
