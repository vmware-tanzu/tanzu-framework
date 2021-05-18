// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"strings"
	"time"

	"github.com/vmware-tanzu-private/core/addons/testutil"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/cluster-api/util/secret"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
	"github.com/vmware-tanzu-private/core/addons/testutil"

	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
)

const (
	waitTimeout             = time.Second * 90
	pollingInterval         = time.Second * 2
	appSyncPeriod           = 5 * time.Minute
	appWaitTimeout          = 30 * time.Second
	addonNamespace          = "tkg-system"
	addonServiceAccount     = "tkg-addons-app-sa"
	addonClusterRole        = "tkg-addons-app-cluster-role"
	addonClusterRoleBinding = "tkg-addons-app-cluster-role-binding"
	addonImagePullPolicy    = "IfNotPresent"
	corePackageRepoName     = "core"
)

var _ = Describe("Addon Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster, tkr, BOM config map and addon secret")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).To(Succeed())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, "default", k8sClient)).To(Succeed())
	})

	AfterEach(func() {
		By("Deleting cluster, tkr, BOM config map and addon secret")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.DeleteResources(f, cfg, dynamicClient, true)).To(Succeed())

		By("Deleting Addon data-values secrets")
		addonSecretKey := client.ObjectKey{
			Namespace: addonNamespace,
			Name:      "antrea-data-values",
		}
		dataValuesSecret := &v1.Secret{}
		Expect(k8sClient.Get(ctx, addonSecretKey, dataValuesSecret)).To(Succeed())
		Expect(k8sClient.Delete(ctx, dataValuesSecret)).To(Succeed())

		By("Deleting Addon app CR")
		appKey := client.ObjectKey{
			Namespace: addonNamespace,
			Name:      "antrea",
		}
		antreaApp := &kappctrl.App{}
		// some testcases don't create App CR
		k8sClient.Get(ctx, appKey, antreaApp) // nolint:errcheck
		k8sClient.Delete(ctx, antreaApp)      // nolint:errcheck

		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: "default",
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &v1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	Context("reconcileAddonNormal for a tkr 1.18.1", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-1"
			clusterResourceFilePath = "testdata/test-cluster-1.yaml"
		})

		It("Should create addon namespace, service account cluster admin service role and role binding", func() {

			Eventually(func() bool {
				ns := &v1.NamespaceList{}
				err := k8sClient.List(ctx, ns)
				if err != nil {
					return false
				}
				for _, n := range ns.Items {
					if n.Name == addonNamespace {
						return true
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      addonServiceAccount,
				}
				svc := &v1.ServiceAccount{}
				err := k8sClient.Get(ctx, key, svc)
				return err == nil
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				roles := &rbacv1.ClusterRoleList{}
				err := k8sClient.List(ctx, roles)
				if err != nil {
					return false
				}
				for _, r := range roles.Items {
					if r.Name == addonClusterRole {
						rule := r.Rules[0]
						if rule.APIGroups[0] == "*" && rule.Verbs[0] == "*" && rule.Resources[0] == "*" {
							return true
						}
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				roleBindings := &rbacv1.ClusterRoleBindingList{}
				err := k8sClient.List(ctx, roleBindings)
				if err != nil {
					return false
				}
				for _, r := range roleBindings.Items {
					if r.Name == addonClusterRoleBinding &&
						r.RoleRef.Name == addonClusterRole {
						if r.Subjects[0].Name == addonServiceAccount &&
							r.Subjects[0].Namespace == addonNamespace {
							return true
						}

					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

		It("Should create addon resources", func() {

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      "antrea-data-values",
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, key, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "serviceCidr: 100.64.0.0/13")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      "antrea",
				}
				app := &kappctrl.App{}
				Expect(k8sClient.Get(ctx, key, app)).To(Succeed())

				Expect(app.Annotations[addontypes.AddonTypeAnnotation]).Should(Equal("cni/antrea"))
				Expect(app.Annotations[addontypes.AddonNameAnnotation]).Should(Equal("test-cluster-1-antrea"))
				// TODO why is this needed
				Expect(app.Annotations[addontypes.AddonNamespaceAnnotation]).Should(Equal("default"))

				Expect(app.Spec.ServiceAccountName).Should(Equal(addonServiceAccount))

				Expect(app.Spec.Fetch[0].Image.URL).Should(Equal("projects-stg.registry.vmware.com/tkg/addons/antrea-templates:98adbf4"))

				appTmplYtt := kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									Name: "antrea-data-values",
								},
							},
						},
					},
				}

				Expect(*app.Spec.Template[0].Ytt).Should(Equal(appTmplYtt))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})
	})

	Context("reconcileAddonNormal for a tkr 1.20.5", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-2"
			clusterResourceFilePath = "testdata/test-cluster-2.yaml"
		})

		It("Should create addon resources", func() {

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      "antrea-data-values",
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, key, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(secretData).Should(Equal("serviceCidr: 100.64.0.0/13\n"))
				imageInfoData := string(secret.Data["imageInfo.yaml"])
				Expect(strings.Contains(imageInfoData, "imageRepository: projects.registry.vmware.com/tkg")).Should(BeTrue())
				Expect(strings.Contains(imageInfoData, "imagePath: antrea/antrea-debian")).Should(BeTrue())
				Expect(strings.Contains(imageInfoData, "tag: v0.11.3_vmware.2")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      "antrea",
				}
				app := &kappctrl.App{}
				Expect(k8sClient.Get(ctx, key, app)).To(Succeed())

				Expect(app.Annotations[addontypes.AddonTypeAnnotation]).Should(Equal("cni/antrea"))
				Expect(app.Annotations[addontypes.AddonNameAnnotation]).Should(Equal("test-cluster-2-antrea"))
				// TODO why is this needed
				Expect(app.Annotations[addontypes.AddonNamespaceAnnotation]).Should(Equal("default"))

				Expect(app.Spec.ServiceAccountName).Should(Equal(addonServiceAccount))

				Expect(app.Spec.Fetch[0].Image.URL).Should(Equal("projects.registry.vmware.com/tkg/tanzu_core/addons/antrea-templates:v1.3.1"))

				appTmplYtt := kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									Name: "antrea-data-values",
								},
							},
						},
					},
				}

				Expect(*app.Spec.Template[0].Ytt).Should(Equal(appTmplYtt))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

	})

	Context("reconcileAddonNormal for a tkr 1.20.6", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-3"
			clusterResourceFilePath = "testdata/test-cluster-3.yaml"
		})

		It("Should create addon resources", func() {

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      "antrea-data-values",
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, key, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(secretData).Should(Equal("serviceCidr: 100.64.0.0/13\n"))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Name:      corePackageRepoName,
					Namespace: addonNamespace,
				}
				pkgr := &pkgiv1alpha1.PackageRepository{}
				Expect(k8sClient.Get(ctx, key, pkgr)).To(Succeed())

				pkgrSpec := pkgiv1alpha1.PackageRepositorySpec{
					Fetch: &pkgiv1alpha1.PackageRepositoryFetch{
						ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{
							Image: "projects.registry.vmware.com/tkg/tanzu_core_repo/core-package-repository:v1.4.0+vmware.0",
						},
					},
				}

				Expect(pkgr.Spec).Should(Equal(pkgrSpec))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonNamespace,
					Name:      "antrea",
				}
				ipkg := &pkgiv1alpha1.PackageInstall{}
				Expect(k8sClient.Get(ctx, key, ipkg)).To(Succeed())

				Expect(ipkg.Annotations[addontypes.AddonTypeAnnotation]).Should(Equal("cni/antrea"))
				Expect(ipkg.Annotations[addontypes.AddonNameAnnotation]).Should(Equal("test-cluster-3-antrea"))
				// TODO why is this needed
				Expect(ipkg.Annotations[addontypes.AddonNamespaceAnnotation]).Should(Equal("default"))

				Expect(ipkg.Spec.ServiceAccountName).Should(Equal(addonServiceAccount))

				Expect(ipkg.Spec.PackageRef).ShouldNot(BeNil())
				Expect(ipkg.Spec.PackageRef.RefName).Should(Equal("antrea.vmware.com"))
				Expect(ipkg.Spec.PackageRef.VersionSelection.Prereleases).ShouldNot(Equal(nil))

				ipkgValues := []pkgiv1alpha1.PackageInstallValues{
					{
						SecretRef: &pkgiv1alpha1.PackageInstallValuesSecretRef{
							Name: "antrea-data-values",
						},
					},
				}

				Expect(ipkg.Spec.Values).Should(Equal(ipkgValues))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: "default",
					Name:      "test-cluster-3-kapp-controller",
				}
				app := &kappctrl.App{}
				Expect(k8sClient.Get(ctx, key, app)).To(Succeed())

				Expect(app.Annotations[addontypes.AddonTypeAnnotation]).Should(Equal("addons-management/kapp-controller"))
				Expect(app.Annotations[addontypes.AddonNameAnnotation]).Should(Equal("test-cluster-3-kapp-controller"))
				// TODO why is this needed
				Expect(app.Annotations[addontypes.AddonNamespaceAnnotation]).Should(Equal("default"))

				appCluster := &kappctrl.AppCluster{
					KubeconfigSecretRef: &kappctrl.AppClusterKubeconfigSecretRef{
						Name: "test-cluster-3-kubeconfig",
						Key:  "value",
					},
				}
				Expect(app.Spec.Cluster).Should(Equal(appCluster))

				Expect(app.Spec.Fetch[0].ImgpkgBundle.Image).Should(Equal("projects.registry.vmware.com/tkg/tanzu_core/addons/kapp-controller-package:v1.4.0+vmware.1"))

				appTmplYtt := kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Paths:                 []string{"config"},
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									Name: "test-cluster-3-kapp-controller-data-values",
								},
							},
						},
					},
				}

				Expect(*app.Spec.Template[0].Ytt).Should(Equal(appTmplYtt))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

	})

})
