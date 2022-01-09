// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"io"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/matchers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests/unit/ytt"
)

type yttValues map[string]string

func (v yttValues) toReader() io.Reader {
	return strings.NewReader(createDataValues(v))
}

func (v yttValues) Set(key, value string) {
	v[key] = value
}

func (v yttValues) Delete(key string) {
	delete(v, key)
}

func (v yttValues) DeepCopy() yttValues {
	other := make(yttValues)
	for key, value := range v {
		other[key] = value
	}
	return other
}

func assertNotFound(docs []string, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(docs).To(HaveLen(0))
}

func assertFoundOne(docs []string, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(docs).To(HaveLen(1))
}

const (
	AviUsername            = "admin"
	AviPassword            = "pass_1234"
	AviCaName              = "avi-controller-ca"
	AviCaDataB64           = "LS0tLS1CR...LS0tLQo="
	AviAdminCredentialName = "avi-controller-credentials"
)

func NoAVIRelatedObjects(output string) {
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Secret",
		"$.metadata.name": "avi-secret",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ServiceAccount",
		"$.metadata.name": "ako-sa",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ConfigMap",
		"$.metadata.name": "ako-sa",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRole",
		"$.metadata.name": "ako-cr",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRoleBinding",
		"$.metadata.name": "ako-crb",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "StatefulSet",
		"$.metadata.name": "ako",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Service",
		"$.metadata.name": "ako-operator-controller-manager-metrics-service",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRole",
		"$.metadata.name": "ako-operator-manager-bootstrap",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRoleBinding",
		"$.metadata.name": "ako-operator-bootstrap-rolebinding",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "StatefulSet",
		"$.metadata.name": "ako-operator-controller-manager",
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Secret",
		"$.metadata.name": AviAdminCredentialName,
	}))
	assertNotFound(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Secret",
		"$.metadata.name": AviCaName,
	}))
}

func AllAVIRelatedObjects(output string) {
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Secret",
		"$.metadata.name": "avi-secret",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ServiceAccount",
		"$.metadata.name": "ako-sa",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ConfigMap",
		"$.metadata.name": "avi-k8s-config",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRole",
		"$.metadata.name": "ako-cr",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRoleBinding",
		"$.metadata.name": "ako-crb",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "StatefulSet",
		"$.metadata.name": "ako",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Service",
		"$.metadata.name": "ako-operator-controller-manager-metrics-service",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRole",
		"$.metadata.name": "ako-operator-manager-bootstrap",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "ClusterRoleBinding",
		"$.metadata.name": "ako-operator-bootstrap-rolebinding",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Deployment",
		"$.metadata.name": "ako-operator-controller-manager",
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Secret",
		"$.metadata.name": AviAdminCredentialName,
	}))
	assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
		"$.kind":          "Secret",
		"$.metadata.name": AviCaName,
	}))
}

var _ = Describe("AKO-operator Ytt Templating", func() {

	var paths []string
	var baseVal yttValues

	BeforeEach(func() {
		paths = []string{
			filepath.Join(yamlRoot, "config_default.yaml"),
			filepath.Join("./fixtures/tkr-bom-v1.21.1.yaml"),
			filepath.Join("./fixtures/tkg-bom-v1.4.0.yaml"),
			filepath.Join(yamlRoot, "ytt"),
		}

		baseVal = map[string]string{
			// required fields
			"TKG_DEFAULT_BOM":    "tkg-bom-v1.4.0.yaml",
			"KUBERNETES_RELEASE": "v1.21.2---vmware.1-tkg.1",

			// required fields to enable AVI
			"PROVIDER_TYPE":                 "vsphere",
			"TKG_CLUSTER_ROLE":              "management",
			"AVI_ENABLE":                    "true",
			"AVI_CONTROL_PLANE_HA_PROVIDER": "true",

			// required vsphere configurations
			"VSPHERE_USERNAME":           "user_blah",
			"VSPHERE_PASSWORD":           "pass_1234",
			"VSPHERE_SERVER":             "vmware-tanzu.com",
			"VSPHERE_DATACENTER":         "vmware-tanzu-dc.com",
			"VSPHERE_RESOURCE_POOL":      "myrp",
			"VSPHERE_FOLDER":             "ds0",
			"VSPHERE_SSH_AUTHORIZED_KEY": "ssh-rsa AAAA...+M7Q== vmware-tanzu.local",
			"VSPHERE_INSECURE":           "true",

			// required avi related values
			"AVI_USERNAME":              AviUsername,
			"AVI_PASSWORD":              AviPassword,
			"AVI_CA_NAME":               AviCaName,
			"AVI_CA_DATA_B64":           AviCaDataB64,
			"AVI_ADMIN_CREDENTIAL_NAME": AviAdminCredentialName,
		}
	})

	When("basic values are provided", func() {
		It("renders without error", func() {
			_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, baseVal.toReader())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("use public cloud aws", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
			value.Set("PROVIDER_TYPE", "aws")
		})

		When("avi enabled", func() {
			It("does not render avi objects as multi-cloud is not yet implemented", func() {
				value.Set("AVI_ENABLE", "true")
				output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
				Expect(err).NotTo(HaveOccurred())
				NoAVIRelatedObjects(output)
			})
		})

		When("avi disabled", func() {
			It("does not render avi objects as avi is disabled", func() {
				value.Set("AVI_ENABLE", "false")
				output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
				Expect(err).NotTo(HaveOccurred())
				NoAVIRelatedObjects(output)
			})
		})
	})

	When("use vsphere", func() {
		var value yttValues
		BeforeEach(func() {
			value = baseVal.DeepCopy()
			value.Set("PROVIDER_TYPE", "vsphere")
		})

		When("enable avi", func() {
			BeforeEach(func() {
				value.Set("AVI_ENABLE", "true")
			})

			When("avi as HA provider", func() {
				BeforeEach(func() {
					value.Set("AVI_CONTROL_PLANE_HA_PROVIDER", "true")
				})

				It("render all avi related objects", func() {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
					Expect(err).NotTo(HaveOccurred())
					AllAVIRelatedObjects(output)
				})

				When("AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP is not empty", func() {
					const (
						mgmtSEG = "mgmt-seg"
						aviSEG  = "seg"
					)
					It("prefer AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP as SEG name", func() {
						value.Set("AVI_SERVICE_ENGINE_GROUP", aviSEG)
						value.Set("AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP", mgmtSEG)
						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
						Expect(err).NotTo(HaveOccurred())
						assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
							"$.kind":                        "ConfigMap",
							"$.metadata.name":               "avi-k8s-config",
							"$.data.serviceEngineGroupName": mgmtSEG,
						}))
					})
				})

				When("AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP is empty", func() {
					const (
						mgmtSEG = ""
						aviSEG  = "seg"
					)
					It("fallback to AVI_SERVICE_ENGINE_GROUP as SEG name", func() {
						value.Set("AVI_SERVICE_ENGINE_GROUP", aviSEG)
						value.Set("AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP", mgmtSEG)
						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
						Expect(err).NotTo(HaveOccurred())
						assertFoundOne(matchers.FindDocsMatchingYAMLPath(output, map[string]string{
							"$.kind":                        "ConfigMap",
							"$.metadata.name":               "avi-k8s-config",
							"$.data.serviceEngineGroupName": aviSEG,
						}))
					})
				})
			})

			When("avi is not HA provider", func() {
				It("does not render avi objects as avi is not control plane HA provider", func() {
					value.Set("AVI_CONTROL_PLANE_HA_PROVIDER", "false")
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
					Expect(err).NotTo(HaveOccurred())
					NoAVIRelatedObjects(output)
				})
			})
		})

		When("disable avi", func() {
			It("does not render avi objects as it is disabled", func() {
				value.Set("AVI_ENABLE", "false")
				output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
				Expect(err).NotTo(HaveOccurred())
				NoAVIRelatedObjects(output)
			})
		})
	})
})
