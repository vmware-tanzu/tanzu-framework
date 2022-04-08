// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	vspherecpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
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

var _ = FDescribe("ClusterBootstrap Reconciler", func() {
	var (
		clusterName             string
		clusterNamespace        string
		clusterResourceFilePath string
	)

	// Constants defined in testdata manifests
	const (
		foobarCarvelPackageRefName  = "foobar.example.com"
		foobarCarvelPackageName     = "foobar.example.com.1.17.2"
		foobar1CarvelPackageRefName = "foobar1.example.com"
		foobar1CarvelPackageName    = "foobar1.example.com.1.17.2"
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).To(Succeed())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, clusterNamespace, k8sClient)).To(Succeed())
	})

	AfterEach(func() {
		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: clusterNamespace,
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	When("cluster is created with topology", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-tcbt"
			clusterNamespace = "cluster-namespace"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-1.yaml"
		})
		Context("from a ClusterBootstrapTemplate", func() {
			It("should create ClusterBootstrap CR and the related objects for the cluster", func() {

				By("verifying CAPI cluster is created properly")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				copiedCluster := cluster.DeepCopy()
				copiedCluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, copiedCluster)).To(Succeed())

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

				By("verifying that CNI has been populated properly")
				// Verify CNI is populated in the cloned object with the value from the cluster bootstrap template
				Expect(clusterBootstrap.Spec.CNI).NotTo(BeNil())
				Expect(strings.HasPrefix(clusterBootstrap.Spec.CNI.RefName, "antrea")).To(BeTrue())

				Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal("antrea.tanzu.vmware.com.1.2.3--vmware.1-tkg.1"))
				Expect(*clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.APIGroup).To(Equal("cni.tanzu.vmware.com"))
				Expect(clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Kind).To(Equal("AntreaConfig"))
				providerName := fmt.Sprintf("%s-antrea.tanzu.vmware.com-package", clusterName)
				Expect(clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Name).To(Equal(providerName))

				By("verifying that the proxy related annotations are populated to cluster object properly")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), cluster)
					if err != nil {
						return false
					}
					if cluster.Annotations != nil &&
						cluster.Annotations[addontypes.HTTPProxyConfigAnnotation] == "foo.com" &&
						cluster.Annotations[addontypes.HTTPSProxyConfigAnnotation] == "bar.com" &&
						cluster.Annotations[addontypes.NoProxyConfigAnnotation] == "foobar.com" &&
						cluster.Annotations[addontypes.ProxyCACertConfigAnnotation] == "dummyCertificate" &&
						cluster.Annotations[addontypes.IPFamilyConfigAnnotation] == "ipv4" {
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

					fooPackage := clusterBootstrap.Spec.AdditionalPackages[1]
					Expect(fooPackage.RefName == foobarCarvelPackageName).To(BeTrue())
					Expect(*fooPackage.ValuesFrom.ProviderRef.APIGroup == "run.tanzu.vmware.com").To(BeTrue())
					Expect(fooPackage.ValuesFrom.ProviderRef.Kind == "FooBar").To(BeTrue())

					providerName := fmt.Sprintf("%s-%s-package", clusterName, foobarCarvelPackageRefName)
					Expect(fooPackage.ValuesFrom.ProviderRef.Name == providerName).To(BeTrue())

					gvr = schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha1", Resource: "foobars"}
					var err error
					object, err = dynamicClient.Resource(gvr).Namespace(clusterNamespace).Get(ctx, providerName, metav1.GetOptions{})

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

				By("verifying the cloned secret of an additional package which with a secretRef")
				// Verify that secret object mapping to foobar1 package exists with ownerReferences to cluster and ClusterBootstrap
				Eventually(func() bool {
					// "foobar1.example.com" is the carvel package ref name
					foobar1SecretName := fmt.Sprintf("%s-%s-package", cluster.Name, foobar1CarvelPackageRefName)
					foobar1Secret := &corev1.Secret{}
					if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: foobar1SecretName}, foobar1Secret); err != nil {
						return false
					}
					var foundClusterOwnerRef bool
					var foundClusterBootstrapOwnerRef bool
					var foundLabels bool
					var foundType bool
					for _, ownerRef := range foobar1Secret.GetOwnerReferences() {
						if ownerRef.UID == cluster.UID {
							foundClusterOwnerRef = true
						}
						if ownerRef.UID == clusterBootstrap.UID {
							foundClusterBootstrapOwnerRef = true
						}
					}
					secretLabels := foobar1Secret.GetLabels()
					if secretLabels[addontypes.ClusterNameLabel] == clusterName &&
						secretLabels[addontypes.PackageNameLabel] == util.ParseStringForLabel(foobar1CarvelPackageName) {
						// "foobar1.example.com.1.17.2" is clusterBootstrap.additionalPackages[0].refName
						foundLabels = true
					}
					if foobar1Secret.Type == constants.ClusterBootstrapManagedSecret {
						foundType = true
					}
					if foundClusterOwnerRef && foundClusterBootstrapOwnerRef && foundLabels && foundType {
						return true
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

				// Update the secret created before for foobar1
				// Verify that the updated secret leads to an updated data-values secret
				By("Updating the foobar1 package secret", func() {
					// In test setup, foobar1 is an additional package with a secretRef as valuesFrom. The referenced secret
					// is cloned into cluster namespace on mgmt cluster, and eventually get mirrored on workload cluster
					// under tkg-system namespace.
					// Any updates to the data value secret under cluster namespace should be eventually reflected on the
					// workload cluster under tkg-system namespace.
					s := &corev1.Secret{}
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: fmt.Sprintf("%s-%s-package", cluster.Name, foobar1CarvelPackageRefName)}, s)).To(Succeed())
					s.Data["values.yaml"] = []byte("foobar1-updated")
					Expect(k8sClient.Update(ctx, s)).To(Succeed())

					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobar1CarvelPackageRefName)}, s); err != nil {
							return false
						}
						if string(s.Data["values.yaml"]) != "foobar1-updated" {
							return false
						}

						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

				// Simulate a controller adding secretRef to provider status and
				// verify that a data-values secret has been created for the Foobar package
				By("patching foobar provider object's status resource with a secret ref", func() {
					s := &corev1.Secret{}
					s.Name = util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)
					s.Namespace = clusterNamespace
					s.Data = map[string][]byte{}
					s.Data["values.yaml"] = []byte("foobar")
					Expect(k8sClient.Create(ctx, s)).To(Succeed())

					Expect(unstructured.SetNestedField(object.Object, s.Name, "status", "secretRef")).To(Succeed())

					_, err := dynamicClient.Resource(gvr).Namespace(clusterNamespace).UpdateStatus(ctx, object, metav1.UpdateOptions{})
					Expect(err).ToNot(HaveOccurred())

					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)}, s); err != nil {
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
							Namespace: clusterNamespace,
							Name:      util.GeneratePackageInstallName(clusterName, "kapp-controller.tanzu.vmware.com"),
						}, kappControllerPkgi); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				Expect(len(kappControllerPkgi.OwnerReferences) == 1).To(BeTrue())
				Expect(kappControllerPkgi.OwnerReferences[0].APIVersion).To(Equal(clusterapiv1beta1.GroupVersion.String()))
				Expect(kappControllerPkgi.OwnerReferences[0].Name).To(Equal(cluster.Name))
				Expect(kappControllerPkgi.Annotations).ShouldNot(BeNil())
				Expect(kappControllerPkgi.Annotations[addontypes.ClusterNamespaceAnnotation]).Should(Equal(clusterNamespace))
				Expect(kappControllerPkgi.Annotations[addontypes.ClusterNameAnnotation]).Should(Equal(clusterName))

				remoteClient, err := util.GetClusterClient(ctx, k8sClient, scheme, clusterapiutil.ObjectKey(cluster))
				Expect(err).NotTo(HaveOccurred())
				Expect(remoteClient).NotTo(BeNil())

				By("verifying that ServiceAccount, ClusterRole and ClusterRoleBinding are created on the workload cluster properly")
				sa := &corev1.ServiceAccount{}
				Eventually(func() bool {
					if err := remoteClient.Get(ctx,
						client.ObjectKey{Namespace: constants.TKGSystemNS, Name: constants.PackageInstallServiceAccount},
						sa); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				clusterRole := &rbacv1.ClusterRole{}
				Eventually(func() bool {
					if err := remoteClient.Get(ctx,
						client.ObjectKey{Name: constants.PackageInstallClusterRole},
						clusterRole); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
				Eventually(func() bool {
					if err := remoteClient.Get(ctx,
						client.ObjectKey{Name: constants.PackageInstallClusterRoleBinding},
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
				Expect(remotePkgi.Annotations).ShouldNot(BeNil())
				Expect(remotePkgi.Annotations[addontypes.ClusterNamespaceAnnotation]).Should(Equal(clusterNamespace))
				Expect(remotePkgi.Annotations[addontypes.ClusterNameAnnotation]).Should(Equal(clusterName))

				// Update the secret created before for foobar resource
				// Verify that the updated secret leads to an updated data-values secret
				By("Updating the foobar package provider secret", func() {
					s := &corev1.Secret{}
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)}, s)).To(Succeed())
					s.OwnerReferences = []metav1.OwnerReference{
						{
							APIVersion: clusterapiv1beta1.GroupVersion.String(),
							Kind:       "Cluster",
							Name:       cluster.Name,
							UID:        cluster.UID,
						},
					}
					s.Data["values.yaml"] = []byte("foobar-updated")
					Expect(k8sClient.Update(ctx, s)).To(Succeed())

					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)}, s); err != nil {
							return false
						}
						if string(s.Data["values.yaml"]) != "foobar-updated" {
							return false
						}

						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

				By("verifying the embedded local object reference is cloned into cluster namespace", func() {
					name := "cpi-vsphere-credential"
					assertEventuallyExistInNamespace(ctx, k8sClient, clusterNamespace, name, &corev1.Secret{})
					assertSecretContains(ctx, k8sClient, clusterNamespace, name, map[string][]byte{
						"username": []byte("Zm9v"), // foo
						"password": []byte("YmFy"), // bar
					})
					assertOwnerReferencesExist(ctx, k8sClient, clusterNamespace, name, &corev1.Secret{}, []metav1.OwnerReference{
						{APIVersion: clusterapiv1beta1.GroupVersion.String(), Kind: "Cluster", Name: clusterName},
						{APIVersion: vspherecpiv1alpha1.GroupVersion.String(), Kind: "VSphereCPIConfig", Name: "test-cluster-cpi"},
					})
				})

				By("Updating cluster TKR version", func() {
					newTKRVersion := "v1.23.3"
					cluster := &clusterapiv1beta1.Cluster{}
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
					cluster.Labels[constants.TKRLabelClassyClusters] = newTKRVersion
					Expect(k8sClient.Update(ctx, cluster)).To(Succeed())

					// Wait for ClusterBootstrap upgrade reconciliation
					Eventually(func() bool {
						upgradedClusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), upgradedClusterBootstrap)
						if err != nil || upgradedClusterBootstrap.Status.ResolvedTKR != newTKRVersion {
							return false
						}
						// Validate CNI
						cni := upgradedClusterBootstrap.Spec.CNI
						fmt.Println(cni.RefName)
						Expect(strings.HasPrefix(cni.RefName, "antrea")).To(BeTrue())
						Expect(cni.RefName).To(Equal("antrea.tanzu.vmware.com.1.2.3--vmware.4-tkg.2-advanced-zshippable"))
						Expect(*cni.ValuesFrom.ProviderRef.APIGroup).To(Equal("cni.tanzu.vmware.com"))
						Expect(cni.ValuesFrom.ProviderRef.Kind).To(Equal("AntreaConfig"))
						Expect(cni.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-antrea.tanzu.vmware.com-package", clusterName)))

						// Validate Kapp
						kapp := upgradedClusterBootstrap.Spec.Kapp
						Expect(kapp.RefName).To(Equal("kapp-controller.tanzu.vmware.com.0.30.1"))
						Expect(*kapp.ValuesFrom.ProviderRef.APIGroup).To(Equal("run.tanzu.vmware.com"))
						Expect(kapp.ValuesFrom.ProviderRef.Kind).To(Equal("KappControllerConfig"))
						Expect(kapp.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-kapp-controller.tanzu.vmware.com-package", clusterName)))

						// Validate additional packages
						// foobar3 should be added, while foobar should be kept even it was removed from the template
						Expect(len(upgradedClusterBootstrap.Spec.AdditionalPackages)).To(Equal(3))
						for _, pkg := range upgradedClusterBootstrap.Spec.AdditionalPackages {
							if pkg.RefName == "foobar1.example.com.1.18.2" {
								Expect(pkg.ValuesFrom.SecretRef).To(Equal(fmt.Sprintf("%s-foobar1.example.com-package", clusterName)))
							} else if pkg.RefName == "foobar3.example.com.1.17.2" {
								Expect(*pkg.ValuesFrom.ProviderRef.APIGroup).To(Equal("run.tanzu.vmware.com"))
								Expect(pkg.ValuesFrom.ProviderRef.Kind).To(Equal("FooBar"))
								Expect(pkg.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-foobar3.example.com-package", clusterName)))
							} else if pkg.RefName == "foobar.example.com.1.17.2" {
								Expect(*pkg.ValuesFrom.ProviderRef.APIGroup).To(Equal("run.tanzu.vmware.com"))
								Expect(pkg.ValuesFrom.ProviderRef.Kind).To(Equal("FooBar"))
								Expect(pkg.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-foobar.example.com-package", clusterName)))
							} else {
								return false
							}
						}

						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

			})
		})
	})

	When("ClusterBootstrap is paused", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-tcbt-2"
			clusterNamespace = "cluster-namespace-2"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-2.yaml"
		})
		Context("from a ClusterBootstrapTemplate", func() {
			It("should block ClusterBootstrap reconciliation if it is paused", func() {

				By("verifying CAPI cluster is created properly")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				copiedCluster := cluster.DeepCopy()
				copiedCluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, copiedCluster)).To(Succeed())

				By("ClusterBootstrap CR is created")
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

				By("Should have remote secret value same as foobar1secret secret value")
				remoteClient, err := util.GetClusterClient(ctx, k8sClient, scheme, clusterapiutil.ObjectKey(cluster))
				Expect(err).NotTo(HaveOccurred())
				Expect(remoteClient).NotTo(BeNil())
				s := &corev1.Secret{}
				remoteSecret := &corev1.Secret{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: fmt.Sprintf("%s-%s-package", clusterName, foobar1CarvelPackageRefName)}, s)
					if err != nil {
						return false
					}
					err = remoteClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobar1CarvelPackageRefName)}, remoteSecret)
					if err != nil {
						return false
					}
					if string(s.Data["values.yaml"]) != string(remoteSecret.Data["values.yaml"]) {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("Pause ClusterBootstrap CR")
				clusterBootstrap.Spec.Paused = true
				Expect(k8sClient.Update(ctx, clusterBootstrap)).To(Succeed())

				By("Should not reconcile foobar1secret secret change", func() {
					// Get secretRef
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: fmt.Sprintf("%s-%s-package", clusterName, foobar1CarvelPackageRefName)}, s)).To(Succeed())
					s.Data["values.yaml"] = []byte("values changed")
					Expect(k8sClient.Update(ctx, s)).To(Succeed())

					// Wait 10 seconds in case reconciliation happens
					time.Sleep(10 * time.Second)

					Eventually(func() bool {
						remoteSecret := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobar1CarvelPackageRefName)}, remoteSecret); err != nil {
							return false
						}
						// values.yaml should not update
						Expect(string(s.Data["values.yaml"]) == string(remoteSecret.Data["values.yaml"])).ToNot(BeTrue())
						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

			})
		})
	})

})

func assertSecretContains(ctx context.Context, k8sClient client.Client, namespace, name string, secretContent map[string][]byte) {
	s := &corev1.Secret{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, s)
	Expect(err).ToNot(HaveOccurred())
	for k, v := range secretContent {
		Expect(string(s.Data[k]) == string(v)).ToNot(BeTrue())
	}
}

func assertEventuallyExistInNamespace(ctx context.Context, k8sClient client.Client, namespace, name string, obj client.Object) {
	Eventually(func() error {
		key := client.ObjectKey{Name: name, Namespace: namespace}
		return k8sClient.Get(ctx, key, obj)
	}, waitTimeout, pollingInterval).Should(Succeed())
}

func assertOwnerReferencesExist(ctx context.Context, k8sClient client.Client, namespace, name string, obj client.Object, ownerReferencesToCheck []metav1.OwnerReference) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	Expect(k8sClient.Get(ctx, key, obj)).NotTo(HaveOccurred())

	for _, ownerReferenceToCheck := range ownerReferencesToCheck {
		found := false
		for _, ownerReferenceFromObj := range obj.GetOwnerReferences() {
			// skip the comparison of UID on purpose, caller does not know the UIDs of ownerReferencesToCheck beforehand
			if ownerReferenceToCheck.APIVersion == ownerReferenceFromObj.APIVersion &&
				ownerReferenceToCheck.Kind == ownerReferenceFromObj.Kind &&
				ownerReferenceToCheck.Name == ownerReferenceFromObj.Name {
				found = true
			}
		}
		Expect(found).To(BeTrue())
	}
}
