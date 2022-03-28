// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

const TestPinnipedLabel = "pinniped.tanzu.vmware.com.1.2.3--vmware.1-tkg.1"

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
					constants.TKRLabelClassyClusters: "v1.23.3",
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
				Name:      fmt.Sprintf("%s-%s-package", cluster.Name, "pinniped.tanzu.vmware.com"),
				Labels: map[string]string{
					constants.PackageNameLabel:    TestPinnipedLabel,
					constants.TKGClusterNameLabel: cluster.Name,
				},
			},
			Type: constants.ClusterBootstrapManagedSecret,
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
				Eventually(cbSecretFunc(ctx, cluster, nil)).Should(Succeed())
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
					constants.TKGDataValueFieldName: []byte(fmt.Sprintf("%s: fire", constants.IdentityManagementTypeKey)),
				}
				secretCopy.Data = updatedSecretDataValues
				Expect(k8sClient.Update(ctx, secretCopy)).To(Succeed())
			})

			It("updates the secret with the proper data values", func() {
				Eventually(cbSecretFunc(ctx, cluster, nil)).Should(Succeed())
			})
		})

		When("random values are added to the secret", func() {
			var secretCopy *corev1.Secret
			BeforeEach(func() {
				secretCopy = pinnipedCBSecret.DeepCopy()
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secretCopy), secretCopy)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretCopy.Data).NotTo(BeNil())
				}).Should(Succeed())
				dataValues := secretCopy.Data[constants.TKGDataValueFieldName]
				dataValues = append(dataValues, "sweetest_cat: lionel"...)
				secretCopy.Data[constants.TKGDataValueFieldName] = dataValues
				Expect(k8sClient.Update(ctx, secretCopy)).To(Succeed())
			})

			It("they are preserved", func() {
				Eventually(func(g Gomega) {
					actualSecret := pinnipedCBSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					var gotValuesYAML map[string]interface{}
					var wantValuesYAML map[string]interface{}
					g.Expect(yaml.Unmarshal(actualSecret.Data[constants.TKGDataValueFieldName], &gotValuesYAML)).Should(Succeed())
					g.Expect(yaml.Unmarshal(secretCopy.Data[constants.TKGDataValueFieldName], &wantValuesYAML)).Should(Succeed())
					g.Expect(gotValuesYAML).Should(Equal(wantValuesYAML))
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
				secretCopy.Data[constants.TKGDataOverlayFieldName] = []byte(`#@ load("@ytt:overlay", "overlay")
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
				Expect(k8sClient.Update(ctx, secretCopy)).To(Succeed())
			})

			It("they are preserved", func() {
				Eventually(func(g Gomega) {
					actualSecret := pinnipedCBSecret.DeepCopy()
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSecret), actualSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(actualSecret.Data[constants.TKGDataOverlayFieldName]).Should(Equal(secretCopy.Data[constants.TKGDataOverlayFieldName]))
					Eventually(cbSecretFunc(ctx, cluster, nil)).Should(Succeed())
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
					constants.TKGAddonLabel:       "pinniped",
					constants.TKGClusterNameLabel: cluster.Name,
				}
				secretData = map[string][]byte{
					constants.TKGDataValueFieldName: []byte(fmt.Sprintf("%s: moses", constants.IdentityManagementTypeKey)),
				}
				secretCopy.Name = "another-secret"
				secretCopy.Labels = secretLabels
				secretCopy.Type = constants.ClusterBootstrapManagedSecret
				secretCopy.Data = secretData
				Expect(k8sClient.Create(ctx, secretCopy)).To(Succeed())
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
				secretCopy.Name = "newest-secret"
				secretLabels = map[string]string{
					constants.PackageNameLabel:    "pinniped.fun.times",
					constants.TKGClusterNameLabel: cluster.Name,
				}
				secretData = map[string][]byte{
					constants.TKGDataValueFieldName: []byte(fmt.Sprintf("%s: moses", constants.IdentityManagementTypeKey)),
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
		const (
			issuer             = "cats.meow"
			issuerCABundleData = "secret-blanket"
		)

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
					// TODO: do we want to add the other fields??
					constants.IssuerKey:         "tuna.io",
					constants.IssuerCABundleKey: "ball-of-fluff",
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
					Name:      fmt.Sprintf("%s-%s-package", cluster2.Name, "pinniped.tanzu.vmware.com"),
					Labels: map[string]string{
						constants.PackageNameLabel:    TestPinnipedLabel,
						constants.TKGClusterNameLabel: cluster2.Name,
					},
				},
				Type: constants.ClusterBootstrapManagedSecret,
			}

			clusters = []*clusterapiv1beta1.Cluster{cluster, cluster2}
			secrets = []*corev1.Secret{pinnipedCBSecret, secret2}
			createObject(ctx, configMap)

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
		When("the configmap gets created", func() {
			It("updates all the ClusterBootstrap secrets", func() {
				for _, c := range clusters {
					Eventually(cbSecretFunc(ctx, c, configMap)).Should(Succeed())
				}
			})
		})

		When("the configmap gets updated", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				configMapCopy = configMap.DeepCopy()
				configMapCopy.Data[constants.IssuerKey] = issuer
				configMapCopy.Data[constants.IssuerCABundleKey] = issuerCABundleData
				err := k8sClient.Update(ctx, configMapCopy)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(configMapCopy), configMapCopy)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
			})

			It("updates all the ClusterBootStrap secrets", func() {
				for _, c := range clusters {
					Eventually(cbSecretFunc(ctx, c, configMapCopy)).Should(Succeed())
				}
			})
		})

		When("the configmap gets does not have an issuer or caBundle", func() {
			var configMapCopy *corev1.ConfigMap
			BeforeEach(func() {
				configMapCopy = configMap.DeepCopy()
				delete(configMapCopy.Data, constants.IssuerKey)
				delete(configMapCopy.Data, constants.IssuerCABundleKey)
				err := k8sClient.Update(ctx, configMapCopy)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(configMapCopy), configMapCopy)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
			})

			It("updates all the ClusterBootStrap secrets", func() {
				for _, c := range clusters {
					Eventually(cbSecretFunc(ctx, c, nil)).Should(Succeed())
				}
			})
		})

		When("the configmap gets deleted", func() {
			BeforeEach(func() {
				deleteObject(ctx, configMap)
			})

			It("updates all the ClusterBootstrap secrets", func() {
				for _, c := range clusters {
					Eventually(cbSecretFunc(ctx, c, nil)).Should(Succeed())
				}
			})
		})

		When("a configmap in a different namespace gets created", func() {
			BeforeEach(func() {
				configMapCopy := configMap.DeepCopy()
				configMapCopy.Namespace = pinnipedNamespace
				configMapCopy.Data[constants.IssuerKey] = issuer
				configMapCopy.Data[constants.IssuerCABundleKey] = issuerCABundleData
				createObject(ctx, configMapCopy)
			})

			It("does not update secrets", func() {
				for _, c := range clusters {
					Eventually(cbSecretFunc(ctx, c, configMap)).Should(Succeed())
				}
			})
		})

		When("a configmap with a different name gets created", func() {
			BeforeEach(func() {
				configMapCopy := configMap.DeepCopy()
				configMapCopy.Name = "kitties"
				configMapCopy.Data = make(map[string]string)
				configMapCopy.Data[constants.IssuerKey] = issuer
				configMapCopy.Data[constants.IssuerCABundleKey] = issuerCABundleData
				createObject(ctx, configMapCopy)
			})

			It("does not update secrets", func() {
				for _, c := range clusters {
					Eventually(cbSecretFunc(ctx, c, configMap)).Should(Succeed())
				}
			})
		})
	})
})

func createObject(ctx context.Context, o client.Object) {
	oCopy := o.DeepCopyObject().(client.Object)
	err := k8sClient.Create(ctx, oCopy)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func(g Gomega) {
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(o), oCopy)
		g.Expect(err).NotTo(HaveOccurred())
	}).Should(Succeed())
}

func deleteObject(ctx context.Context, o client.Object) {
	err := k8sClient.Delete(ctx, o)

	// Accept cases where the object has already been deleted.
	if err != nil {
		Expect(k8serrors.IsNotFound(err)).To(BeTrue(), "got error: %#v", err)
	}

	oCopy := o.DeepCopyObject().(client.Object)
	Eventually(func(g Gomega) {
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(o), oCopy)
		g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
	}).Should(Succeed())
}

func cbSecretFunc(ctx context.Context, cluster *clusterapiv1beta1.Cluster, configMap *corev1.ConfigMap) func(Gomega) {
	return func(g Gomega) {
		clusterCopy := cluster.DeepCopy()
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), clusterCopy)
		g.Expect(err).NotTo(HaveOccurred())

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: clusterCopy.Namespace,
				Name:      fmt.Sprintf("%s-%s-package", clusterCopy.Name, "pinniped.tanzu.vmware.com"),
			},
		}

		wantSecretLabel := map[string]string{
			constants.PackageNameLabel:    TestPinnipedLabel,
			constants.TKGClusterNameLabel: clusterCopy.Name,
		}
		wantValuesYAML := map[string]interface{}{
			constants.IdentityManagementTypeKey: constants.None,
			"infrastructure_provider":           "vsphere",
			"tkg_cluster_role":                  "workload",
			"pinniped": map[string]interface{}{
				"concierge": map[string]interface{}{
					"audience": fmt.Sprintf("%s-%s", clusterCopy.Name, string(clusterCopy.UID)),
				},
			},
		}
		if configMap != nil {
			wantValuesYAML[constants.IdentityManagementTypeKey] = constants.OIDC

			m := wantValuesYAML["pinniped"].(map[string]interface{})
			m[constants.SupervisorEndpointKey] = configMap.Data[constants.IssuerKey]
			m[constants.SupervisorCABundleKey] = configMap.Data[constants.IssuerCABundleKey]
		}

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(secret.Labels).To(Equal(wantSecretLabel))
		var gotValuesYAML map[string]interface{}
		g.Expect(yaml.Unmarshal(secret.Data[constants.TKGDataValueFieldName], &gotValuesYAML)).Should(Succeed())
		g.Expect(gotValuesYAML).Should(Equal(wantValuesYAML))
	}
}
