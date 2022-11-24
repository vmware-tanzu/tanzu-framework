// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

var _ = Describe("Hardware version upgrade", func() {
	var paths []string
	var baseVal yttValues

	BeforeEach(func() {
		paths = []string{
			filepath.Join(yamlRoot, "config_default.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "overlay.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
			filepath.Join("./fixtures/tkr-bom-v1.21.1.yaml"),
			filepath.Join("./fixtures/tkg-bom-v1.4.0.yaml"),
			filepath.Join(yamlRoot, "ytt"),
		}

		baseVal = map[string]interface{}{
			// required fields
			"TKG_DEFAULT_BOM":    "tkg-bom-v1.4.0.yaml",
			"KUBERNETES_RELEASE": "v1.21.2---vmware.1-tkg.1",
			"CLUSTER_NAME":       "test-cluster",

			// required fields for CAPV
			"PROVIDER_TYPE":    "vsphere",
			"TKG_CLUSTER_ROLE": "management",
			"TKG_IP_FAMILY":    "ipv4",
			"SERVICE_CIDR":     "5.5.5.5/16",

			// required vsphere configurations
			"VSPHERE_USERNAME":           "user_blah",
			"VSPHERE_PASSWORD":           "pass_1234",
			"VSPHERE_SERVER":             "vmware-tanzu.com",
			"VSPHERE_DATACENTER":         "vmware-tanzu-dc.com",
			"VSPHERE_RESOURCE_POOL":      "myrp",
			"VSPHERE_FOLDER":             "ds0",
			"VSPHERE_SSH_AUTHORIZED_KEY": "ssh-rsa AAAA...+M7Q== vmware-tanzu.local",
			"VSPHERE_INSECURE":           "true",
			"CLUSTER_CIDR":               "192.168.1.0/16",
		}
	})

	When("VSPHERE_WORKER_HARDWARE_VERSION and VSPHERE_CONTROL_PLANE_HARDWARE_VERSION are set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("sets the hardware version on the worker VSphereMachineTemplate object", func() {
			value.Set("VSPHERE_WORKER_HARDWARE_VERSION", "vmx-17")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "test-cluster-worker",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(vsphereMachineTemplates)).To(Equal(1))
			Expect(vsphereMachineTemplates[0]).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.hardwareVersion", "vmx-17"))
		})

		It("sets the hardware version on the control plane VSphereMachineTemplate object", func() {
			value.Set("VSPHERE_CONTROL_PLANE_HARDWARE_VERSION", "vmx-18")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "VSphereMachineTemplate",
				"$.metadata.name": "test-cluster-control-plane",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(vsphereMachineTemplates)).To(Equal(1))
			Expect(vsphereMachineTemplates[0]).To(matchers.HaveYAMLPathWithValue("$.spec.template.spec.hardwareVersion", "vmx-18"))
		})
	})

	When("VSPHERE_WORKER_HARDWARE_VERSION and VSPHERE_CONTROL_PLANE_HARDWARE_VERSION are unset", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("does not set the hardware version on the VSphereMachineTemplate objects", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			clusterTypes := []string{"worker", "control-plane"}
			for _, clusterType := range clusterTypes {
				vsphereMachineTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
					"$.kind":          "VSphereMachineTemplate",
					"$.metadata.name": fmt.Sprintf("test-cluster-%s", clusterType),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(len(vsphereMachineTemplates)).NotTo(Equal(0))

				for _, vsphereMachineTemplate := range vsphereMachineTemplates {
					Expect(vsphereMachineTemplate).NotTo(matchers.HaveYAMLPath("$.spec.template.spec.hardwareVersion"))
				}
			}
		})
	})
})
