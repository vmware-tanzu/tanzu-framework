// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"io"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

type yttValues map[string]interface{}

func (v yttValues) toReader() io.Reader {
	return strings.NewReader(createDataValues(v))
}

func (v yttValues) Set(key string, value interface{}) {
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
			"CLUSTER_CIDR":               "192.168.1.0/16",

			// required by CAPV
			"TKG_IP_FAMILY": "ipv4",
			"SERVICE_CIDR":  "5.5.5.5/16",

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

			Context("network separation feature", func() {
				getAVIK8sConfig := func(v yttValues) string {
					output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, v.toReader())
					Expect(err).NotTo(HaveOccurred())
					cm, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
						"$.kind":          "ConfigMap",
						"$.metadata.name": "avi-k8s-config",
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(cm).To(HaveLen(1))
					return cm[0]
				}

				BeforeEach(func() {
					value.Set("AVI_DATA_NETWORK", "data-network")
					value.Set("AVI_DATA_NETWORK_CIDR", "10.0.1.0/24")
				})

				It("render data network as vip network", func() {
					Expect(getAVIK8sConfig(value)).To(ContainSubstring("[{\"cidr\":\"10.0.1.0/24\",\"networkName\":\"data-network\"}]"))
				})

				When("mc vip is provided", func() {
					BeforeEach(func() {
						value.Set("AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_NAME", "mc-network")
						value.Set("AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_CIDR", "10.0.2.0/24")
					})

					It("render mc network as vip network", func() {
						Expect(getAVIK8sConfig(value)).To(ContainSubstring("[{\"cidr\":\"10.0.2.0/24\",\"networkName\":\"mc-network\"}]"))
					})

					When("mc control plane vip is provided", func() {
						BeforeEach(func() {
							value.Set("AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_NAME", "mc-control-plane-network")
							value.Set("AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_CIDR", "10.0.3.0/24")
						})

						It("render mc control plane as vip network", func() {
							Expect(getAVIK8sConfig(value)).To(ContainSubstring("[{\"cidr\":\"10.0.3.0/24\",\"networkName\":\"mc-control-plane-network\"}]"))
						})
					})
				})
			})

			Context("setting AVI_LABELS feature", func() {
				BeforeEach(func() {
					value.Set("AVI_LABELS", map[string]string{
						"foo1": "bar1",
						"foo2": "bar2",
					})
				})
				When("management cluster", func() {
					BeforeEach(func() {
						value.Set("TKG_CLUSTER_ROLE", "management")
					})

					It("adds labelSelector for install-ako-for-all", func() {
						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
						Expect(err).NotTo(HaveOccurred())
						docs, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
							"$.kind":          "Secret",
							"$.metadata.name": "test-cluster-ako-operator-addon",
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(docs).To(HaveLen(1))
						Expect(docs[0]).To(ContainSubstring("avi_labels: '{\"foo1\":\"bar1\",\"foo2\":\"bar2\"}'"))
					})
				})

				When("workload cluster", func() {
					BeforeEach(func() {
						value.Set("TKG_CLUSTER_ROLE", "workload")
					})

					It("labels the workload cluster", func() {
						output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
						Expect(err).NotTo(HaveOccurred())
						docs, err := matchers.FindDocsMatchingYAMLPath(output, map[string]string{
							"$.kind":          "Cluster",
							"$.metadata.name": "test-cluster",
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(docs).To(HaveLen(1))
						Expect(docs[0]).To(ContainSubstring("    foo1: bar1\n    foo2: bar2"))
					})
				})

			})

			When("workload cluster and AVI_LABELS not provided", func() {
				BeforeEach(func() {
					value.Set("TKG_CLUSTER_ROLE", "workload")
				})

				It("should render without error", func() {
					_, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, value.toReader())
					Expect(err).NotTo(HaveOccurred())
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
