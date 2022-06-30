// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package compatibility

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/sets"
)

func TestCompatibility(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "TKR Compatibility: unit tests", suiteConfig)
}

var _ = Describe("Compatibility", func() {
	var (
		objects []client.Object
		c       *Compatibility
	)

	BeforeEach(func() {
		objects = nil
		c = &Compatibility{
			Log: logr.Discard(),
			Config: Config{
				TKRNamespace: "tkg-system",
			},
		}
	})

	JustBeforeEach(func() {
		scheme := initScheme()
		c.Client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
	})

	Describe("getMCCompatibleTKRVersions()", func() {
		When("Management Cluster object doesn't exist", func() {
			It("should return an empty set", func() {
				compatibleVersions, err := c.getMCCompatibleTKRVersions(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(compatibleVersions).To(Equal(sets.Strings()))
			})
		})

		When("Management Cluster object exists", func() {
			var (
				mc *clusterv1.Cluster
			)

			BeforeEach(func() {
				mc = &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mc",
						Namespace: "not-important",
						Labels: map[string]string{
							constants.ManagementClusterRoleLabel: "",
						},
						Annotations: map[string]string{
							constants.TKGVersionKey: currentTKGVersion,
						},
					},
				}
				objects = append(objects, mc)
			})

			When("bom-metadata ConfigMap doesn't exist", func() {
				It("should return an empty set", func() {
					compatibleVersions, err := c.getMCCompatibleTKRVersions(context.Background())
					Expect(err).ToNot(HaveOccurred())
					Expect(compatibleVersions).To(Equal(sets.Strings()))
				})
			})

			When("bom-metadata ConfigMap exists", func() {
				var (
					cm *corev1.ConfigMap
				)

				BeforeEach(func() {
					cm = &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: c.Config.TKRNamespace,
							Name:      constants.BOMMetadataConfigMapName,
						},
					}
					objects = append(objects, cm)
				})

				When("binaryData.compatibility key does not exist", func() {
					It("should return an empty set", func() {
						compatibleVersions, err := c.getMCCompatibleTKRVersions(context.Background())
						Expect(err).ToNot(HaveOccurred())
						Expect(compatibleVersions).To(Equal(sets.Strings()))
					})
				})

				When("binaryData.compatibility key exists", func() {
					BeforeEach(func() {
						cm.BinaryData = map[string][]byte{}
					})

					When("binaryData.compatibility cannot be parsed", func() {
						BeforeEach(func() {
							cm.BinaryData[constants.BOMMetadataCompatibilityKey] = []byte("garbage")
						})

						It("should return an empty set", func() {
							compatibleVersions, err := c.getMCCompatibleTKRVersions(context.Background())
							Expect(err).ToNot(HaveOccurred())
							Expect(compatibleVersions).To(Equal(sets.Strings()))
						})
					})

					When("binaryData.compatibility can be parsed", func() {
						BeforeEach(func() {
							cm.BinaryData[constants.BOMMetadataCompatibilityKey] = []byte(bomMetadataStr)
						})

						It("should return an actual set of compatible versions", func() {
							compatibleVersions, err := c.getMCCompatibleTKRVersions(context.Background())
							Expect(err).ToNot(HaveOccurred())
							Expect(compatibleVersions).To(Equal(sets.Strings(v1235vmware1tkg1zshippable, v1228vmware1tkg2zshippable, v12111vmware1tkg2zshippable)))
						})

						When("the current cluster version is not listed in bom-metadata", func() {
							BeforeEach(func() {
								mc.Annotations[constants.TKGVersionKey] = "v1.6.0-not-listed"
							})

							It("should return an empty set", func() {
								compatibleVersions, err := c.getMCCompatibleTKRVersions(context.Background())
								Expect(err).ToNot(HaveOccurred())
								Expect(compatibleVersions).To(Equal(sets.Strings()))
							})
						})
					})
				})
			})
		})
	})

	Describe("getAdditionalCompatibleTKRVersions()", func() {
		Context("ConfigMaps listing additional compatible TKRs", func() {
			When("do not exist", func() {
				It("should return an empty set", func() {
					compatibleVersions, err := c.getAdditionalCompatibleTKRVersions(context.Background())
					Expect(err).ToNot(HaveOccurred())
					Expect(compatibleVersions).To(Equal(sets.Strings()))
				})
			})

			When("exist outside the configured TKR namespace", func() {
				var cm *corev1.ConfigMap

				BeforeEach(func() {
					cm = &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "user1",
							Name:      "my-tkrs",
							Labels: map[string]string{
								LabelAdditionalTKRs: "",
							},
						},
						Data: map[string]string{
							fieldTKRVersions: fmt.Sprintf("[%s, %s]", v1240vmware1tkg1, v1250vmware1tkg1),
						},
					}
					objects = append(objects, cm)
				})

				It("should return an empty set", func() {
					compatibleVersions, err := c.getAdditionalCompatibleTKRVersions(context.Background())
					Expect(err).ToNot(HaveOccurred())
					Expect(compatibleVersions).To(Equal(sets.Strings()))
				})
			})

			When("exist inside the configured TKR namespace", func() {
				var cm *corev1.ConfigMap

				BeforeEach(func() {
					cm = &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: c.Config.TKRNamespace,
							Name:      "admin-byo-tkrs",
							Labels: map[string]string{
								LabelAdditionalTKRs: "",
							},
						},
						Data: map[string]string{
							fieldTKRVersions: fmt.Sprintf("[%s, %s]", v1240vmware1tkg1, v1250vmware1tkg1),
						},
					}
					objects = append(objects, cm)
				})

				It("should return the listed TKRs", func() {
					compatibleVersions, err := c.getAdditionalCompatibleTKRVersions(context.Background())
					Expect(err).ToNot(HaveOccurred())
					Expect(compatibleVersions).To(Equal(sets.Strings(v1240vmware1tkg1, v1250vmware1tkg1)))
				})

				When("the tkrVersions field cannot be parsed", func() {
					BeforeEach(func() {
						cm.Data[fieldTKRVersions] = "garbage"
					})

					It("should return an empty set", func() {
						compatibleVersions, err := c.getAdditionalCompatibleTKRVersions(context.Background())
						Expect(err).ToNot(HaveOccurred())
						Expect(compatibleVersions).To(Equal(sets.Strings()))
					})
				})
			})
		})
	})

	Describe("CompatibleVersions()", func() {
		var (
			mc      *clusterv1.Cluster
			cm, cm1 *corev1.ConfigMap
		)

		BeforeEach(func() {
			mc = &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mc",
					Namespace: "not-important",
					Labels: map[string]string{
						constants.ManagementClusterRoleLabel: "",
					},
					Annotations: map[string]string{
						constants.TKGVersionKey: currentTKGVersion,
					},
				},
			}
			objects = append(objects, mc)
		})

		BeforeEach(func() {
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: c.Config.TKRNamespace,
					Name:      constants.BOMMetadataConfigMapName,
				},
				BinaryData: map[string][]byte{},
			}
			objects = append(objects, cm)
		})

		BeforeEach(func() {
			cm.BinaryData[constants.BOMMetadataCompatibilityKey] = []byte(bomMetadataStr)
		})

		BeforeEach(func() {
			cm1 = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: c.Config.TKRNamespace,
					Name:      "admin-byo-tkrs",
					Labels: map[string]string{
						LabelAdditionalTKRs: "",
					},
				},
				Data: map[string]string{
					fieldTKRVersions: fmt.Sprintf("[%s, %s]", v1240vmware1tkg1, v1250vmware1tkg1),
				},
			}
			objects = append(objects, cm1)
		})

		It("should return TKRs listed in both ways", func() {
			compatibleVersions, err := c.CompatibleVersions(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(compatibleVersions).To(Equal(sets.Strings(
				v1235vmware1tkg1zshippable,
				v1228vmware1tkg2zshippable,
				v12111vmware1tkg2zshippable,
				v1240vmware1tkg1,
				v1250vmware1tkg1,
			)))
		})

	})
})

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	return scheme
}

const (
	v1235vmware1tkg1zshippable  = "v1.23.5+vmware.1-tkg.1-zshippable"
	v1228vmware1tkg2zshippable  = "v1.22.8+vmware.1-tkg.2-zshippable"
	v12111vmware1tkg2zshippable = "v1.21.11+vmware.1-tkg.2-zshippable"
	v1240vmware1tkg1            = "v1.24.0+vmware.1-tkg.1"
	v1250vmware1tkg1            = "v1.25.0+vmware.1-tkg.1"
	currentTKGVersion           = "v1.6.0-zshippable"
)

var bomMetadataStr = fmt.Sprintf(`
managementClusterVersions:
- supportedKubernetesVersions:
  - v1.19.12+vmware.1-tkg.1
  - v1.21.2+vmware.1-tkg.1
  - v1.20.8+vmware.1-tkg.2
  version: v1.4.0
- supportedKubernetesVersions:
  - v1.21.2+vmware.1-tkg.2
  - v1.20.8+vmware.1-tkg.3
  - v1.19.12+vmware.1-tkg.2
  version: v1.4.1
- supportedKubernetesVersions:
  - v1.21.8+vmware.1-tkg.2
  - v1.19.16+vmware.1-tkg.1
  - v1.20.14+vmware.1-tkg.2
  version: v1.4.2
- supportedKubernetesVersions:
  - v1.22.5+vmware.1-tkg.1
  - v1.21.8+vmware.1-tkg.1
  - v1.20.14+vmware.1-tkg.1
  version: v1.5.0
- supportedKubernetesVersions:
  - v1.22.5+vmware.1-tkg.3
  - v1.21.8+vmware.1-tkg.4
  - v1.20.14+vmware.1-tkg.4
  version: v1.5.1
- supportedKubernetesVersions:
  - v1.22.5+vmware.1-tkg.4
  - v1.20.14+vmware.1-tkg.5
  - v1.21.8+vmware.1-tkg.5
  version: v1.5.2
- supportedKubernetesVersions:
  - v1.20.15+vmware.1-tkg.1
  - v1.22.8+vmware.1-tkg.1
  - v1.21.11+vmware.1-tkg.1
  version: v1.5.3
- supportedKubernetesVersions:
  - v1.20.15+vmware.1-tkg.2
  - v1.22.9+vmware.1-tkg.1
  - v1.21.11+vmware.1-tkg.3
  version: v1.5.4
- supportedKubernetesVersions:
  - %s
  - %s
  - %s
  version: %s
`,
	v1235vmware1tkg1zshippable,
	v1228vmware1tkg2zshippable,
	v12111vmware1tkg2zshippable,
	currentTKGVersion)
