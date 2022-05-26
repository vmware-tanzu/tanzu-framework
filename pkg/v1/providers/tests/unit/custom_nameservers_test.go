// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

var _ = Describe("Control Plane/Workload Node Nameserver Ytt Templating", func() {
	var paths []string
	BeforeEach(func() {
		paths = []string{
			filepath.Join("fixtures", "yttmocks"),
			filepath.Join("..", "..", "infrastructure-vsphere", capvVersion, "ytt", "overlay.yaml"),
			filepath.Join("..", "..", "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
			filepath.Join("..", "..", "config_default.yaml"),
		}
	})

	When("TKG_IP_FAMILY is ipv4", func() {
		var values string
		BeforeEach(func() {
			values = createDataValues(map[string]interface{}{
				"CLUSTER_NAME":                   "foo",
				"KUBERNETES_RELEASE":             "v1.22.11---vmware.1-tkg.1",
				"TKG_CLUSTER_ROLE":               "workload",
				"TKG_IP_FAMILY":                  "ipv4",
				"CONTROL_PLANE_NODE_NAMESERVERS": "1.1.1.1,2.2.2.2",
				"WORKER_NODE_NAMESERVERS":        "3.3.3.3,4.4.4.4",
				"SERVICE_CIDR":                   "5.5.5.5/16",
				"CLUSTER_CIDR":                   "6.6.6.6/16",
			})
		})

		It("renders the control plane VSphereMachineTemplate with the custom control plane node nameservers", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplate, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "foo-control-plane",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(vsphereMachineTemplate).To(HaveLen(1))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[0]", "1.1.1.1"))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[1]", "2.2.2.2"))
		})

		It("renders the worker VSphereMachineTemplate with the custom worker node nameservers", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplate, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "foo-worker",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(vsphereMachineTemplate).To(HaveLen(1))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[0]", "3.3.3.3"))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[1]", "4.4.4.4"))
		})
	})

	When("TKG_IP_FAMILY is ipv6", func() {
		var values string
		BeforeEach(func() {
			values = createDataValues(map[string]interface{}{
				"CLUSTER_NAME":                   "foo",
				"KUBERNETES_RELEASE":             "v1.22.11---vmware.1-tkg.1",
				"TKG_CLUSTER_ROLE":               "workload",
				"TKG_IP_FAMILY":                  "ipv6",
				"CONTROL_PLANE_NODE_NAMESERVERS": "fd00::1,fd00::2",
				"WORKER_NODE_NAMESERVERS":        "fd00::3,fd00::4",
				"SERVICE_CIDR":                   "5.5.5.5/16",
				"CLUSTER_CIDR":                   "6.6.6.6/16",
			})
		})

		It("renders the control plane VSphereMachineTemplate with the custom control plane node nameservers", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplate, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "foo-control-plane",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(vsphereMachineTemplate).To(HaveLen(1))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[0]", "fd00::1"))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[1]", "fd00::2"))
		})

		It("renders the worker VSphereMachineTemplate with the custom worker node nameservers", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplate, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "foo-worker",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(vsphereMachineTemplate).To(HaveLen(1))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[0]", "fd00::3"))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[1]", "fd00::4"))
		})
	})

	When("TKG_IP_FAMILY is ipv4,ipv6", func() {
		var values string
		BeforeEach(func() {
			values = createDataValues(map[string]interface{}{
				"CLUSTER_NAME":                   "foo",
				"KUBERNETES_RELEASE":             "v1.22.11---vmware.1-tkg.1",
				"TKG_CLUSTER_ROLE":               "workload",
				"TKG_IP_FAMILY":                  "ipv4,ipv6",
				"CONTROL_PLANE_NODE_NAMESERVERS": "1.1.1.1,fd00::2",
				"WORKER_NODE_NAMESERVERS":        "3.3.3.3,fd00::4",
				"SERVICE_CIDR":                   "5.5.5.5/16",
				"CLUSTER_CIDR":                   "6.6.6.6/16",
			})
		})

		It("renders the control plane VSphereMachineTemplate with the custom control plane node nameservers", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplate, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "foo-control-plane",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(vsphereMachineTemplate).To(HaveLen(1))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[0]", "1.1.1.1"))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[1]", "fd00::2"))
		})

		It("renders the worker VSphereMachineTemplate with the custom worker node nameservers", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplate, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "foo-worker",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(vsphereMachineTemplate).To(HaveLen(1))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[0]", "3.3.3.3"))
			Expect(vsphereMachineTemplate[0]).To(HaveYAMLPathWithValue("$.spec.template.spec.network.devices[0].nameservers[1]", "fd00::4"))
		})
	})
})
