// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

var _ = Describe("POD_SECURITY_STANDARD Ytt Templating", func() {
	Describe("Pod Security Standard ytt validations", func() {
		var paths []string
		var baseVal yttValues
		BeforeEach(func() {
			paths = []string{
				//  Map item (key 'infraProvider') on line stdin.yml:8:
				filepath.Join(yamlRoot, "config_default.yaml"),
				"./fixtures/tkr-bom-v1.21.1.yaml",
				"./fixtures/tkg-bom-v1.4.0.yaml",
				filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "base-template.yaml"),
				filepath.Join(yamlRoot, "infrastructure-vsphere", capvVersion, "ytt", "overlay-windows.yaml"),
				filepath.Join(yamlRoot, "ytt"), // lib/helpers.star, lib/config_variable_association.star, lib/validate.star

				filepath.Join(yamlRoot, "ytt", "03_customizations", "pod_security_standard"),
			}

			baseVal = map[string]interface{}{
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
				"CORE_DNS_IP":                 "10.64.0.10",
				"CLUSTER_CIDR":                "192.168.1.0/16",
			}
		})

		When("Configuring POD_SECURITY_STANDARD_DEACTIVATED", func() {
			It("allows undefined", func() {
				baseVal["POD_SECURITY_STANDARD_DEACTIVATED"] = ""
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not allow garbage", func() {
				baseVal["POD_SECURITY_STANDARD_DEACTIVATED"] = "garbage"
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})

			It("does allow true", func() {
				baseVal["POD_SECURITY_STANDARD_DEACTIVATED"] = "true"
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})

			It("does allow false", func() {
				baseVal["POD_SECURITY_STANDARD_DEACTIVATED"] = "false"
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("Configuring POD_SECURITY_STANDARD_* to different values", func() {
			It("does allow baseline", func() {
				baseline := "baseline"
				baseVal["POD_SECURITY_STANDARD_WARN"] = baseline
				baseVal["POD_SECURITY_STANDARD_AUDIT"] = baseline
				baseVal["POD_SECURITY_STANDARD_ENFORCE"] = baseline
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
			It("does allow restricted", func() {
				restricted := "restricted"
				baseVal["POD_SECURITY_STANDARD_WARN"] = restricted
				baseVal["POD_SECURITY_STANDARD_AUDIT"] = restricted
				baseVal["POD_SECURITY_STANDARD_ENFORCE"] = restricted
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
			It("does allow privileged", func() {
				privileged := "privileged"
				baseVal["POD_SECURITY_STANDARD_WARN"] = privileged
				baseVal["POD_SECURITY_STANDARD_AUDIT"] = privileged
				baseVal["POD_SECURITY_STANDARD_ENFORCE"] = privileged
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).NotTo(HaveOccurred())
			})
			It("does not allow to misconfigure POD_SECURITY_STANDARD_WARN", func() {
				baseVal["POD_SECURITY_STANDARD_WARN"] = "this"
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
			It("does not allow to misconfigure POD_SECURITY_STANDARD_AUDIT", func() {
				baseVal["POD_SECURITY_STANDARD_AUDIT"] = "is"
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
			It("does not allow to misconfigure POD_SECURITY_STANDARD_ENFORCE", func() {
				baseVal["POD_SECURITY_STANDARD_ENFORCE"] = "misconfigured"
				values := createDataValues(baseVal)
				_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
				Expect(err).To(HaveOccurred())
			})
		})

	})
})
