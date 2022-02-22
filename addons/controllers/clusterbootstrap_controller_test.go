// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		s := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	When("cluster is created", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-tcbt"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap.yaml"
		})
		Context("with a ClusterBootstrapTemplate", func() {
			It("should create ClusterBootstrap CR and the related objects for the cluster", func() {

				By("verifying CAPI cluster is created properly")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: "default", Name: clusterName}, cluster)).To(Succeed())

				By("ClusterBootstrap CR is created with correct ownerReference added")
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

				By("verifying that ClusterBootstrap CR has status resolved to correct TKR")
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

				By("verifying that the providerRef from additionalPackages is cloned into cluster namespace and ownerReferences set properly")
				var gvr schema.GroupVersionResource
				var object *unstructured.Unstructured
				// Verify providerRef exists and also the cloned provider object with ownerReferences to cluster and ClusterBootstrap
				Eventually(func() bool {
					Expect(len(clusterBootstrap.Spec.AdditionalPackages) > 0).To(BeTrue())

					fooPackage := clusterBootstrap.Spec.AdditionalPackages[0]
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
						providerLabels[addontypes.PackageNameLabel] == util.ParseStringForLabel(fooPackage.RefName) {
						foundLabels = true
					}

					if foundClusterOwnerRef && foundClusterBootstrapOwnerRef && foundLabels {
						return true
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

				// Simulate a controller adding secretRef to provider status and
				// verify that a data-values secret has been created for the package
				By("patching foobar provider object's status resource with secret", func() {
					s := &corev1.Secret{}
					s.Name = "foobar-data-values"
					s.Namespace = "default"
					s.Data = map[string][]byte{}
					s.Data["values.yaml"] = []byte("foobar")
					Expect(k8sClient.Create(ctx, s)).To(Succeed())

					Expect(unstructured.SetNestedField(object.Object, s.Name, "status", "secretRef")).To(Succeed())

					_, err := dynamicClient.Resource(gvr).Namespace("default").UpdateStatus(ctx, object, metav1.UpdateOptions{})
					Expect(err).ToNot(HaveOccurred())

					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: fmt.Sprintf("%s-foobar-data-values", clusterName)}, s); err != nil {
							return false
						}
						if string(s.Data["values.yaml"]) != "foobar" {
							return false
						}

						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

				By("verifying that kapp-controller PackageInstall CR is created under cluster namespace properly on the management cluster")
				// Verify kapp-controller PackageInstall CR has been created under cluster namespace on management cluster
				kappControllerPkgi := &kapppkgiv1alpha1.PackageInstall{}
				Eventually(func() bool {
					if err := k8sClient.Get(ctx,
						client.ObjectKey{
							Namespace: "default",
							Name:      util.GeneratePackageInstallName(clusterName, "kapp-controller.tanzu.vmware.com"),
						}, kappControllerPkgi); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				Expect(len(kappControllerPkgi.OwnerReferences) == 1).To(BeTrue())
				Expect(kappControllerPkgi.OwnerReferences[0].APIVersion).To(Equal(clusterapiv1beta1.GroupVersion.String()))
				Expect(kappControllerPkgi.OwnerReferences[0].Name).To(Equal(cluster.Name))

				remoteClient, err := util.GetClusterClient(ctx, k8sClient, scheme, clusterapiutil.ObjectKey(cluster))
				Expect(err).NotTo(HaveOccurred())
				Expect(remoteClient).NotTo(BeNil())

				By("verifying that ServiceAccount, ClusterRole and ClusterRoleBinding are created on the workload cluster properly")
				sa := &corev1.ServiceAccount{}
				Eventually(func() bool {
					if err := remoteClient.Get(ctx,
						client.ObjectKey{Namespace: constants.TKGSystemNS, Name: "tanzu-addons-manager-sa"},
						sa); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				clusterRole := &rbacv1.ClusterRole{}
				Eventually(func() bool {
					if err := remoteClient.Get(ctx,
						client.ObjectKey{Name: "tanzu-addons-manager-clusterrole"},
						clusterRole); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
				Eventually(func() bool {
					if err := remoteClient.Get(ctx,
						client.ObjectKey{Name: "tanzu-addons-manager-clusterrolebinding"},
						clusterRoleBinding); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("verifying that Package CRs of additionalPackages are created on the workload cluster properly")
				pkg := &kapppkgv1alpha1.Package{}
				pkgRefNameMap := make(map[string]string)
				Eventually(func() bool {
					for _, clusterBootstrapPackage := range clusterBootstrap.Spec.AdditionalPackages {
						if err := remoteClient.Get(ctx,
							client.ObjectKey{Namespace: constants.TKGSystemNS, Name: clusterBootstrapPackage.RefName},
							pkg); err != nil {
							return false
						}
						pkgRefNameMap[clusterBootstrapPackage.RefName] = pkg.Spec.RefName
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("verifying that the data value secret is created on the workload cluster properly")
				remoteSecret := &corev1.Secret{}
				Eventually(func() bool {
					for _, clusterBootstrapPackage := range clusterBootstrap.Spec.AdditionalPackages {
						if err := remoteClient.Get(ctx,
							client.ObjectKey{
								Namespace: constants.TKGSystemNS,
								Name:      util.GenerateDataValueSecretName(cluster.Name, pkgRefNameMap[clusterBootstrapPackage.RefName]),
							}, remoteSecret); err != nil {
							return false
						}
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("verifying that the PackageInstall CRs are created on the workload cluster properly")
				remotePkgi := &kapppkgiv1alpha1.PackageInstall{}
				Eventually(func() bool {
					for _, clusterBootstrapPackage := range clusterBootstrap.Spec.AdditionalPackages {
						if err := remoteClient.Get(ctx,
							client.ObjectKey{
								Namespace: constants.TKGSystemNS,
								Name:      util.GeneratePackageInstallName(clusterName, pkgRefNameMap[clusterBootstrapPackage.RefName]),
							}, remotePkgi); err != nil {
							return false
						}
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				Expect(remotePkgi.Spec.PackageRef.RefName).To(Equal(pkg.Spec.RefName))
				Expect(len(remotePkgi.Spec.Values)).NotTo(BeZero())
				Expect(remotePkgi.Spec.Values[0].SecretRef.Name).To(Equal(util.GenerateDataValueSecretName(cluster.Name, pkg.Spec.RefName)))
			})
		})
	})

})
