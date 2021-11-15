// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	utilyaml "sigs.k8s.io/cluster-api/util/yaml"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/ytt"
)

var _ = Describe("Windows Ytt Templating", func() {
	var paths []string
	BeforeEach(func() {
		paths = []string{
			//  Map item (key 'infraProvider') on line stdin.yml:8:
			filepath.Join(yamlRoot, "config_default.yaml"),
			filepath.Join("./fixtures/tkr-bom-v1.21.1.yaml"),
			filepath.Join("./fixtures/tkg-bom-v1.4.0.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", "v1.0.1", "ytt", "base-template.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", "v1.0.1", "ytt", "overlay-windows.yaml"),
			filepath.Join(yamlRoot, "infrastructure-vsphere", "ytt", "vsphere-overlay.yaml"),
			filepath.Join(yamlRoot, "ytt", "02_addons", "cni", "antrea", "antrea_addon_data.lib.yaml"),
			filepath.Join(yamlRoot, "ytt", "02_addons", "cpi", "cpi_addon_data.lib.yaml"),
			filepath.Join(yamlRoot, "ytt", "03_customizations", "02_avi", "ako-deployment.lib.yaml"),
			//filepath.Join(YAML_ROOT, "provider-bundle", "providers", "ytt", "02_addons", "cpi", "cpi_addon_data.lib.yaml"),
			filepath.Join(yamlRoot, "ytt", "03_customizations", "03_windows"),
			filepath.Join(yamlRoot, "ytt"), // lib/helpers.star, lib/config_variable_association.star, lib/validate.star
		}
	})
	It("Has a windows overlay", func() {
		values := createDataValues(map[string]string{
			"VSPHERE_INSECURE":            "true",
			"PROVIDER_TYPE":               "vsphere",
			"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
			"TKG_CLUSTER_ROLE":            "management",
			"VSPHERE_USERNAME":            "user_blah",
			"VSPHERE_PASSWORD":            "pass_1234",
			"VSPHERE_SERVER":              "vmware-tanzu.com",
			"VSPHERE_DATACENTER":          "vmware-tanzu-dc.com",
			"VSPHERE_RESOURCE_POOL":       "myrp",
			"VSPHERE_FOLDER":              "ds0",
			"VSPHERE_SSH_AUTHORIZED_KEY":  "ssh-rsa AAAA...+M7Q== vmware-tanzu.local",
			"TKG_DEFAULT_BOM":             "tkg-bom-v1.4.0.yaml",
			"KUBERNETES_RELEASE":          "v1.21.2---vmware.1-tkg.1",
		})

		// Printing these values helps for debugging the ytt command if needed
		fmt.Println(values)

		// useful debugging information that we don't actually need for day-to-day testing
		rawClusterAPIYaml, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
		Expect(err).NotTo(HaveOccurred())

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

		// TODO add more validations for things like the antrea installation contents etc...
	})
})

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
