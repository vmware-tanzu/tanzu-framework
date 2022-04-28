// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

const testPinnipedLabel = "pinniped.tanzu.vmware.com.1.2.3--vmware.1-tkg.1"

var _ = Describe("Controller", func() {
	var (
		cluster          *clusterapiv1beta1.Cluster
		pinnipedCBSecret *corev1.Secret
	)

	BeforeEach(func() {
		cluster = &clusterapiv1beta1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pinnipedNamespace,
				Name:      "some-name",
				Labels: map[string]string{
					tkrLabelClassyClusters: "v1.23.3",
				},
			},
			Spec: clusterapiv1beta1.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					Kind: tkgconstants.InfrastructureRefVSphere,
					Name: "some-name",
				},
			},
		}

		pinnipedCBSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pinnipedNamespace,
				Name:      fmt.Sprintf("%s-pinniped.tanzu.vmware.com-package", cluster.Name),
				Labels: map[string]string{
					packageNameLabel:    testPinnipedLabel,
					tkgClusterNameLabel: cluster.Name,
				},
			},
			Type: clusterBootstrapManagedSecret,
		}
	})

	Context("ClusterBootstrap Secret", func() {
		BeforeEach(func() {
			createObject(ctx, cluster)
			createObject(ctx, pinnipedCBSecret)
		})

		AfterEach(func() {
			deleteObject(ctx, cluster)
			deleteObject(ctx, pinnipedCBSecret)
		})

		When("the secret gets created", func() {
			It("updates the secret with the proper data values", func() {
				Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
			})
		})

		When("the secret gets updated", func() {
			BeforeEach(func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pinnipedCBSecret), pinnipedCBSecret)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
				secretCopy := pinnipedCBSecret.DeepCopy()
				updatedSecretDataValues := map[string][]byte{
					tkgDataValueFieldName: []byte(fmt.Sprintf("%s: fire", identityManagementTypeKey)),
				}
				secretCopy.Data = updatedSecretDataValues
				updateObject(ctx, secretCopy)
			})

			It("updates the secret with the proper data values", func() {
				Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
			})
		})

		When("random values are added to the secret", func() {
			var expectedSecret *corev1.Secret
			BeforeEach(func() {
				expectedSecret = pinnipedCBSecret.DeepCopy()
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
					actualSecret := pinnipedCBSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data).Should(Equal(expectedSecret.Data))
				}).Should(Succeed())
			})
		})

		When("the secret contains overlays", func() {
			var secretCopy *corev1.Secret
			BeforeEach(func() {
				secretCopy = pinnipedCBSecret.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretCopy.Data).NotTo(BeNil())
				}).Should(Succeed())
				secretCopy.Data[tkgDataOverlayFieldName] = []byte(`#@ load("@ytt:overlay", "overlay")
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
				updateObject(ctx, secretCopy)
			})

			It("they are preserved", func() {
				Eventually(func(g Gomega) {
					actualSecret := pinnipedCBSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data[tkgDataOverlayFieldName]).Should(Equal(secretCopy.Data[tkgDataOverlayFieldName]))
					Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		When("the secret does not have the Pinniped package label", func() {
			var secretCopy *corev1.Secret
			var secretLabels map[string]string
			var secretData map[string][]byte

			BeforeEach(func() {
				secretCopy = pinnipedCBSecret.DeepCopy()
				secretLabels = map[string]string{
					tkgAddonLabel:       "pinniped",
					tkgClusterNameLabel: cluster.Name,
				}
				secretData = map[string][]byte{
					tkgDataValueFieldName: []byte(fmt.Sprintf("%s: moses", identityManagementTypeKey)),
				}
				secretCopy.Name = "another-one"
				secretCopy.Labels = secretLabels
				secretCopy.Type = clusterBootstrapManagedSecret
				secretCopy.Data = secretData
				createObject(ctx, secretCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, secretCopy)
			})

			It("does not get updated", func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretCopy.Labels).To(Equal(secretLabels))
					g.Expect(secretCopy.Data).To(Equal(secretData))
				}).Should(Succeed())
			})
		})

		When("the secret is not a ClusterBootstrap secret type", func() {
			var secretCopy *corev1.Secret
			var secretLabels map[string]string
			var secretData map[string][]byte

			BeforeEach(func() {
				secretCopy = pinnipedCBSecret.DeepCopy()
				secretCopy.Name = "newest-one"
				secretLabels = map[string]string{
					packageNameLabel:    "pinniped.fun.times",
					tkgClusterNameLabel: cluster.Name,
				}
				secretData = map[string][]byte{
					tkgDataValueFieldName: []byte(fmt.Sprintf("%s: moses", identityManagementTypeKey)),
				}
				secretCopy.Labels = secretLabels
				secretCopy.Type = "not-an-cb-managed-secret"
				secretCopy.Data = secretData
				createObject(ctx, secretCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, secretCopy)
			})

			It("does not get updated", func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretCopy.Labels).To(Equal(secretLabels))
					g.Expect(secretCopy.Data).To(Equal(secretData))
				}).Should(Succeed())
			})
		})
	})

	Context("pinniped-info configmap", func() {
		var (
			clusters  []*clusterapiv1beta1.Cluster
			configMap *corev1.ConfigMap
			secrets   []*corev1.Secret
		)
		BeforeEach(func() {
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-public",
					Name:      "pinniped-info",
				},
				Data: map[string]string{
					issuerKey:         "tuna.io",
					issuerCABundleKey: "ball-of-fluff",
				},
			}

			cluster2 := &clusterapiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      "another-name",
				},
				Spec: clusterapiv1beta1.ClusterSpec{
					InfrastructureRef: &corev1.ObjectReference{
						Kind: tkgconstants.InfrastructureRefVSphere,
						Name: "another-name",
					},
				},
			}

			secret2 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      fmt.Sprintf("%s-pinniped.tanzu.vmware.com-package", cluster2.Name),
					Labels: map[string]string{
						packageNameLabel:    testPinnipedLabel,
						tkgClusterNameLabel: cluster2.Name,
					},
				},
				Type: clusterBootstrapManagedSecret,
			}

			clusters = []*clusterapiv1beta1.Cluster{cluster, cluster2}
			secrets = []*corev1.Secret{pinnipedCBSecret, secret2}

			for _, c := range clusters {
				createObject(ctx, c)
			}

			for _, s := range secrets {
				createObject(ctx, s)
			}
		})

		AfterEach(func() {
			deleteObject(ctx, configMap)
			for _, c := range clusters {
				deleteObject(ctx, c)
			}

			for _, s := range secrets {
				deleteObject(ctx, s)
			}
		})
		Context("the configmap gets created", func() {
			When("there are no ClusterBootstrap secrets", func() {
				BeforeEach(func() {
					for _, c := range clusters {
						deleteObject(ctx, c)
					}
					for _, s := range secrets {
						deleteObject(ctx, s)
					}
					createObject(ctx, configMap)
				})

				It("does not create any secrets", func() {
					for _, c := range clusters {
						Eventually(verifyNoSecretFunc(ctx, c, false)).Should(Succeed())
					}
				})
			})

			When("there are ClusterBootstrap secrets", func() {
				BeforeEach(func() {
					createObject(ctx, configMap)
				})

				It("updates all the ClusterBootstrap secrets", func() {
					for _, c := range clusters {
						Eventually(verifySecretFunc(ctx, c, configMap, false)).Should(Succeed())
					}
				})
			})
		})

		When("the configmap gets updated", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Data[issuerKey] = "cats.meow"
				configMapCopy.Data[issuerCABundleKey] = "secret-blanket"
				updateObject(ctx, configMapCopy)
			})

			It("updates all the ClusterBootStrap secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMapCopy, false)).Should(Succeed())
				}
			})
		})

		When("the configmap does not have an issuer or caBundle", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Data["fakeData"] = "fake"
				delete(configMapCopy.Data, issuerKey)
				delete(configMapCopy.Data, issuerCABundleKey)
				createObject(ctx, configMapCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, configMapCopy)
			})

			It("passes through an empty string for the value", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMapCopy, false)).Should(Succeed())
				}
			})
		})

		When("the configmap gets deleted", func() {
			BeforeEach(func() {
				createObject(ctx, configMap)
				deleteObject(ctx, configMap)
			})

			It("updates all the ClusterBootstrap secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, nil, false)).Should(Succeed())
				}
			})
		})

		When("a configmap in a different namespace gets created", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Namespace = pinnipedNamespace
				configMapCopy.Data[issuerKey] = "lionel.cat"
				configMapCopy.Data[issuerCABundleKey] = "catdog"
				createObject(ctx, configMapCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, configMapCopy)
			})

			It("does not update secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, false)).Should(Succeed())
				}
			})
		})

		When("a configmap with a different name gets created", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Name = "kitties"
				configMapCopy.Data = make(map[string]string)
				configMapCopy.Data[issuerKey] = "pumpkin.me"
				configMapCopy.Data[issuerCABundleKey] = "punkiest"
				createObject(ctx, configMapCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, configMapCopy)
			})

			It("does not update secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, false)).Should(Succeed())
				}
			})
		})

		When("there are no clusters and a configmap gets created", func() {
			BeforeEach(func() {
				for _, c := range clusters {
					deleteObject(ctx, c)
				}
				for _, s := range secrets {
					deleteObject(ctx, s)
				}
				createObject(ctx, configMap)
			})

			It("does not create secrets", func() {
				for _, s := range secrets {
					Eventually(func(g Gomega) {
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(s), s)
						g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
					}).Should(Succeed())
				}
			})
		})
	})
})
