// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("PackageInstallStatus Reconciler", func() {
	var (
		clusterResourceFilePath string
		clusterNameMng          = "test-mng"
		clusterNamespaceMng     = "tkg-system"
		clusterNameWlc          = "test-wlc"
		clusterNamespaceWlc     = "test-ns-wlc"
		antreaPkgName           = "antrea.tanzu.vmware.com"
		kappPkgName             = "kapp-controller.tanzu.vmware.com"
	)

	JustBeforeEach(func() {
		By("Creating resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ShouldNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).Should(Succeed())

		By("Creating kubeconfig for management cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterNameMng, clusterNamespaceMng, k8sClient)).Should(Succeed())

		By("Creating kubeconfig for workload cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterNameWlc, clusterNamespaceWlc, k8sClient)).Should(Succeed())
	})

	AfterEach(func() {
		By("Deleting resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ShouldNot(HaveOccurred())
		defer f.Close()
		// Best effort resource deletion
		_ = testutil.DeleteResources(f, cfg, dynamicClient, true)
	})

	Context("reconcile PackageInstallStatus", func() {
		BeforeEach(func() {
			clusterResourceFilePath = "testdata/test-packageinstallstatus.yaml"
		})

		It("Should reconcile PackageInstallStatus on management & workload clusters", func() {

			By("verifying management cluster is created properly")
			mngCluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespaceMng, Name: clusterNameMng}, mngCluster)).Should(Succeed())
			// update management cluster status phase to 'Provisioned' as it is required for ClusterBootstrap reconciler
			mngCluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
			Expect(k8sClient.Status().Update(ctx, mngCluster)).Should(Succeed())

			By("verifying workload cluster is created properly")
			wlcCluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespaceWlc, Name: clusterNameWlc}, wlcCluster)).Should(Succeed())
			// update workload cluster status phase to 'Provisioned' as it is required for ClusterBootstrap reconciler
			wlcCluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
			Expect(k8sClient.Status().Update(ctx, wlcCluster)).Should(Succeed())

			By("verifying management and workload clusters' ClusterBootstrap has been created")
			mngClusterBootstrap := testClusterBootstrapGet(client.ObjectKeyFromObject(mngCluster))
			Expect(mngClusterBootstrap).ShouldNot(BeNil())
			wlcClusterBootstrap := testClusterBootstrapGet(client.ObjectKeyFromObject(wlcCluster))
			Expect(wlcClusterBootstrap).ShouldNot(BeNil())

			By("verifying un-managed packages do not update the 'Status.Conditions' for ClusterBootstrap")
			// install unmanaged package into management cluster. Make sure ClusterBootstrap conditions does not get changed fro un-managed package
			installUnmanagedPackage(mngCluster.Name, mngCluster.Namespace)
			mngClusterBootstrap = testClusterBootstrapGet(client.ObjectKeyFromObject(mngCluster))
			Expect(len(mngClusterBootstrap.Status.Conditions)).Should(Equal(0))
			// install unmanaged package into workload cluster. Make sure cluster bootstrap conditions does not get changed fro un-managed package
			installUnmanagedPackage(wlcCluster.Name, wlcCluster.Namespace)
			wlcClusterBootstrap = testClusterBootstrapGet(client.ObjectKeyFromObject(wlcCluster))
			Expect(len(wlcClusterBootstrap.Status.Conditions)).Should(Equal(0))

			mngAntreaObjKey := client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GeneratePackageInstallName(clusterNameMng, antreaPkgName)}
			wlcAntreaObjKey := client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GeneratePackageInstallName(clusterNameWlc, antreaPkgName)}
			wlcKappObjKey := client.ObjectKey{Namespace: clusterNamespaceWlc, Name: util.GeneratePackageInstallName(clusterNameWlc, kappPkgName)}

			By("verifying ClusterBootstrap 'Status.Conditions' does not get updated when PackageInstall's summarized condition is Unknown")
			// verify for management cluster
			testUpdatePkgInstallStatus(mngAntreaObjKey, util.UnknownCondition)
			mngClusterBootstrap = testClusterBootstrapGet(client.ObjectKeyFromObject(mngCluster))
			Expect(len(mngClusterBootstrap.Status.Conditions)).Should(Equal(0))
			// verify for workload cluster
			testUpdatePkgInstallStatus(wlcAntreaObjKey, util.UnknownCondition)
			testUpdatePkgInstallStatus(wlcKappObjKey, util.UnknownCondition)
			wlcClusterBootstrap = testClusterBootstrapGet(client.ObjectKeyFromObject(wlcCluster))
			Expect(len(wlcClusterBootstrap.Status.Conditions)).Should(Equal(0))

			By("verifying ClusterBootstrap 'Status.Conditions' gets updated for managed packages")
			// verify for management cluster
			testUpdatePkgInstallStatus(mngAntreaObjKey, kappctrlv1alpha1.Reconciling)
			antreaCondType := "Antrea" + runtanzuv1alpha3.ConditionType(v1alpha1.Reconciling)
			mngClusterBootstrapStatus := testCheckClusterBootstrapStatus(client.ObjectKeyFromObject(mngCluster), antreaCondType)
			Expect(len(mngClusterBootstrapStatus.Conditions)).Should(Equal(1))
			Expect(mngClusterBootstrapStatus.Conditions[0].Type).Should(Equal(antreaCondType))
			// verify for workload cluster
			testUpdatePkgInstallStatus(wlcAntreaObjKey, kappctrlv1alpha1.ReconcileSucceeded)
			testUpdatePkgInstallStatus(wlcKappObjKey, kappctrlv1alpha1.Reconciling)
			antreaCondType = "Antrea" + runtanzuv1alpha3.ConditionType(v1alpha1.ReconcileSucceeded)
			kappCondType := "Kapp-Controller" + runtanzuv1alpha3.ConditionType(v1alpha1.Reconciling)
			testCheckClusterBootstrapStatus(client.ObjectKeyFromObject(wlcCluster), antreaCondType)
			wlcClusterBootstrapStatus := testCheckClusterBootstrapStatus(client.ObjectKeyFromObject(wlcCluster), kappCondType)
			Expect(wlcClusterBootstrapStatus).ShouldNot(BeNil())
			Expect(len(wlcClusterBootstrapStatus.Conditions)).Should(Equal(2))
			Expect(wlcClusterBootstrapStatus.Conditions[0].Type).Should(Equal(antreaCondType))
			Expect(wlcClusterBootstrapStatus.Conditions[1].Type).Should(Equal(kappCondType))
		})
	})
})

// testUpdatePkgInstallStatus simulates kapp controller PackageInstall status update
func testUpdatePkgInstallStatus(objKey client.ObjectKey, appCondType kappctrlv1alpha1.AppConditionType) {
	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	Eventually(func() bool {
		if err := k8sClient.Get(ctx, objKey, pkgInstall); err != nil {
			return false
		}
		condition := kappctrlv1alpha1.AppCondition{Type: appCondType, Status: corev1.ConditionTrue}
		pkgInstall.Status.Conditions = append(pkgInstall.Status.Conditions, condition)
		err := k8sClient.Status().Update(ctx, pkgInstall)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())
}

// testClusterBootstrapGet gets ClusterBootstrap resource with the provided object key
func testClusterBootstrapGet(objKey client.ObjectKey) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, clusterBootstrap)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	return clusterBootstrap
}

// testCheckClusterBootstrapStatus checks ClusterBootstrap's 'Status.Conditions' includes provided condition type
func testCheckClusterBootstrapStatus(objKey client.ObjectKey, condType runtanzuv1alpha3.ConditionType) *runtanzuv1alpha3.ClusterBootstrapStatus {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	Eventually(func() bool {
		if err := k8sClient.Get(ctx, objKey, clusterBootstrap); err != nil {
			return false
		}
		for _, cond := range clusterBootstrap.Status.Conditions {
			switch {
			case cond.Type == condType:
				return true
			}
		}
		return false
	}, waitTimeout, pollingInterval).Should(BeTrue())

	return &clusterBootstrap.Status
}

// installUnmanagedPackage installs an unmanaged package into the provided namespace
func installUnmanagedPackage(clusterName, clusterNamespace string) {
	packageRefName, _, err := util.GetPackageMetadata(ctx, k8sClient, "pkg.test.carvel.dev.1.0.0", clusterNamespace)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(packageRefName).ShouldNot(Equal(""))

	unmanagedPkg := &kapppkgv1alpha1.Package{}
	key := client.ObjectKey{Namespace: clusterNamespace, Name: "pkg.test.carvel.dev.1.0.0"}
	Expect(k8sClient.Get(ctx, key, unmanagedPkg)).Should(Succeed())

	unmanagedPkgi := &kapppkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unmanaged-pkgi",
			Namespace: clusterNamespace,
			Labels:    map[string]string{types.ClusterNameLabel: clusterName, types.ClusterNamespaceLabel: clusterNamespace},
		},
		Spec: kapppkgiv1alpha1.PackageInstallSpec{
			PackageRef: &kapppkgiv1alpha1.PackageRef{
				RefName:          unmanagedPkg.Name,
				VersionSelection: &versions.VersionSelectionSemver{Constraints: "1.0.0"},
			},
		},
	}
	_, err = controllerutil.CreateOrPatch(ctx, k8sClient, unmanagedPkgi, nil)
	Expect(err).ShouldNot(HaveOccurred())

	unmanagedPkgi = &kapppkgiv1alpha1.PackageInstall{}
	key = client.ObjectKey{Namespace: clusterNamespace, Name: "unmanaged-pkgi"}
	Expect(k8sClient.Get(ctx, key, unmanagedPkgi)).Should(Succeed())
}
