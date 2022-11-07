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

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	antreaconfigv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
	vspherecpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
	vspherecsiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/builder"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
)

var _ = Describe("ClusterBootstrap Reconciler", func() {
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
		foobar2CarvelPackageRefName = "foobar2.example.com"
		foobar                      = "foobar"
		foobarUpdated               = "foobar-updated"
		foobar1Updated              = "foobar1-updated"
	)

	JustBeforeEach(func() {
		// set up the certificates and webhook before creating any objects
		By("Creating and installing new certificates for ClusterBootstrap Webhooks")
		err := testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
		Expect(err).ToNot(HaveOccurred())

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

		// delete cluster
		By("Deleting cluster")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.DeleteResources(f, cfg, dynamicClient, false)).To(Succeed())
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
				cluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

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

				By("cluster should be marked with finalizer")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), cluster)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(cluster, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("clusterbootstrap should be marked with finalizer")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(clusterBootstrap, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("cluster kubeconfig secret should be marked with finalizer")
				clusterKubeConfigSecret := &corev1.Secret{}
				key := client.ObjectKey{Namespace: cluster.Namespace, Name: secret.Name(clusterName, secret.Kubeconfig)}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, key, clusterKubeConfigSecret)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(clusterKubeConfigSecret, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("packageinstall should have been created for each additional package in the clusterBoostrap")
				Expect(hasPackageInstalls(ctx, k8sClient, cluster, constants.TKGSystemNS,
					clusterBootstrap.Spec.AdditionalPackages, setupLog)).To(BeTrue())

				By("packageinstalls for core packages should not have owner references")
				var corePackages []*runtanzuv1alpha3.ClusterBootstrapPackage
				corePackages = append(corePackages, clusterBootstrap.Spec.CNI, clusterBootstrap.Spec.CPI, clusterBootstrap.Spec.CSI)
				pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
				for _, pkg := range corePackages {
					pkgInstallName := util.GeneratePackageInstallName(cluster.Name, pkg.RefName)
					err := k8sClient.Get(ctx, client.ObjectKey{Name: pkgInstallName, Namespace: constants.TKGSystemNS}, pkgInstall)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(pkgInstall.OwnerReferences) == 0).To(BeTrue())

				}

				By("packageinstalls for additional packages should not have owner references")
				pkgInstall = &kapppkgiv1alpha1.PackageInstall{}
				for _, pkg := range clusterBootstrap.Spec.AdditionalPackages {
					pkgInstallName := util.GeneratePackageInstallName(cluster.Name, pkg.RefName)
					err := k8sClient.Get(ctx, client.ObjectKey{Name: pkgInstallName, Namespace: constants.TKGSystemNS}, pkgInstall)
					// Since Foobar provider does not have a controller to create the data values secret and update to the status, the foobar packageInstall
					// should not be created so far.
					if pkg.RefName == foobarCarvelPackageName {
						Expect(err).Should(HaveOccurred())
						Expect(strings.Contains(err.Error(), "packageinstalls.packaging.carvel.dev \"test-cluster-tcbt-foobar\" not found")).To(BeTrue())
					} else {
						Expect(err).ToNot(HaveOccurred())
						Expect(len(pkgInstall.OwnerReferences) == 0).To(BeTrue())
					}
				}

				By("verifying that CNI has been populated properly")
				// Verify CNI is populated in the cloned object with the value from the cluster bootstrap template
				Expect(clusterBootstrap.Spec.CNI).NotTo(BeNil())
				Expect(strings.HasPrefix(clusterBootstrap.Spec.CNI.RefName, "antrea")).To(BeTrue())

				Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal("antrea.tanzu.vmware.com.1.2.3--vmware.1-tkg.1"))
				Expect(*clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.APIGroup).To(Equal("cni.tanzu.vmware.com"))
				Expect(clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Kind).To(Equal("AntreaConfig"))
				providerName := fmt.Sprintf("%s-antrea-package", clusterName)
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
						cluster.Annotations[addontypes.NoProxyConfigAnnotation] == "foobar.com,google.com" &&
						cluster.Annotations[addontypes.ProxyCACertConfigAnnotation] == "aGVsbG8=\nbHWtcH9=\n" &&
						cluster.Annotations[addontypes.IPFamilyConfigAnnotation] == "ipv4" &&
						cluster.Annotations[addontypes.SkipTLSVerifyConfigAnnotation] == "registry1, registry2" {
						return true
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("verifying that the providerRef from additionalPackages is cloned into cluster namespace and ownerReferences set properly")
				var gvr schema.GroupVersionResource
				var object *unstructured.Unstructured
				// Verify providerRef exists and also the cloned provider object with ownerReferences to cluster and ClusterBootstrap
				Eventually(func() bool {
					Expect(len(clusterBootstrap.Spec.AdditionalPackages) > 1).To(BeTrue())

					fooPackage := clusterBootstrap.Spec.AdditionalPackages[1]
					Expect(fooPackage.RefName == foobarCarvelPackageName).To(BeTrue())
					Expect(*fooPackage.ValuesFrom.ProviderRef.APIGroup == "run.tanzu.vmware.com").To(BeTrue())
					Expect(fooPackage.ValuesFrom.ProviderRef.Kind == "FooBar").To(BeTrue())

					providerName := fmt.Sprintf("%s-foobar-package", clusterName)
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
					foobar1SecretName := fmt.Sprintf("%s-foobar1-package", cluster.Name)
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
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: fmt.Sprintf("%s-foobar1-package", cluster.Name)}, s)).To(Succeed())
					s.StringData = make(map[string]string)
					s.StringData["values.yaml"] = foobar1Updated
					Expect(k8sClient.Update(ctx, s)).To(Succeed())
					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobar1CarvelPackageRefName)}, s); err != nil {
							return false
						}
						if string(s.Data["values.yaml"]) != foobar1Updated {
							return false
						}
						// TKGS data value should not exist because no VirtualMachine is related to cluster1
						_, ok := s.Data[constants.TKGSDataValueFileName]
						Expect(ok).ToNot(BeTrue())
						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

				// Simulate a controller adding secretRef to provider status and
				// verify that a data-values secret has been created for the Foobar package
				By("patching foobar provider object's status resource with a secret ref", func() {
					s := &corev1.Secret{}
					s.Name = util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)
					s.Namespace = clusterNamespace
					s.StringData = map[string]string{}
					s.StringData["values.yaml"] = string(foobar)
					Expect(k8sClient.Create(ctx, s)).To(Succeed())

					Expect(unstructured.SetNestedField(object.Object, s.Name, "status", "secretRef")).To(Succeed())

					_, err := dynamicClient.Resource(gvr).Namespace(clusterNamespace).UpdateStatus(ctx, object, metav1.UpdateOptions{})
					Expect(err).ToNot(HaveOccurred())

					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)}, s); err != nil {
							return false
						}
						if string(s.Data["values.yaml"]) != foobar {
							return false
						}

						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

				By("verifying that data value secret is created for a package with inline config")
				Eventually(func() bool {
					s := &corev1.Secret{}
					if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobar2CarvelPackageRefName)}, s); err != nil {
						return false
					}
					dataValue := string(s.Data["values.yaml"])

					if !strings.Contains(dataValue, "key1") || !strings.Contains(dataValue, "sample-value1") ||
						!strings.Contains(dataValue, "key2") || !strings.Contains(dataValue, "sample-value2") {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())

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
						if clusterBootstrapPackage.ValuesFrom != nil {
							Expect(remoteSecret.Data).To(HaveKey("values.yaml"))
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
				Expect(remotePkgi.Spec.SyncPeriod.Seconds()).To(Equal(constants.PackageInstallSyncPeriod.Seconds()))
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
					s.StringData = make(map[string]string)
					s.StringData["values.yaml"] = foobarUpdated
					Expect(k8sClient.Update(ctx, s)).To(Succeed())

					Eventually(func() bool {
						s := &corev1.Secret{}
						if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobarCarvelPackageRefName)}, s); err != nil {
							return false
						}
						if string(s.Data["values.yaml"]) != foobarUpdated {
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

				By("verifying the embedded local csi object reference is cloned into cluster namespace", func() {
					name := "csi-vsphere-credential"
					assertEventuallyExistInNamespace(ctx, k8sClient, clusterNamespace, name, &corev1.Secret{})
					assertSecretContains(ctx, k8sClient, clusterNamespace, name, map[string][]byte{
						"username": []byte("Zm9v"), // foo
						"password": []byte("YmFy"), // bar
					})
					assertOwnerReferencesExist(ctx, k8sClient, clusterNamespace, name, &corev1.Secret{}, []metav1.OwnerReference{
						{APIVersion: clusterapiv1beta1.GroupVersion.String(), Kind: "Cluster", Name: clusterName},
						{APIVersion: vspherecsiv1alpha1.GroupVersion.String(), Kind: "VSphereCSIConfig", Name: "test-cluster-csi"},
					})
				})

				By("Updating cluster TKR version", func() {
					newTKRVersion := "v1.23.3"
					cluster = &clusterapiv1beta1.Cluster{}
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
					cluster.Labels[constants.TKRLabelClassyClusters] = newTKRVersion

					// Mock cluster pause mutating webhook
					cluster.Spec.Paused = true
					if cluster.Annotations == nil {
						cluster.Annotations = map[string]string{}
					}
					cluster.Annotations[constants.ClusterPauseLabel] = newTKRVersion
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
						Expect(strings.HasPrefix(cni.RefName, "antrea")).To(BeTrue())
						// Note: The value of CNI has been bumped to the one in TKR after cluster upgrade
						Expect(cni.RefName).To(Equal("antrea.tanzu.vmware.com.1.5.2--vmware.3-tkg.1-advanced-zshippable"))
						Expect(*cni.ValuesFrom.ProviderRef.APIGroup).To(Equal("cni.tanzu.vmware.com"))
						Expect(cni.ValuesFrom.ProviderRef.Kind).To(Equal("AntreaConfig"))
						Expect(cni.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-antrea-package", clusterName)))

						// Validate Kapp
						kapp := upgradedClusterBootstrap.Spec.Kapp
						Expect(kapp.RefName).To(Equal("kapp-controller.tanzu.vmware.com.0.30.2"))
						Expect(*kapp.ValuesFrom.ProviderRef.APIGroup).To(Equal("run.tanzu.vmware.com"))
						Expect(kapp.ValuesFrom.ProviderRef.Kind).To(Equal("KappControllerConfig"))
						Expect(kapp.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-kapp-controller-package", clusterName)))

						// Validate additional packages
						// foobar3 should be added, while foobar should be kept even it was removed from the template
						Expect(len(upgradedClusterBootstrap.Spec.AdditionalPackages)).To(Equal(4))
						for _, pkg := range upgradedClusterBootstrap.Spec.AdditionalPackages {
							if pkg.RefName == "foobar1.example.com.1.18.2" {
								Expect(pkg.ValuesFrom.SecretRef).To(Equal(fmt.Sprintf("%s-foobar1-package", clusterName)))
							} else if pkg.RefName == "foobar3.example.com.1.17.2" {
								Expect(*pkg.ValuesFrom.ProviderRef.APIGroup).To(Equal("run.tanzu.vmware.com"))
								Expect(pkg.ValuesFrom.ProviderRef.Kind).To(Equal("FooBar"))
								Expect(pkg.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-foobar3-package", clusterName)))
							} else if pkg.RefName == "foobar.example.com.1.17.2" {
								Expect(*pkg.ValuesFrom.ProviderRef.APIGroup).To(Equal("run.tanzu.vmware.com"))
								Expect(pkg.ValuesFrom.ProviderRef.Kind).To(Equal("FooBar"))
								Expect(pkg.ValuesFrom.ProviderRef.Name).To(Equal(fmt.Sprintf("%s-foobar-package", clusterName)))
							} else if pkg.RefName == "foobar2.example.com.1.18.2" {
								Expect(pkg.ValuesFrom.Inline).NotTo(BeNil())
							} else {
								return false
							}
						}

						// The cluster should be unpaused
						Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
						Expect(cluster.Spec.Paused).ToNot(BeTrue())
						if cluster.Annotations != nil {
							_, ok := cluster.Annotations[constants.ClusterPauseLabel]
							Expect(ok).ToNot(BeTrue())
						}

						return true
					}, waitTimeout, pollingInterval).Should(BeTrue())
				})

				By("Test ClusterBootstrap webhook validateUpdate ", func() {
					// fetch the latest clusterbootstrap
					Eventually(func() bool {
						err = k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
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

					// CNI can't be nil
					mutateClusterBootstrap := clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.CNI = nil
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(err).To(HaveOccurred())
					Expect(strings.Contains(err.Error(), "spec.cni: Invalid value: \"null\": package can't be nil")).To(BeTrue())

					// Kapp can't be nil
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.Kapp = nil
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(err).To(HaveOccurred())
					Expect(strings.Contains(err.Error(), "spec.kapp: Invalid value: \"null\": package can't be nil")).To(BeTrue())

					// CSI can be nil
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.CSI = nil
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(err).ToNot(HaveOccurred())

					err = k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
					Expect(err).ToNot(HaveOccurred())

					// Package CR must exist
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.CNI.RefName = "antrea.tanzu.vmware.com.1.2.5--vmware.1-tkg.1"
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(strings.Contains(err.Error(), "\"antrea.tanzu.vmware.com.1.2.5--vmware.1-tkg.1\" not found")).To(BeTrue())

					// Package Refname can't change
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.CNI.RefName = "foobar1.example.com.1.17.2"
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(strings.Contains(err.Error(), "new package refName and old package refName should be the same")).To(BeTrue())

					// ProviderRef can't be changed to secretRef or inline
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.Kapp.ValuesFrom.ProviderRef = nil
					mutateClusterBootstrap.Spec.Kapp.ValuesFrom.Inline = map[string]interface{}{}
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(strings.Contains(err.Error(), "change from providerRef to other types of data value representation is not allowed")).To(BeTrue())

					// Package can't be downgraded
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.CNI.RefName = "antrea.tanzu.vmware.com.0.13.3--vmware.1-tkg.1"
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(strings.Contains(err.Error(), "package downgrade is not allowed")).To(BeTrue())

					// Additional package can't be removed
					mutateClusterBootstrap = clusterBootstrap.DeepCopy()
					mutateClusterBootstrap.Spec.AdditionalPackages = mutateClusterBootstrap.Spec.AdditionalPackages[:len(mutateClusterBootstrap.Spec.AdditionalPackages)-1]
					err = k8sClient.Update(ctx, mutateClusterBootstrap)
					Expect(strings.Contains(err.Error(), "missing updated additional package")).To(BeTrue())
				})

				By("Test ClusterBootstrap webhook validateCreate", func() {
					namespace := "default"

					// case1
					in := builder.ClusterBootstrap(namespace, "test-cb-1").
						WithCNIPackage(builder.ClusterBootstrapPackage("cni.example.com.1.17.2").WithProviderRef("run.tanzu.vmware.com", "foo", "bar").Build()).
						WithAdditionalPackage(builder.ClusterBootstrapPackage("pinniped.example.com.1.11.3").Build()).Build()
					err = k8sClient.Create(ctx, in)
					Expect(err).Should(HaveOccurred())
					Expect(strings.Contains(err.Error(), "packages.data.packaging.carvel.dev \"cni.example.com.1.17.2\" not found")).To(BeTrue())
					Expect(strings.Contains(err.Error(), "package can't be nil")).To(BeTrue())

					// case2
					in = builder.ClusterBootstrap(addonNamespace, "test-cb-1").
						WithKappPackage(builder.ClusterBootstrapPackage("kapp-controller.tanzu.vmware.com.0.30.2").WithProviderRef("run.tanzu.vmware.com", "foo", "bar").Build()).
						WithCNIPackage(builder.ClusterBootstrapPackage("calico.tanzu.vmware.com.3.19.1--vmware.1-tkg.1").WithProviderRef("cni.tanzu.vmware.com", "CalicoConfig", "invalidName").Build()).
						WithAdditionalPackage(builder.ClusterBootstrapPackage("foobar.example.com.1.17.2").WithSecretRef("invalidSecret").Build()).Build()
					err = k8sClient.Create(ctx, in)
					Expect(err).Should(HaveOccurred())
					Expect(strings.Contains(err.Error(), "calicoconfigs.cni.tanzu.vmware.com \"invalidName\" not found")).To(BeTrue())
					Expect(strings.Contains(err.Error(), "unable to find server preferred resource run.tanzu.vmware.com/foo")).To(BeTrue())

				})

				By("Test ClusterBootstrapTemplate webhook validateUpdate", func() {
					// fetch the latest clusterbootstraptemplate
					clusterBootstrapTemplate := &runtanzuv1alpha3.ClusterBootstrapTemplate{}
					key = client.ObjectKey{Namespace: addonNamespace, Name: "v1.22.3"}
					err = k8sClient.Get(ctx, key, clusterBootstrapTemplate)
					Expect(err).ToNot(HaveOccurred())

					// ClusterBootstrapTemplate spec should be immutable
					clusterBootstrapTemplate.Spec.Kapp = nil
					err = k8sClient.Update(ctx, clusterBootstrapTemplate)
					Expect(err).Should(HaveOccurred())
					Expect(strings.Contains(err.Error(), "ClusterBootstrapTemplate has immutable spec, update is not allowed")).To(BeTrue())
				})
				By("finalizers should be added back automatically to bootstrap")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(clusterBootstrap, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("delete cluster with foreground propagation policy")
				deletePropagation := metav1.DeletePropagationForeground
				deleteOptions := client.DeleteOptions{PropagationPolicy: &deletePropagation}
				Expect(k8sClient.Delete(ctx, cluster, &deleteOptions)).To(Succeed())

				By("instacllpackages for additional packages should have been removed.")
				Expect(hasPackageInstalls(ctx, k8sClient, cluster, constants.TKGSystemNS,
					clusterBootstrap.Spec.AdditionalPackages, setupLog)).To(BeTrue())
				Eventually(func() bool {
					return hasPackageInstalls(ctx, k8sClient, cluster, constants.TKGSystemNS,
						clusterBootstrap.Spec.AdditionalPackages, setupLog)
				}, waitTimeout, pollingInterval).Should(BeFalse())

				By("finalizer should be removed from clusterboostrap")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(clusterBootstrap, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeFalse())

				By("finalizer should be removed from cluster")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), cluster)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(cluster, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeFalse())

				By("finalizer should be removed from cluster kubeconfig secret")
				key = client.ObjectKey{Namespace: cluster.Namespace, Name: secret.Name(clusterName, secret.Kubeconfig)}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, key, clusterKubeConfigSecret)
					if err != nil {
						return false
					}
					return controllerutil.ContainsFinalizer(clusterKubeConfigSecret, addontypes.AddonFinalizer)
				}, waitTimeout, pollingInterval).Should(BeFalse())
			})
		})
	})

	When("cluster is created from clusterBootstrapTemplate", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-tcbt-2"
			clusterNamespace = "cluster-namespace-2"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-2.yaml"
		})
		Context("from a ClusterBootstrapTemplate", func() {
			It("should perform ClusterBootstrap reconciliation & block reconciliation if ClusterBootstrap gets paused", func() {

				By("verifying CAPI cluster is created properly")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				cluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

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

				By("should create secret for packages with empty valuesFrom & corresponding data values secret in workload cluster.", func() {
					remoteClient, err := util.GetClusterClient(ctx, k8sClient, scheme, clusterapiutil.ObjectKey(cluster))
					Expect(err).NotTo(HaveOccurred())
					Expect(remoteClient).NotTo(BeNil())

					clusterBootstrapPackages := []*runtanzuv1alpha3.ClusterBootstrapPackage{
						clusterBootstrap.Spec.CNI,
						clusterBootstrap.Spec.CPI,
						clusterBootstrap.Spec.CSI,
						clusterBootstrap.Spec.Kapp,
					}
					clusterBootstrapPackages = append(clusterBootstrapPackages, clusterBootstrap.Spec.AdditionalPackages...)
					for _, pkg := range clusterBootstrapPackages {
						if pkg == nil || pkg.ValuesFrom != nil {
							continue
						}

						localSecret := &corev1.Secret{}
						Eventually(func() bool {
							if err := k8sClient.Get(ctx,
								client.ObjectKey{
									Namespace: clusterNamespace,
									Name:      util.GeneratePackageSecretName(cluster.Name, pkg.RefName),
								}, localSecret); err != nil {
								return false
							}
							return true
						}, waitTimeout, pollingInterval).Should(BeTrue())

						remoteSecret := &corev1.Secret{}
						Eventually(func() bool {
							if err := remoteClient.Get(ctx,
								client.ObjectKey{
									Namespace: constants.TKGSystemNS,
									Name:      util.GenerateDataValueSecretName(cluster.Name, foobar2CarvelPackageRefName),
								}, remoteSecret); err != nil {
								return false
							}
							return true
						}, waitTimeout, pollingInterval).Should(BeTrue())

						Expect(remoteSecret.Data).To(HaveKey(constants.TKGSDataValueFileName))
					}
				})

				By("Should have remote secret value same as foobar1secret secret value and should have created TKGS data values")
				remoteClient, err := util.GetClusterClient(ctx, k8sClient, scheme, clusterapiutil.ObjectKey(cluster))
				Expect(err).NotTo(HaveOccurred())
				Expect(remoteClient).NotTo(BeNil())
				s := &corev1.Secret{}
				remoteSecret := &corev1.Secret{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: fmt.Sprintf("%s-foobar1-package", clusterName)}, s)
					if err != nil {
						return false
					}
					Expect(s.Data).To(HaveKey("values.yaml"))
					Expect(s.Data).NotTo(HaveKey(constants.TKGSDataValueFileName))
					err = remoteClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GenerateDataValueSecretName(clusterName, foobar1CarvelPackageRefName)}, remoteSecret)
					if err != nil {
						return false
					}
					if string(s.Data["values.yaml"]) != string(remoteSecret.Data["values.yaml"]) {
						return false
					}
					Expect(remoteSecret.Data).To(HaveKey("values.yaml"))
					Expect(remoteSecret.Data).To(HaveKey(constants.TKGSDataValueFileName))
					// TKGS data values should be added because cluster has related VirtualMachine
					valueTexts, ok := remoteSecret.Data[constants.TKGSDataValueFileName]
					if !ok {
						return false
					}
					fmt.Println(string(valueTexts))
					Expect(strings.Contains(string(valueTexts), "nodeSelector:\n    run.tanzu.vmware.com/tkr: v1.22.4")).To(BeTrue())
					Expect(strings.Contains(string(valueTexts), "deployment:\n    updateStrategy: RollingUpdate")).To(BeTrue())
					Expect(strings.Contains(string(valueTexts), "daemonset:\n    updateStrategy: OnDelete")).To(BeTrue())
					Expect(strings.Contains(string(valueTexts), "maxUnavailable: 0")).To(BeTrue())
					Expect(strings.Contains(string(valueTexts), "maxSurge: 1")).To(BeTrue())
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())

				By("Pause ClusterBootstrap CR")
				clusterBootstrap.Spec.Paused = true
				Expect(k8sClient.Update(ctx, clusterBootstrap)).To(Succeed())

				// should block ClusterBootstrap reconciliation if it is paused
				By("Should not reconcile foobar1secret secret change", func() {
					// Get secretRef
					Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: fmt.Sprintf("%s-foobar1-package", clusterName)}, s)).To(Succeed())
					s.StringData = make(map[string]string)
					s.StringData["values.yaml"] = "values changed"
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

	When("Cluster is created", func() {

		var routableAntreaConfig *antreaconfigv1alpha1.AntreaConfig

		BeforeEach(func() {
			clusterName = "test-cluster-tcbt-3"
			clusterNamespace = "cluster-namespace-3"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-3.yaml"
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil {
				Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
			}

			routableAntreaConfig = &antreaconfigv1alpha1.AntreaConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-antrea-package", clusterName),
					Namespace: clusterNamespace,
				},
				Spec: antreaconfigv1alpha1.AntreaConfigSpec{
					Antrea: antreaconfigv1alpha1.Antrea{
						AntreaConfigDataValue: antreaconfigv1alpha1.AntreaConfigDataValue{
							TrafficEncapMode: "hybrid",
							NoSNAT:           true,
							FeatureGates: antreaconfigv1alpha1.AntreaFeatureGates{
								AntreaProxy:   true,
								EndpointSlice: true,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, routableAntreaConfig)).NotTo(HaveOccurred())
			assertEventuallyExistInNamespace(ctx, k8sClient, clusterNamespace, routableAntreaConfig.Name, &antreaconfigv1alpha1.AntreaConfig{})
		})

		Context("with a routable AntreaConfig resource already exist", func() {
			It("clusterbootstrap_controller should not overwrite the existing AntreaConfig Specs", func() {

				// get clusterBootstrap object for the cluster and inspect that its CNI matches the AntreaConfig that was pre-created.
				clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, clusterBootstrap)
					return err == nil
				}, waitTimeout, pollingInterval).Should(BeTrue())

				antreaConfig := &antreaconfigv1alpha1.AntreaConfig{}
				// use the name from cloned clusterBootstrap to verify it is the same
				assertOwnerReferencesExist(ctx, k8sClient, clusterNamespace, clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Name, antreaConfig, []metav1.OwnerReference{
					{
						APIVersion: clusterapiv1beta1.GroupVersion.String(),
						Kind:       "Cluster",
						Name:       clusterName,
					},
				})
				Expect(antreaConfig.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode).To(
					Equal(routableAntreaConfig.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode))
				Expect(antreaConfig.Spec.Antrea.AntreaConfigDataValue.NoSNAT).To(
					Equal(routableAntreaConfig.Spec.Antrea.AntreaConfigDataValue.NoSNAT))
				Expect(antreaConfig.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaProxy).To(
					Equal(routableAntreaConfig.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaProxy))
				Expect(antreaConfig.Spec.Antrea.AntreaConfigDataValue.FeatureGates.EndpointSlice).To(
					Equal(routableAntreaConfig.Spec.Antrea.AntreaConfigDataValue.FeatureGates.EndpointSlice))
			})
		})
	})

	When("Legacy cluster is created", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-legacy"
			clusterNamespace = "legacy-namespace"
			clusterResourceFilePath = "testdata/test-cluster-legacy.yaml"
		})
		Context("and clusterboostrap template does not exists", func() {
			It("clusterbootstrap controller should not attempt to reconcile it", func() {
				By("verifying CAPI cluster is created properly")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				cluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())
			})
		})
	})

	When("Cluster with no valuesFrom for kapp-controller", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-4"
			clusterNamespace = "cluster-namespace-4"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-4.yaml"
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil {
				Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
			}
		})
		Context("controller should not crash", func() {
			It("and create package install for kapp ", func() {
				By("setting cluster phase to provisioned")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				cluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

				pkgiName := util.GeneratePackageInstallName(clusterName, "kapp-controller.tanzu.vmware.com.0.31.0")
				pkgi := &kapppkgiv1alpha1.PackageInstall{}
				Eventually(func() bool {
					if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: pkgiName}, pkgi); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
			})
		})
	})

	// This test case is for ensuring that controller is watching providerRef for core packages.
	// 1. create antrea with foobar cr as provider, packageinstall gets generated with a given secret
	// 2. replace secret in .status.secretRef of foobar cr provider
	When("Cluster with external provider for CNI", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-5"
			clusterNamespace = "cluster-namespace-5"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-5.yaml"
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())
		})
		Context("When providerRef exists", func() {
			It("Should create package install for antrea with package datavalues for secret and "+
				"update package install for antrea with renamed secret as controller is watching for changes to providerRef", func() {

				By("setting cluster phase to provisioned")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				cluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

				Eventually(func(g Gomega) {
					providerName := fmt.Sprintf("%s-antrea-package", clusterName)
					gvr := schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha1", Resource: "foobars"}
					provider, err := dynamicClient.Resource(gvr).Namespace(clusterNamespace).Get(ctx, providerName, metav1.GetOptions{})
					g.Expect(err).ToNot(HaveOccurred())

					s := &corev1.Secret{}
					s.Name = util.GenerateDataValueSecretName(clusterName, "antrea.tanzu.vmware.com.1.10.5--vmware.1-tkg.2")
					s.Namespace = clusterNamespace
					s.StringData = map[string]string{}
					s.StringData["values.yaml"] = foobar
					err = k8sClient.Create(ctx, s)

					if err != nil {
						g.Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
					}
					g.Expect(unstructured.SetNestedField(provider.Object, s.Name, "status", "secretRef")).To(Succeed())

					_, err = dynamicClient.Resource(gvr).Namespace(clusterNamespace).UpdateStatus(ctx, provider, metav1.UpdateOptions{})
					g.Expect(err).ToNot(HaveOccurred())

					pkgiName := util.GeneratePackageInstallName(clusterName, "antrea.tanzu.vmware.com.1.10.5--vmware.1-tkg.2")
					pkgi := &kapppkgiv1alpha1.PackageInstall{}
					g.Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: pkgiName}, pkgi)).To(Succeed())

					g.Expect(len(pkgi.Spec.Values) > 0).To(BeTrue())

					dataValueSecret := pkgi.Spec.Values[0].SecretRef.Name
					dataValuesGenerated := &corev1.Secret{}

					g.Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: dataValueSecret}, dataValuesGenerated)).To(Succeed())
					g.Expect(string(dataValuesGenerated.Data["values.yaml"])).To(Equal(foobar))
				}, waitTimeout, pollingInterval).Should(Succeed())

				Eventually(func(g Gomega) {
					providerName := fmt.Sprintf("%s-antrea-package", clusterName)
					gvr := schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha1", Resource: "foobars"}
					provider, err := dynamicClient.Resource(gvr).Namespace(clusterNamespace).Get(ctx, providerName, metav1.GetOptions{})
					g.Expect(err).ToNot(HaveOccurred())

					// update the secret in provider
					s := &corev1.Secret{}
					s.Name = "updated-antrea-secret"
					s.Namespace = clusterNamespace
					s.StringData = map[string]string{}
					s.StringData["values.yaml"] = "foobarbaz"
					err = k8sClient.Create(ctx, s)

					if err != nil {
						g.Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
					}

					g.Expect(unstructured.SetNestedField(provider.Object, s.Name, "status", "secretRef")).To(Succeed())

					_, err = dynamicClient.Resource(gvr).Namespace(clusterNamespace).UpdateStatus(ctx, provider, metav1.UpdateOptions{})
					g.Expect(err).ToNot(HaveOccurred())

					pkgiName := util.GeneratePackageInstallName(clusterName, "antrea.tanzu.vmware.com.1.10.5--vmware.1-tkg.2")
					pkgi := &kapppkgiv1alpha1.PackageInstall{}
					g.Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: pkgiName}, pkgi)).To(Succeed())

					g.Expect(len(pkgi.Spec.Values) > 0).To(BeTrue())

					dataValueSecret := pkgi.Spec.Values[0].SecretRef.Name
					dataValuesGenerated := &corev1.Secret{}
					g.Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: constants.TKGSystemNS, Name: dataValueSecret}, dataValuesGenerated)).To(Succeed())

					g.Expect(string(dataValuesGenerated.Data["values.yaml"])).To(Equal("foobarbaz"))

				}, waitTimeout, pollingInterval).Should(Succeed())

			})
		})
	})

	When("cluster is created with custom clusterbootstrap annotation present", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-6"
			clusterNamespace = "cluster-namespace-6"
			clusterResourceFilePath = "testdata/test-cluster-bootstrap-6.yaml"
		})
		Context("custom ClusterBootstrap being used", func() {
			It("should not create ClusterBootstrap CR by cloning from the template", func() {

				By("verifying cluster has the annotation and CB is not cloned from template")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				Expect(cluster.Annotations).NotTo(BeNil())
				_, ok := cluster.Annotations[constants.CustomClusterBootstrap]
				Expect(ok).To(BeTrue())

				clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
				Expect(err).Should(HaveOccurred())
				Expect(strings.Contains(err.Error(), fmt.Sprintf("clusterbootstraps.run.tanzu.vmware.com \"%s\" not found", clusterName))).To(BeTrue())
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
	Eventually(func() bool {
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
			if !found {
				return false
			}
		}
		return true
	}, waitTimeout, pollingInterval).Should(BeTrue())
}
