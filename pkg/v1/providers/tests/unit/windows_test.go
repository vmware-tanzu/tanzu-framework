// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "sigs.k8s.io/cluster-api/util/yaml"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

var _ = Describe("Windows YTT Template", func() {
	var (
		values map[string]interface{}
		paths  []string
	)
	BeforeEach(func() {
		// file paths used on template rendering
		paths = []string{
			//  Map item (key 'infraProvider') on line stdin.yml:8:
			filepath.Join(yamlRoot, "config_default.yaml"),
			filepath.Join("./fixtures/tkr-bom-v1.21.1.yaml"),
			filepath.Join("./fixtures/tkg-bom-v1.4.0.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "overlay-windows.yaml"),
			filepath.Join(yamlRoot, "ytt", "02_addons", "cni", "antrea", "antrea_addon_data.lib.yaml"),
			filepath.Join(yamlRoot, "ytt", "02_addons", "cpi", "cpi_addon_data.lib.yaml"),
			filepath.Join(yamlRoot, "ytt", "03_customizations", "02_avi", "ako-deployment.lib.yaml"),
			filepath.Join(yamlRoot, "ytt"), // lib/helpers.star, lib/config_variable_association.star, lib/validate.star
		}
		// input values settings for template rendering
		values = map[string]interface{}{
			"VSPHERE_INSECURE":            "true",
			"PROVIDER_TYPE":               "vsphere",
			"VSPHERE_USERNAME":            "user_blah",
			"VSPHERE_PASSWORD":            "pass_1234",
			"VSPHERE_SERVER":              "vmware-tanzu.com",
			"VSPHERE_DATACENTER":          "vmware-tanzu-dc.com",
			"VSPHERE_RESOURCE_POOL":       "myrp",
			"VSPHERE_FOLDER":              "ds0",
			"VSPHERE_SSH_AUTHORIZED_KEY":  "ssh-rsa AAAA...+M7Q== vmware-tanzu.local",
			"TKG_DEFAULT_BOM":             "tkg-bom-v1.4.0.yaml",
			"KUBERNETES_RELEASE":          "v1.21.2---vmware.1-tkg.1",
			"CORE_DNS_IP":                 "10.64.0.10",
			"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
			"TKG_CLUSTER_ROLE":            "workload",
			"CLUSTER_NAME":                "win-wl",
		}
	})
	It("Support multiple Windows workload nodes", func() {
		workloadCount := 10
		values["WORKER_MACHINE_COUNT"] = workloadCount

		rawClusterAPIYaml, err := renderYTTTemplate(paths, values)
		Expect(err).NotTo(HaveOccurred())

		unstructObjs, err := searchObjectByKind("MachineDeployment", rawClusterAPIYaml)
		Expect(err).NotTo(HaveOccurred())

		for _, unstruct := range unstructObjs {
			replicas, _, err := unstructured.NestedFloat64(unstruct.Object, "spec", "replicas")
			Expect(err).NotTo(HaveOccurred())
			Expect(replicas).To(Equal(float64(workloadCount)))
		}
	})
	It("Has cluster_plan=prod cluster annotation", func() {
		values["CLUSTER_PLAN"] = "prod"

		rawClusterAPIYaml, err := renderYTTTemplate(paths, values)
		Expect(err).NotTo(HaveOccurred())
		obj, err := searchObjectByName("win-wl", rawClusterAPIYaml) // cluster object
		Expect(err).NotTo(HaveOccurred())

		compare := reflect.DeepEqual(obj.GetAnnotations(), map[string]string{"osInfo": ",,", "tkg/plan": "prod"})
		Expect(compare).To(BeTrue())
	})
	It("Has correct kubeadmConfigTemplate for CNI:None", func() {
		values["CNI"] = "none"
		rawClusterAPIYaml, err := renderYTTTemplate(paths, values)
		Expect(err).NotTo(HaveOccurred())
		Expect(rawClusterAPIYaml).To(ContainSubstring("CNI: none"))

		unstruct, err := searchObjectByKind("KubeadmConfigTemplate", rawClusterAPIYaml)
		Expect(err).NotTo(HaveOccurred())
		obj := unstruct[0].Object

		// Checking files registered.
		files, _, err := unstructured.NestedSlice(obj, "spec", "template", "spec", "files")
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(flattenMapSlice(files, "path")).Should(ContainElement("c:\\k\\prevent_windows_updates.ps1"))

		// Checking installed commands on postKubeadmCommands
		cmds, _, err := unstructured.NestedSlice(obj, "spec", "template", "spec", "postKubeadmCommands")
		Expect(err).NotTo(HaveOccurred())
		Expect(cmds).To(HaveLen(1))
		Expect(cmds).Should(ContainElement("powershell c:/k/prevent_windows_updates.ps1 -ExecutionPolicy Bypass"))
	})
	It("Has correct kubeadmConfigTemplate on CNI:Antrea", func() {
		rawClusterAPIYaml, err := renderYTTTemplate(paths, values)
		Expect(err).NotTo(HaveOccurred())
		Expect(rawClusterAPIYaml).To(ContainSubstring("CNI: antrea"))

		// Test kubeadmConfigTemplate rendering.
		unstruct, err := searchObjectByKind("KubeadmConfigTemplate", rawClusterAPIYaml)
		Expect(err).NotTo(HaveOccurred())
		obj := unstruct[0].Object

		// Checking files registered.
		files, _, err := unstructured.NestedSlice(obj, "spec", "template", "spec", "files")
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(4))
		Expect(flattenMapSlice(files, "path")).Should(ContainElements([]string{
			"C:\\Temp\\antrea.ps1",
			"c:\\k\\prevent_windows_updates.ps1",
			"C:\\k\\antrea_cleanup.ps1",
			"C:\\k\\register_antrea_cleanup.ps1",
		}))

		// Checking installed commands on postKubeadmCommands
		cmds, _, err := unstructured.NestedSlice(obj, "spec", "template", "spec", "postKubeadmCommands")
		Expect(err).NotTo(HaveOccurred())
		Expect(cmds).To(HaveLen(3))
		Expect(cmds).Should(ContainElements([]string{
			"powershell C:/k/register_antrea_cleanup.ps1 -ExecutionPolicy Bypass",
			"powershell C:/Temp/antrea.ps1 -ExecutionPolicy Bypass",
			"powershell c:/k/prevent_windows_updates.ps1 -ExecutionPolicy Bypass",
		}))
	})
	It("Has cluster objects for CNI:antrea", func() {
		var (
			objs []unstructured.Unstructured
			obj  *unstructured.Unstructured
		)

		rawClusterAPIYaml, err := renderYTTTemplate(paths, values)
		Expect(err).NotTo(HaveOccurred())

		// Test default configuration values.
		var config string
		if obj, err = searchObjectByName("win-wl-config-values", rawClusterAPIYaml); err == nil {
			stringData, _, err := unstructured.NestedMap(obj.Object, "stringData")
			Expect(err).NotTo(HaveOccurred())
			config = stringData["value"].(string)
		}
		Expect(config).To(ContainSubstring("CNI: antrea"))
		Expect(err).NotTo(HaveOccurred())

		// check Antrea version listed on components
		var version string
		if objs, err = utilyaml.ToUnstructured([]byte(rawClusterAPIYaml)); err == nil {
			antreaComponent, _, err := unstructured.NestedSlice(objs[0].Object, "components", "antrea")
			Expect(err).NotTo(HaveOccurred())

			for _, c := range antreaComponent {
				version, _, err = unstructured.NestedString(c.(map[string]interface{}), "version")
				Expect(err).NotTo(HaveOccurred())
			}
		}
		Expect(version).To(Equal("v0.13.3+vmware.1"))
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
				name: "win-wl-md-0-windows-containerd",
				kind: "KubeadmConfigTemplate",
			},
		}
		// Check existence of default Antrea installation objects
		for _, tt := range tests {
			tmp, err := searchObjectByName(tt.name, rawClusterAPIYaml)
			Expect(err).NotTo(HaveOccurred())
			Expect(tmp).NotTo(BeNil())
			Expect(tmp.GetKind()).To(Equal(tt.kind))
		}
	})
	It("Has a windows workload overlay", func() {
		// useful debugging information that we don't actually need for day-to-day testing
		rawClusterAPIYaml, _ := renderYTTTemplate(paths, values)

		// Test 1: Making sure that we have a few basic ClusterAPI objects in the windows templates...
		// clusterAPIComponents is a list of all the 'kind' objects that we want to see.
		// for windows, the most important thing to confirm is that we have 2 VsphereMachineTemplates,
		// since there are obviously going to be linux as well as windows machine types.
		clusterAPIComponents := map[string]int{
			"Cluster":                1,
			"VSphereCluster":         1,
			"VSphereMachineTemplate": 2,
		}

		seen, err := countCapiCompKinds(rawClusterAPIYaml)
		Expect(err).NotTo(HaveOccurred())
		for k, v := range clusterAPIComponents {
			val, ok := seen[k]
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal(v), fmt.Sprintf("Of type %s", k))
		}
	})
})

func searchObjectByKind(kind, rawYAML string) ([]unstructured.Unstructured, error) {
	var unstructObjects []unstructured.Unstructured
	objs, err := utilyaml.ToUnstructured([]byte(rawYAML))
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if kind == obj.GetKind() {
			unstructObjects = append(unstructObjects, obj)
		}
	}
	return unstructObjects, nil
}

func searchObjectByName(name, rawYAML string) (*unstructured.Unstructured, error) {
	objs, err := utilyaml.ToUnstructured([]byte(rawYAML))
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if name == obj.GetName() {
			return &obj, nil
		}
	}
	return nil, fmt.Errorf("object %s not found", name)
}

func flattenMapSlice(mapper []interface{}, key string) []string {
	result := make([]string, len(mapper))
	for k, v := range mapper {
		if x, ok := v.(map[string]interface{})[key]; ok {
			result[k] = x.(string)
		}
	}
	return result
}

// countCapiCompKinds counts up the number of different api types in the final YAML output
func countCapiCompKinds(rawClusterAPIYaml string) (map[string]int, error) {
	kinds := make(map[string]int)
	objs, err := utilyaml.ToUnstructured([]byte(rawClusterAPIYaml))

	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		kinds[obj.GetKind()] = kinds[obj.GetKind()] + 1
	}

	return kinds, nil
}

func renderYTTTemplate(paths []string, values map[string]interface{}) (string, error) {
	dataValues := createDataValues(values)
	return ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(dataValues))
}

/**
Adopted from this hacky string...
/usr/local/bin/ytt
--ignore-unknown-comments
--data-value=infraProvider=vsphere
--data-value=TKG_CLUSTER_ROLE=management
--data-value=IS_WINDOWS_WORKLOAD_CLUSTER=true
--data-value=VSPHERE_USERNAME=a
--data-value=VSPHERE_PASSWORD=a
--data-value=VSPHERE_SERVER=a
--data-value=VSPHERE_DATACENTER=a
--data-value=VSPHERE_RESOURCE_POOL=a
--data-value=VSPHERE_FOLDER=a,
--data-value=VSPHERE_SSH_AUTHORIZED_KEY="a"
-f tkr-bom-v1.21.1.yaml
-f tkg-bom-v1.4.0.yaml
-f config.yaml
--data-value=TKG_DEFAULT_BOM=tkg-bom-v1.4.0.yaml
--data-value=KUBERNETES_RELEASE=v1.21.2---vmware.1-tkg.2-20210924-539f8b15
-f ../../config_default.yaml
-f ../../infrastructure-vsphere/v0.7.10/ytt/base-template.yaml
-f ../../infrastructure-vsphere/v0.7.10/ytt/overlay-windows.yaml
-f ../..//ytt/02_addons/cni/antrea/antrea_addon_data.lib.yaml
-f ../../ytt/02_addons/cpi/cpi_addon_data.lib.yaml
-f ../../provider-bundle/providers/ytt/02_addons/cpi/cpi_addon_data.lib.yaml
-f ./ytt_libs_4_test/
*/
