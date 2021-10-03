// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/ytt"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Windows Ytt Templating", func() {
	var paths []string
	YAML_ROOT := "../../"
	BeforeEach(func() {
		paths = []string{
			//  Map item (key 'infraProvider') on line stdin.yml:8:
			filepath.Join(YAML_ROOT, "config_default.yaml"),
			filepath.Join("./tkr-bom-v1.21.1.yaml"),
			filepath.Join("./tkg-bom-v1.4.0.yaml"),
			filepath.Join(YAML_ROOT, "infrastructure-vsphere", "v0.7.10", "ytt", "base-template.yaml"),
			filepath.Join(YAML_ROOT, "infrastructure-vsphere", "v0.7.10", "ytt", "overlay-windows.yaml"),
			filepath.Join(YAML_ROOT, "ytt", "02_addons", "cni", "antrea", "antrea_addon_data.lib.yaml"),
			filepath.Join(YAML_ROOT, "ytt", "02_addons", "cpi", "cpi_addon_data.lib.yaml"),
			filepath.Join(YAML_ROOT, "provider-bundle", "providers", "ytt", "02_addons", "cpi", "cpi_addon_data.lib.yaml"),
			filepath.Join(YAML_ROOT, "ytt"), // lib/helpers.star, lib/config_variable_association.star, lib/validate.star
		}
	})
	It("Has a windows overlay", func() {
		values := createDataValues(map[string]string{
			"VSPHERE_INSECURE":            "true",
			"PROVIDER_TYPE":               "vsphere",
			"IS_WINDOWS_WORKLOAD_CLUSTER": "true",
			"TKG_CLUSTER_ROLE":            "management",
			"VSPHERE_USERNAME":            "a",
			"VSPHERE_PASSWORD":            "a",
			"VSPHERE_SERVER":              "a",
			"VSPHERE_DATACENTER":          "a",
			"VSPHERE_RESOURCE_POOL":       "a",
			"VSPHERE_FOLDER":              "a",
			"VSPHERE_SSH_AUTHORIZED_KEY":  "a",
			"TKG_DEFAULT_BOM":             "tkg-bom-v1.4.0.yaml",
			"KUBERNETES_RELEASE":          "v1.21.2---vmware.1-tkg.2-20210924-539f8b15",
		})

		fmt.Println(values)
		x, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
		fmt.Println(x)
		fmt.Println(values)
		Expect(err).NotTo(HaveOccurred())

		for _, capiString := range strings.Split(x, "---") {
			capiObject := make(map[string]interface{})
			yaml.Unmarshal([]byte(capiString), capiObject)
			for k, v := range capiObject {
				fmt.Println(fmt.Sprintf("%v %v", k, v))
			}
		}
		// add assertions for data collected above...
	})
})

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
