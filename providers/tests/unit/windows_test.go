// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

var _ = Describe("Windows Ytt Templating", func() {
	var paths []string
	var baseVal yttValues

	BeforeEach(func() {
		paths = []string{
			filepath.Join(yamlRoot, "config_default.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "overlay-windows.yaml"),
			filepath.Join("./fixtures/tkr-bom-v1.21.1.yaml"),
			filepath.Join("./fixtures/tkg-bom-v1.4.0.yaml"),
			filepath.Join(yamlRoot, "ytt"),
		}

		baseVal = map[string]interface{}{
			"CLUSTER_NAME":                "win-wl",
			"CORE_DNS_IP":                 "10.64.0.10",
			"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
			"KUBERNETES_RELEASE":          "v1.21.2---vmware.1-tkg.1",
			"PROVIDER_TYPE":               "vsphere",
			"TKG_CLUSTER_ROLE":            "workload",
			"TKG_DEFAULT_BOM":             "tkg-bom-v1.4.0.yaml",
			"WORKER_MACHINE_COUNT":        3,
			"VSPHERE_DATACENTER":          "vmware-tanzu-dc.com",
			"VSPHERE_FOLDER":              "ds0",
			"VSPHERE_INSECURE":            "true",
			"VSPHERE_PASSWORD":            "pass_1234",
			"VSPHERE_RESOURCE_POOL":       "myrp",
			"VSPHERE_SERVER":              "vmware-tanzu.com",
			"VSPHERE_SSH_AUTHORIZED_KEY":  "ssh-rsa AAAA...+M7Q== vmware-tanzu.local",
			"VSPHERE_USERNAME":            "user_blah",
		}
	})

	When("Using the default configuration", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("has default Cluster vSphere objects", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			tests := []struct {
				kind    string
				counter int
			}{
				{
					kind:    "Cluster",
					counter: 1,
				},
				{
					kind:    "VSphereCluster",
					counter: 1,
				},
				{
					kind:    "VSphereMachineTemplate",
					counter: 2,
				},
			}
			for _, tt := range tests {
				templates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{"$.kind": tt.kind})
				Expect(err).NotTo(HaveOccurred())
				Expect(len(templates)).To(Equal(tt.counter))
			}
		})
	})

	When("WORKER_MACHINE_COUNT is set", func() {
		var value yttValues
		var workloadCount = 6

		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("succeeds when using Windows nodes in the workload", func() {
			value.Set("WORKER_MACHINE_COUNT", workloadCount)
			value.Set("CLUSTER_PLAN", "prod")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			machineDeploymentTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{"$.kind": "MachineDeployment"})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(machineDeploymentTemplates)).To(Equal(3))
		})
	})

	When("CLUSTER_PLAN is set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("succeeds when plan is prod and add cluster annotation", func() {
			value.Set("CLUSTER_PLAN", "prod")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			clusterTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "Cluster",
				"$.metadata.name": "win-wl",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(clusterTemplates)).NotTo(Equal(0))

			for _, clusterTemplate := range clusterTemplates {
				Expect(clusterTemplate).To(matchers.HaveYAMLPathWithValue("$.metadata.annotations['tkg/plan']", "prod"))
			}
		})
	})

	When("CNI is set", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
		})

		It("succeeds rendering kubeadmConfigTemplate with CNI:antrea", func() {
			kubeadmCommands := []string{
				"powershell C:/Temp/antrea.ps1 -ExecutionPolicy Bypass",
				"powershell c:/k/prevent_windows_updates.ps1 -ExecutionPolicy Bypass",
				"powershell C:/k/register_antrea_cleanup.ps1 -ExecutionPolicy Bypass",
			}
			kubeadmFiles := []string{
				"C:\\Temp\\antrea.ps1",
				"c:\\k\\prevent_windows_updates.ps1",
				"C:\\k\\antrea_cleanup.ps1",
				"C:\\k\\register_antrea_cleanup.ps1",
			}
			value.Set("CNI", "antrea")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			const kubeadmConfigName = "win-wl-windows-containerd"
			kubeadmConfigTemplates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "KubeadmConfigTemplate",
				"$.metadata.name": kubeadmConfigName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(kubeadmConfigTemplates)).NotTo(Equal(0))

			for _, kubeadmConfigTemplate := range kubeadmConfigTemplates {
				// Match commands rendered in the template
				kubeadmTemplatePath := "$.spec.template.spec.postKubeadmCommands"
				for i, command := range kubeadmCommands {
					Expect(kubeadmConfigTemplate).To(matchers.HaveYAMLPathWithValue(
						fmt.Sprintf("%s[%d]", kubeadmTemplatePath, i), command,
					))
				}
				// Match files rendered in the template
				filesTemplatePath := "$.spec.template.spec.files"
				for i, file := range kubeadmFiles {
					Expect(kubeadmConfigTemplate).To(matchers.HaveYAMLPathWithValue(
						fmt.Sprintf("%s[%d]['path']", filesTemplatePath, i), file,
					))
				}
			}
		})

		It("succeeds with cluster objects with CNI:antrea", func() {
			value.Set("CNI", "antrea")

			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
			Expect(err).NotTo(HaveOccurred())

			tests := []struct {
				name string
				kind string
			}{
				{
					name: "tkg-antrea-rc-init",
					kind: "ConfigMap",
				},
				{
					name: "win-wl-antrea-addon",
					kind: "Secret",
				},
				{
					name: "win-wl-tkg-antrea-cls-init",
					kind: "ClusterResourceSet",
				},
				{
					name: "win-wl-windows-containerd",
					kind: "KubeadmConfigTemplate",
				},
			}

			// Check the existence of Antrea installation objects
			for _, tt := range tests {
				templates, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
					"$.kind":          tt.kind,
					"$.metadata.name": tt.name,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(len(templates)).ToNot(Equal(0))
			}
		})
	})
})
