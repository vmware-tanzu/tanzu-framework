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

const testTKRLabel = "v1.22.3"

var _ = Describe("Controller", func() {

	var cluster *clusterapiv1beta1.Cluster

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
	})

	Context("pinniped-info configmap", func() {

		var (
			configMap *corev1.ConfigMap
			clusters  []*clusterapiv1beta1.Cluster
		)
		BeforeEach(func() {
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
					secret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: c.Namespace,
							Name:      fmt.Sprintf("%s-pinniped-addon", c.Name),
						},
					}
					Eventually(func(g Gomega) {
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
						g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
					}).Should(Succeed())
				}
			})
		})
	})
})
