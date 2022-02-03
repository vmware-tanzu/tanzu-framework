// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("TanzuClusterBootstrap Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).To(Succeed())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, "default", k8sClient)).To(Succeed())
	})

	AfterEach(func() {
		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: "default",
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &v1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	Context("reconciletanzuClusterBootstrapNormal", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-tcbt"
			clusterResourceFilePath = "testdata/test-cluster-tanzu-cluster-bootstrap.yaml"
		})

		It("Should create clone TanzuClusterBootstrapTemplate and its related objects for the cluster and create package secret for foobar.example.com.1.17.2", func() {
			cluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: "default", Name: clusterName}, cluster)).To(Succeed())

			tanzuClusterBootstrap := &runtanzuv1alpha3.TanzuClusterBootstrap{}

			// Verify ownerReference for cluster in cloned object
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), tanzuClusterBootstrap)
				if err != nil {
					return false
				}
				for _, ownerRef := range tanzuClusterBootstrap.OwnerReferences {
					if ownerRef.UID == cluster.UID {
						return true
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// Verify ResolvedTKR
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), tanzuClusterBootstrap)
				if err != nil {
					return false
				}
				if tanzuClusterBootstrap.Status.ResolvedTKR == "v1.22.3" {
					return true
				}

				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			var gvr schema.GroupVersionResource
			var object *unstructured.Unstructured

			// Verify providerRef exists and also the cloned provider object with ownerReferences to cluster and  TanzuClusterBootstrap
			Eventually(func() bool {
				Expect(len(tanzuClusterBootstrap.Spec.AdditionalPackages) > 0).To(BeTrue())

				fooPackage := tanzuClusterBootstrap.Spec.AdditionalPackages[0]
				Expect(fooPackage.RefName == "foobar.example.com.1.17.2").To(BeTrue())
				Expect(*fooPackage.ValuesFrom.ProviderRef.APIGroup == "run.tanzu.vmware.com").To(BeTrue())
				Expect(fooPackage.ValuesFrom.ProviderRef.Kind == "FooBar").To(BeTrue())

				providerName := fmt.Sprintf("%s-foobar-package", clusterName)
				Expect(fooPackage.ValuesFrom.ProviderRef.Name == providerName).To(BeTrue())

				gvr = schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha1", Resource: "foobars"}
				var err error
				object, err = dynamicClient.Resource(gvr).Namespace("default").Get(ctx, providerName, metav1.GetOptions{})

				Expect(err).ToNot(HaveOccurred())

				var foundClusterOwnerRef bool
				var foundTanzuClusterBootstrapOwnerRef bool
				for _, ownerRef := range object.GetOwnerReferences() {
					if ownerRef.UID == cluster.UID {
						foundClusterOwnerRef = true
					}
					if ownerRef.UID == tanzuClusterBootstrap.UID {
						foundTanzuClusterBootstrapOwnerRef = true
					}
				}

				if foundClusterOwnerRef && foundTanzuClusterBootstrapOwnerRef {
					return true
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// Simulate a controller adding secretRef to provider status and
			// verify that a data-values secret has been created for the package
			By("patching foobar status resource with secret", func() {
				s := &v1.Secret{}
				s.Name = "foobar-data-values"
				s.Namespace = "default"
				s.Data = map[string][]byte{}
				s.Data["values.yaml"] = []byte("foobar")
				Expect(k8sClient.Create(ctx, s)).To(Succeed())

				Expect(unstructured.SetNestedField(object.Object, s.Name, "status", "secretRef")).To(Succeed())

				_, err := dynamicClient.Resource(gvr).Namespace("default").UpdateStatus(ctx, object, metav1.UpdateOptions{})
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() bool {
					s := &v1.Secret{}
					if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "tkg-system", Name: fmt.Sprintf("%s-foobar-data-values", clusterName)}, s); err != nil {
						return false
					}
					if string(s.Data["values.yaml"]) != "foobar" {
						return false
					}

					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
			})

		})
	})

})
