// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

const testTKRLabel = "v1.22.3"

var _ = Describe("Controller", func() {

	var (
		cluster   *clusterapiv1beta1.Cluster
		configMap *corev1.ConfigMap
		secret    *corev1.Secret
	)

	BeforeEach(func() {
		cluster = &clusterapiv1beta1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pinnipedNamespace,
				Name:      "some-name",
				Labels: map[string]string{
					tkrLabel: testTKRLabel,
				},
			},
			Spec: clusterapiv1beta1.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					Kind: tkgconstants.InfrastructureRefVSphere,
					Name: "some-name",
				},
			},
		}

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-public",
				Name:      "pinniped-info",
			},
			Data: map[string]string{
				"issuer":                "tuna.io",
				"issuer_ca_bundle_data": "ball-of-fluff",
			},
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cluster.Namespace,
				Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
			},
		}
	})

	AfterEach(func() {
		deleteObject(ctx, secret)
	})

	Context("Cluster", func() {
		BeforeEach(func() {
			createObject(ctx, cluster)
		})

		AfterEach(func() {
			deleteObject(ctx, cluster)
		})

		Context("cluster is created", func() {
			When("there is no pinniped-info configmap", func() {
				It("does not create a secret", func() {
					Eventually(verifyNoSecretFunc(ctx, cluster, true)).Should(Succeed())
				})
			})

			When("there is a pinniped-info configmap", func() {
				BeforeEach(func() {
					createObject(ctx, configMap)
				})

				AfterEach(func() {
					deleteObject(ctx, configMap)
				})

				It("creates a secret with information from configmap", func() {
					Eventually(verifySecretFunc(ctx, cluster, configMap, true)).Should(Succeed())
				})
			})
		})

		Context("cluster is updated", func() {
			BeforeEach(func() {
				clusterCopy := cluster.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), clusterCopy)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
				annotations := clusterCopy.ObjectMeta.Annotations
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations["sweetest-cat"] = "lionel"
				clusterCopy.ObjectMeta.Annotations = annotations
				updateObject(ctx, clusterCopy)
			})

			When("there is no pinniped-info configmap", func() {
				It("does not create a secret", func() {
					Eventually(verifyNoSecretFunc(ctx, cluster, true)).Should(Succeed())
				})
			})

			When("there is a pinniped-info configmap", func() {
				BeforeEach(func() {
					createObject(ctx, configMap)
				})

				AfterEach(func() {
					deleteObject(ctx, configMap)
				})

				It("updates the secret with information from configmap", func() {
					Eventually(verifySecretFunc(ctx, cluster, configMap, true)).Should(Succeed())
				})
			})
		})

		Context("cluster is deleted", func() {
			When("there is no pinniped-info configmap", func() {
				BeforeEach(func() {
					deleteObject(ctx, cluster)
				})

				It("does not create a secret", func() {
					Eventually(verifyNoSecretFunc(ctx, cluster, true)).Should(Succeed())
				})
			})

			When("there is a pinniped-info configmap", func() {
				BeforeEach(func() {
					createObject(ctx, configMap)
					Eventually(verifySecretFunc(ctx, cluster, configMap, true)).Should(Succeed())
					deleteObject(ctx, cluster)
				})

				AfterEach(func() {
					deleteObject(ctx, configMap)
				})

				It("deletes the secret", func() {
					Eventually(verifyNoSecretFunc(ctx, cluster, true)).Should(Succeed())
				})
			})
		})
	})

	Context("Addon Secret", func() {
		BeforeEach(func() {
			createObject(ctx, configMap)
			createObject(ctx, cluster)
		})

		AfterEach(func() {
			deleteObject(ctx, configMap)
			deleteObject(ctx, cluster)
		})

		When("the secret is deleted", func() {
			It("recreates the secret with values from the configmap", func() {
				Eventually(func(g Gomega) {
					Eventually(verifySecretFunc(ctx, cluster, configMap, true)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		When("essential secret data values are changed", func() {
			var secretCopy *corev1.Secret
			BeforeEach(func() {
				secretCopy = secret.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
				updatedSecretDataValues := &pinnipedDataValues{}
				updatedSecretDataValues.Pinniped.SupervisorEndpoint = "marshmallow.fluff"
				updatedSecretDataValues.Pinniped.SupervisorCABundle = "hairball"
				updatedSecretDataValues.Pinniped.Concierge.Audience = "animal-kingdom"
				updatedSecretDataValues.Infrastructure = "cat-tree"
				updatedSecretDataValues.ClusterRole = "lounge"
				updatedSecretDataValues.IdentityManagementType = "meow"
				dataValueYamlBytes, _ := yaml.Marshal(updatedSecretDataValues)
				secretCopy.Data[tkgDataValueFieldName] = dataValueYamlBytes
				updateObject(ctx, secretCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, secretCopy)
			})

			It("resets the secret with the proper data values", func() {
				Eventually(verifySecretFunc(ctx, cluster, configMap, true)).Should(Succeed())
			})
		})

		When("random values are added to the secret", func() {
			var expectedSecret *corev1.Secret
			BeforeEach(func() {
				expectedSecret = secret.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(expectedSecret), expectedSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(expectedSecret.Data).NotTo(BeNil())
				}).Should(Succeed())
				dataValues := expectedSecret.Data[tkgDataValueFieldName]
				dataValues = append(dataValues, "sweetest_cat: lionel"...)
				expectedSecret.Data[tkgDataValueFieldName] = dataValues
				updateObject(ctx, expectedSecret)
			})

			It("they are preserved", func() {
				Eventually(func(g Gomega) {
					actualSecret := expectedSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data).Should(Equal(expectedSecret.Data))
				}).Should(Succeed())
			})
		})

		When("the secret contains overlays", func() {
			var expectedSecret *corev1.Secret
			BeforeEach(func() {
				expectedSecret = secret.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(expectedSecret), expectedSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(expectedSecret.Data).NotTo(BeNil())
				}).Should(Succeed())
				expectedSecret.Data[tkgDataOverlayFieldName] = []byte(`#@ load("@ytt:overlay", "overlay")
   #@overlay/match by=overlay.subset({"kind": "Service", "metadata": {"name": "pinniped-supervisor", "namespace": "pinniped-supervisor"}})
   ---
   #@overlay/replace
   spec:
     type: LoadBalancer
     selector:
       app: pinniped-supervisor
     ports:
       - name: https
         protocol: TCP
         port: 443
         targetPort: 8443`)
				updateObject(ctx, expectedSecret)
			})

			It("they are preserved", func() {
				Eventually(func(g Gomega) {
					actualSecret := secret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data[tkgDataOverlayFieldName]).Should(Equal(expectedSecret.Data[tkgDataOverlayFieldName]))
					Eventually(verifySecretFunc(ctx, cluster, configMap, true)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		When("the secret does not have the Pinniped addon label", func() {
			var expectedSecret *corev1.Secret

			BeforeEach(func() {
				expectedSecret = secret.DeepCopy()
				expectedSecret.Name = "another-secret"
				expectedSecret.Type = "tkg.tanzu.vmware.com/addon"
				expectedSecret.Labels = map[string]string{
					tkgAddonLabel:       "pumpkin",
					tkgClusterNameLabel: cluster.Name,
				}
				expectedSecret.Data = map[string][]byte{
					"values.yaml": []byte("identity_management_type: moses"),
				}
				createObject(ctx, expectedSecret)
			})

			AfterEach(func() {
				deleteObject(ctx, expectedSecret)
			})

			It("does not get updated", func() {
				Eventually(func(g Gomega) {
					actualSecret := expectedSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Labels).To(Equal(expectedSecret.Labels))
					g.Expect(actualSecret.Data).To(Equal(expectedSecret.Data))
				}).Should(Succeed())
			})
		})

		When("the secret is not an addon type", func() {
			var expectedSecret *corev1.Secret

			BeforeEach(func() {
				expectedSecret = secret.DeepCopy()
				expectedSecret.Type = "not-an-addon"
				expectedSecret.Labels = map[string]string{
					tkgAddonLabel:       pinnipedAddonLabel,
					tkgClusterNameLabel: cluster.Name,
				}
				expectedSecret.Name = "newest-secret"
				expectedSecret.Data = map[string][]byte{
					"values.yaml": []byte("identity_management_type: moses"),
				}
				createObject(ctx, expectedSecret)
			})

			AfterEach(func() {
				deleteObject(ctx, expectedSecret)
			})

			It("does not get updated", func() {
				Eventually(func(g Gomega) {
					actualSecret := expectedSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Labels).To(Equal(expectedSecret.Labels))
					g.Expect(actualSecret.Type).To(Equal(expectedSecret.Type))
					g.Expect(actualSecret.Data).To(Equal(expectedSecret.Data))
				}).Should(Succeed())
			})
		})
	})

	Context("pinniped-info configmap", func() {

		var clusters []*clusterapiv1beta1.Cluster

		BeforeEach(func() {
			cluster2 := &clusterapiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      "another-name",
					Labels: map[string]string{
						tkrLabel: testTKRLabel,
					},
				},
				Spec: clusterapiv1beta1.ClusterSpec{
					InfrastructureRef: &corev1.ObjectReference{
						Kind: tkgconstants.InfrastructureRefVSphere,
						Name: "another-name",
					},
				},
			}

			clusters = []*clusterapiv1beta1.Cluster{cluster, cluster2}

			for _, c := range clusters {
				createObject(ctx, c)
			}
		})

		AfterEach(func() {
			deleteObject(ctx, configMap)

			for _, c := range clusters {
				deleteObject(ctx, c)
			}
		})
		When("the configmap gets created", func() {
			BeforeEach(func() {
				createObject(ctx, configMap)
			})

			It("creates all the addons secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, true)).Should(Succeed())
				}
			})
		})

		When("the configmap gets updated", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Data[issuerKey] = "cats.dev"
				configMapCopy.Data[issuerCABundleKey] = "cattree"
				updateObject(ctx, configMapCopy)
			})

			It("updates all the addon secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMapCopy, true)).Should(Succeed())
				}
			})
		})

		When("the configmap does not have an issuer or caBundle", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Data["fake_data"] = "something-fake"
				delete(configMapCopy.Data, issuerKey)
				delete(configMapCopy.Data, issuerCABundleKey)
				createObject(ctx, configMapCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, configMapCopy)
			})

			It("passes through an empty string for the value", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMapCopy, true)).Should(Succeed())
				}
			})
		})

		When("a configmap in a different namespace gets created", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Namespace = pinnipedNamespace
				configMapCopy.Data["issuer"] = "moses.org"
				configMapCopy.Data["issuer_ca_bundle_data"] = "laziest"
				createObject(ctx, configMapCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, configMapCopy)
			})

			It("does not update addon secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, true)).Should(Succeed())
				}
			})
		})
		When("a configmap with a different name gets created", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Name = "sweet-potato"
				configMapCopy.Data = make(map[string]string)
				configMapCopy.Data["issuer"] = "zelda.cat"
				configMapCopy.Data["issuer_ca_bundle_data"] = "zeldz"
				createObject(ctx, configMapCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, configMapCopy)
			})

			It("does not update addon secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, true)).Should(Succeed())
				}
			})
		})

		When("there are no clusters and a configmap gets created", func() {
			BeforeEach(func() {
				for _, c := range clusters {
					deleteObject(ctx, c)
					secret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: c.Namespace,
							Name:      fmt.Sprintf("%s-pinniped-addon", c.Name),
						},
					}
					deleteObject(ctx, secret)
				}
				createObject(ctx, configMap)
			})

			It("does not create addon secrets", func() {
				for _, c := range clusters {
					Eventually(verifyNoSecretFunc(ctx, c, true)).Should(Succeed())
				}
			})
		})
	})
})
