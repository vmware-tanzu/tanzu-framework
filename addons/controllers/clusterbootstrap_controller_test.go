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

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("ClusterBootstrap Reconciler", func() {
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

	Context("reconcileClusterBootstrapNormal", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-tcbt"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-1.yaml"
		})

		It("Should create clone ClusterBootstrapTemplate and its related objects for the cluster, create package "+
			"secret for foobar.example.com.1.17.2 and set CNI to the one specified using cluster variable", func() {
			cluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: "default", Name: clusterName}, cluster)).To(Succeed())

			clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}

			// Verify ownerReference for cluster in cloned object
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
				if err != nil {
					return false
				}
				for _, ownerRef := range clusterBootstrap.OwnerReferences {
					if ownerRef.UID == cluster.UID {
						return true
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// Verify CNI is populated in the cloned object with the value from the cluster variables definitions
			Expect(len(clusterBootstrap.Spec.CNIs)).NotTo(BeZero())
			cni := clusterBootstrap.Spec.CNIs[0]
			Expect(packageShortName(cni.RefName)).To(Equal("antrea"))

			Expect(cni.RefName).To(Equal("antrea.tanzu.vmware.com.21.2.0--vmware.1-tkg.1-rc.1"))
			Expect(*cni.ValuesFrom.ProviderRef.APIGroup).To(Equal("cni.tanzu.vmware.com"))
			Expect(cni.ValuesFrom.ProviderRef.Kind).To(Equal("AntreaConfig"))
			providerName := fmt.Sprintf("%s-antrea-package", clusterName)
			Expect(cni.ValuesFrom.ProviderRef.Name).To(Equal(providerName))

			// Verify ResolvedTKR
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
				if err != nil {
					return false
				}
				if clusterBootstrap.Status.ResolvedTKR == "v1.22.3" {
					return true
				}

				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			var gvr schema.GroupVersionResource
			var object *unstructured.Unstructured
			var fooPackage *runtanzuv1alpha3.ClusterBootstrapPackage
			var foobar1Package *runtanzuv1alpha3.ClusterBootstrapPackage

			// Verify providerRef exists and also the cloned provider object with ownerReferences to cluster and ClusterBootstrap
			Eventually(func() bool {
				Expect(len(clusterBootstrap.Spec.AdditionalPackages) > 1).To(BeTrue())

				fooPackage = clusterBootstrap.Spec.AdditionalPackages[1]
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
				var foundClusterBootstrapOwnerRef bool
				var foundLabels bool
				for _, ownerRef := range object.GetOwnerReferences() {
					if ownerRef.UID == cluster.UID {
						foundClusterOwnerRef = true
					}
					if ownerRef.UID == clusterBootstrap.UID {
						foundClusterBootstrapOwnerRef = true
					}
				}
				providerLabels := object.GetLabels()
				if providerLabels[addontypes.ClusterNameLabel] == clusterName &&
					providerLabels[addontypes.PackageNameLabel] == fooPackage.RefName {
					foundLabels = true
				}

				if foundClusterOwnerRef && foundClusterBootstrapOwnerRef && foundLabels {
					return true
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// Verify that secret object mapping to foobar1 package exists with ownerReferences to cluster and ClusterBootstrap
			Eventually(func() bool {
				foobar1Package = clusterBootstrap.Spec.AdditionalPackages[0]

				Expect(foobar1Package.RefName == "foobar1.example.com.1.17.2").To(BeTrue())
				secretName := fmt.Sprintf("%s-foobar1-package", clusterName)
				Expect(foobar1Package.ValuesFrom.SecretRef == secretName).To(BeTrue())

				s := &v1.Secret{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: "default", Name: secretName}, s)).To(Succeed())

				var foundClusterOwnerRef bool
				var foundClusterBootstrapOwnerRef bool
				var foundLabels bool
				var foundType bool
				for _, ownerRef := range s.GetOwnerReferences() {
					if ownerRef.UID == cluster.UID {
						foundClusterOwnerRef = true
					}
					if ownerRef.UID == clusterBootstrap.UID {
						foundClusterBootstrapOwnerRef = true
					}
				}
				secretLabels := s.GetLabels()
				if secretLabels[addontypes.ClusterNameLabel] == clusterName &&
					secretLabels[addontypes.PackageNameLabel] == foobar1Package.RefName {
					foundLabels = true
				}
				if s.Type == constants.ClusterBootstrapManagedSecret {
					foundType = true
				}
				if foundClusterOwnerRef && foundClusterBootstrapOwnerRef && foundLabels && foundType {
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

	Context("reconcileClusterBootstrapNormal, CNI selection in case of no CNI cluster variable", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-tcbt-2"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-2.yaml"
		})

		It("Should create cloned ClusterBootstrap for the cluster and set CNI to the first entry in the template (as cluster variable for CNI is not set)", func() {
			cluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: "default", Name: clusterName}, cluster)).To(Succeed())

			clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}

			// Verify ownerReference for cluster in cloned object
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
				if err != nil {
					return false
				}
				for _, ownerRef := range clusterBootstrap.OwnerReferences {
					if ownerRef.UID == cluster.UID {
						return true
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// Verify CNI is populated in the cloned object with the the first entry in the template (as cluster variable for CNI is not set)
			Expect(len(clusterBootstrap.Spec.CNIs)).To(Equal(1))
			cni := clusterBootstrap.Spec.CNIs[0]
			Expect(packageShortName(cni.RefName)).To(Equal("calico"))

			Expect(cni.RefName).To(Equal("calico.tanzu.vmware.com.0.3.0--vmware.1-tkg.1-rc.1"))
			Expect(*cni.ValuesFrom.ProviderRef.APIGroup).To(Equal("cni.tanzu.vmware.com"))
			Expect(cni.ValuesFrom.ProviderRef.Kind).To(Equal("CalicoConfig"))
			providerName := fmt.Sprintf("%s-calico-package", clusterName)
			Expect(cni.ValuesFrom.ProviderRef.Name).To(Equal(providerName))
		})
	})
})
