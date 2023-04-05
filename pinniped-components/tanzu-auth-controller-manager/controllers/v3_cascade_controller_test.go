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
)

const testPinnipedLabel = "pinniped.tanzu.vmware.com.1.2.3--vmware.1-tkg.1"

var _ = Describe("Controller", func() {
	var (
		cluster, managementCluster                          *clusterapiv1beta1.Cluster
		pinnipedCBSecret, managementClusterPinnipedCBSecret *corev1.Secret
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
					Kind: InfrastructureRefVSphere,
					Name: "some-name",
				},
			},
		}

		managementCluster = &clusterapiv1beta1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pinnipedNamespace,
				Name:      "some-management-cluster-name",
				Labels: map[string]string{
					tkrLabelClassyClusters: "v1.23.3",
					tkgManagementLabel:     "",
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

		// This Secret is to configure Pinniped on the management cluster, so this controller should never
		// edit this Secret, since it is not responsible for configuring Pinniped on management clusters.
		managementClusterPinnipedCBSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pinnipedNamespace,
				Name:      fmt.Sprintf("%s-pinniped.tanzu.vmware.com-package", managementCluster.Name),
				Labels: map[string]string{
					packageNameLabel:    testPinnipedLabel,
					tkgClusterNameLabel: managementCluster.Name,
				},
			},
			Data: map[string][]byte{tkgDataValueFieldName: []byte("should-not-be-edited")},
			Type: clusterBootstrapManagedSecret,
		}
	})

	Context("pinniped-info configmap does not exist", func() {
		BeforeEach(func() {
			createObject(ctx, cluster)
			createObject(ctx, pinnipedCBSecret)
		})

		AfterEach(func() {
			deleteObject(ctx, cluster)
			deleteObject(ctx, pinnipedCBSecret)
		})

		When("the pinniped cluster bootstrap secret gets created", func() {
			It("updates the secret with the default data values", func() {
				Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
			})
		})

		When("the pinniped cluster bootstrap secret gets updated by some other actor", func() {
			BeforeEach(func() {
				secretCopy := pinnipedCBSecret.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
				updatedSecretDataValues := map[string][]byte{
					tkgDataValueFieldName: []byte(fmt.Sprintf("%s: fire", identityManagementTypeKey)),
				}
				secretCopy.Data = updatedSecretDataValues
				updateObject(ctx, secretCopy)
			})

			It("resets the secret's data values back to the defaults", func() {
				Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
			})
		})

		When("the pinniped cluster bootstrap secret contains overlays", func() {
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

			It("preserves the overlays", func() {
				Consistently(func(g Gomega) {
					actualSecret := pinnipedCBSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data[tkgDataOverlayFieldName]).Should(Equal(secretCopy.Data[tkgDataOverlayFieldName]))
					Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		When("another similar looking secret does not have the Pinniped package label", func() {
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

			It("does not update that other similar secret", func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretCopy.Labels).To(Equal(secretLabels))
					g.Expect(secretCopy.Data).To(Equal(secretData))
				}).Should(Succeed())
			})
		})

		When("another similar looking secret is not a clusterbootstrap-secret type", func() {
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
				secretCopy.Type = "not-" + clusterBootstrapManagedSecret
				secretCopy.Data = secretData
				createObject(ctx, secretCopy)
			})

			AfterEach(func() {
				deleteObject(ctx, secretCopy)
			})

			It("does not update that other similar secret", func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretCopy.Labels).To(Equal(secretLabels))
					g.Expect(secretCopy.Data).To(Equal(secretData))
				}).Should(Succeed())
			})
		})
	})

	Context("pinniped-info configmap exists", func() {
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
						Kind: InfrastructureRefVSphere,
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

		Context("the pinniped-info configmap gets created", func() {
			When("there are no clusters and no pinniped cluster bootstrap secrets", func() {
				BeforeEach(func() {
					for _, c := range clusters {
						deleteObject(ctx, c)
					}
					for _, s := range secrets {
						deleteObject(ctx, s)
					}
					createObject(ctx, configMap)
				})

				It("does not create any pinniped cluster bootstrap secrets", func() {
					for _, c := range clusters {
						Eventually(verifyNoSecretFunc(ctx, c, false)).Should(Succeed())
					}
				})
			})

			When("there are clusters and pinniped cluster bootstrap secrets", func() {
				BeforeEach(func() {
					createObject(ctx, configMap)
				})

				It("updates all the pinniped cluster bootstrap secrets using the values from the configmap", func() {
					for _, c := range clusters {
						Eventually(verifySecretFunc(ctx, c, configMap, false)).Should(Succeed())
					}
				})
			})
		})

		When("the pinniped-info configmap gets updated", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				createObject(ctx, configMap)
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Data[issuerKey] = "cats.meow"
				configMapCopy.Data[issuerCABundleKey] = "secret-blanket"
				updateObject(ctx, configMapCopy)
			})

			It("updates all the pinniped cluster bootstrap secrets using the new values from the configmap", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMapCopy, false)).Should(Succeed())
				}
			})
		})

		When("the pinniped-info configmap does not have an issuer or caBundle", func() {
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

			It("updates all the pinniped cluster bootstrap secrets using an empty string for the issuer and CA values", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMapCopy, false)).Should(Succeed())
				}
			})
		})

		When("the pinniped-info configmap gets deleted", func() {
			BeforeEach(func() {
				createObject(ctx, configMap)
				deleteObject(ctx, configMap)
			})

			It("updates all the pinniped cluster bootstrap secrets to set their contents back to the default values", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, nil, false)).Should(Succeed())
				}
			})
		})

		When("a configmap in a different namespace (other than kube-public) gets created", func() {
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

			It("does not update the pinniped cluster bootstrap secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, false)).Should(Succeed())
				}
			})
		})

		When("a configmap with a different name (other than pinniped-info) gets created", func() {
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

			It("does not update the pinniped cluster bootstrap secrets", func() {
				for _, c := range clusters {
					Eventually(verifySecretFunc(ctx, c, configMap, false)).Should(Succeed())
				}
			})
		})

		When("there are no clusters and a pinniped-info configmap gets created", func() {
			BeforeEach(func() {
				for _, c := range clusters {
					deleteObject(ctx, c)
				}
				for _, s := range secrets {
					deleteObject(ctx, s)
				}
				createObject(ctx, configMap)
			})

			It("does not create any pinniped cluster bootstrap secrets", func() {
				for _, s := range secrets {
					Eventually(func(g Gomega) {
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(s), s)
						g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
					}).Should(Succeed())
				}
			})
		})
	})

	Context("management cluster exists", func() {
		BeforeEach(func() {
			createObject(ctx, cluster)
			createObject(ctx, managementCluster)
			createObject(ctx, pinnipedCBSecret)
			createObject(ctx, managementClusterPinnipedCBSecret)
		})

		AfterEach(func() {
			deleteObject(ctx, cluster)
			deleteObject(ctx, managementCluster)
			deleteObject(ctx, pinnipedCBSecret)
			deleteObject(ctx, managementClusterPinnipedCBSecret)
		})

		When("the pinniped cluster bootstrap secret gets created", func() {
			It("updates the non-management cluster secret with the default data values", func() {
				Eventually(verifySecretFunc(ctx, cluster, nil, false)).Should(Succeed())
			})

			It("does not change the management cluster bootstrap secret", func() {
				Consistently(func(g Gomega) {
					actualSecret := managementClusterPinnipedCBSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data).Should(Equal(managementClusterPinnipedCBSecret.Data))
				}).Should(Succeed())
			})
		})

		Context("the pinniped-info configmap also exists", func() {
			var configMap *corev1.ConfigMap
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
				createObject(ctx, configMap)
			})

			AfterEach(func() {
				deleteObject(ctx, configMap)
			})

			When("the configmap gets created", func() {
				It("updates the non-management cluster secret using the values from the configmap", func() {
					Eventually(verifySecretFunc(ctx, cluster, configMap, false)).Should(Succeed())
				})

				It("does not change the management cluster bootstrap secret", func() {
					Consistently(func(g Gomega) {
						actualSecret := managementClusterPinnipedCBSecret.DeepCopy()
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(actualSecret.Data).Should(Equal(managementClusterPinnipedCBSecret.Data))
					}).Should(Succeed())
				})
			})
		})
	})
})
