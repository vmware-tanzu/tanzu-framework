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
	const (
		antreaPkg           = "antrea.tanzu.vmware.com.1.2.3--vmware.1-tkg.2"
		antreaPkgRefName    = "antrea.tanzu.vmware.com"
		clusterNameMng      = "test-mng"
		clusterNamespaceMng = "tkg-system"
		clusterNameWlc      = "test-wlc"
		clusterNamespaceWlc = "test-ns-wlc"
		kappPkgRefName      = "kapp-controller.tanzu.vmware.com"
	)

	var (
		clusterResourceFilePath string
		mngAntreaObjKey         = client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GeneratePackageInstallName(clusterNameMng, antreaPkgRefName)}
		wlcAntreaObjKey         = client.ObjectKey{Namespace: constants.TKGSystemNS, Name: util.GeneratePackageInstallName(clusterNameWlc, antreaPkgRefName)}
		wlcKappObjKey           = client.ObjectKey{Namespace: clusterNamespaceWlc, Name: util.GeneratePackageInstallName(clusterNameWlc, kappPkgRefName)}
	)

	JustBeforeEach(func() {
		By("Creating resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ShouldNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).Should(Succeed())

		By("Installing antrea into management cluster")
		installPackage(clusterNameMng, antreaPkg, constants.TKGSystemNS)
		updatePkgInstallStatus(mngAntreaObjKey, kappctrlv1alpha1.ReconcileSucceeded)

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
			mngClusterCopy := mngCluster.DeepCopy()
			mngClusterCopy.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
			Expect(k8sClient.Status().Update(ctx, mngClusterCopy)).Should(Succeed())

			By("verifying workload cluster is created properly")
			wlcCluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespaceWlc, Name: clusterNameWlc}, wlcCluster)).Should(Succeed())
			// update workload cluster status phase to 'Provisioned' as it is required for ClusterBootstrap reconciler
			wlcClusterCopy := wlcCluster.DeepCopy()
			wlcClusterCopy.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
			Expect(k8sClient.Status().Update(ctx, wlcClusterCopy)).Should(Succeed())

			By("verifying management and workload clusters' ClusterBootstrap has been created")
			clusterBootstrapGet(client.ObjectKeyFromObject(mngCluster))
			clusterBootstrapGet(client.ObjectKeyFromObject(wlcCluster))

			By("verifying un-managed packages do not update the 'Status.Conditions' for ClusterBootstrap")
			// install unmanaged package into management cluster. Make sure ClusterBootstrap conditions does not get changed from un-managed package
			installPackage(mngCluster.Name, "pkg.test.carvel.dev.1.0.0", mngCluster.Namespace)
			mngClusterBootstrap := clusterBootstrapGet(client.ObjectKeyFromObject(mngCluster))
			// Antrea is already installed into management cluster's tkg-system namespace
			Expect(len(mngClusterBootstrap.Status.Conditions)).Should(Equal(1))
			antreaCondType := "Antrea-" + runtanzuv1alpha3.ConditionType(v1alpha1.ReconcileSucceeded)
			Expect(mngClusterBootstrap.Status.Conditions[0].Type).Should(Equal(antreaCondType))
			// install unmanaged package into workload cluster. Make sure cluster bootstrap conditions does not get changed fro un-managed package
			installPackage(wlcCluster.Name, "pkg.test.carvel.dev.1.0.0", wlcCluster.Namespace)
			wlcClusterBootstrap := clusterBootstrapGet(client.ObjectKeyFromObject(wlcCluster))
			Expect(len(wlcClusterBootstrap.Status.Conditions)).Should(Equal(0))

			By("verifying ClusterBootstrap 'Status.Conditions' does not get updated when PackageInstall's summarized condition is Unknown")
			// verify for management cluster
			updatePkgInstallStatus(mngAntreaObjKey, "")
			mngClusterBootstrap = clusterBootstrapGet(client.ObjectKeyFromObject(mngCluster))
			// Antrea is already installed into management cluster's tkg-system namespace
			Expect(len(mngClusterBootstrap.Status.Conditions)).Should(Equal(1))
			Expect(mngClusterBootstrap.Status.Conditions[0].Type).Should(Equal(antreaCondType))
			// verify for workload cluster
			updatePkgInstallStatus(wlcAntreaObjKey, "")
			updatePkgInstallStatus(wlcKappObjKey, "")
			wlcClusterBootstrap = clusterBootstrapGet(client.ObjectKeyFromObject(wlcCluster))
			Expect(len(wlcClusterBootstrap.Status.Conditions)).Should(Equal(0))

			By("verifying ClusterBootstrap 'Status.Conditions' gets updated for managed packages")
			// verify for management cluster
			mngClusterBootstrapStatus := waitForClusterBootstrapStatus(client.ObjectKeyFromObject(mngCluster), antreaCondType)
			Expect(len(mngClusterBootstrapStatus.Conditions)).Should(Equal(1))
			Expect(mngClusterBootstrapStatus.Conditions[0].Type).Should(Equal(antreaCondType))
			// verify for workload cluster
			updatePkgInstallStatus(wlcAntreaObjKey, kappctrlv1alpha1.ReconcileSucceeded)
			updatePkgInstallStatus(wlcKappObjKey, kappctrlv1alpha1.Reconciling)
			antreaCondType = "Antrea-" + runtanzuv1alpha3.ConditionType(v1alpha1.ReconcileSucceeded)
			kappCondType := "Kapp-Controller-" + runtanzuv1alpha3.ConditionType(v1alpha1.Reconciling)
			waitForClusterBootstrapStatus(client.ObjectKeyFromObject(wlcCluster), antreaCondType)
			wlcClusterBootstrapStatus := waitForClusterBootstrapStatus(client.ObjectKeyFromObject(wlcCluster), kappCondType)
			Expect(len(wlcClusterBootstrapStatus.Conditions)).Should(Equal(2))
			Expect(wlcClusterBootstrapStatus.Conditions[0].Type).Should(Equal(antreaCondType))
			Expect(wlcClusterBootstrapStatus.Conditions[1].Type).Should(Equal(kappCondType))
		})
	})
})

// updatePkgInstallStatus simulates kapp controller PackageInstall status update
func updatePkgInstallStatus(objKey client.ObjectKey, appCondType kappctrlv1alpha1.AppConditionType) {
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

// clusterBootstrapGet gets ClusterBootstrap resource with the provided object key
func clusterBootstrapGet(objKey client.ObjectKey) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, clusterBootstrap)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(clusterBootstrap).ShouldNot(BeNil())
	return clusterBootstrap
}

// waitForClusterBootstrapStatus checks ClusterBootstrap's 'Status.Conditions' includes provided condition type
func waitForClusterBootstrapStatus(objKey client.ObjectKey, condType runtanzuv1alpha3.ConditionType) *runtanzuv1alpha3.ClusterBootstrapStatus {
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

// installPackage installs a package into the provided namespace
func installPackage(clusterName, pkgName, namespace string) {
	packageRefName, _, err := util.GetPackageMetadata(ctx, k8sClient, pkgName, namespace)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(packageRefName).ShouldNot(Equal(""))

	pkg := &kapppkgv1alpha1.Package{}
	key := client.ObjectKey{Namespace: namespace, Name: pkgName}
	Expect(k8sClient.Get(ctx, key, pkg)).Should(Succeed())

	pkgiName := util.GeneratePackageInstallName(clusterName, packageRefName)
	pkgi := &kapppkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pkgiName,
			Namespace: namespace,
			Labels:    map[string]string{types.ClusterNameLabel: clusterName, types.ClusterNamespaceLabel: namespace},
		},
		Spec: kapppkgiv1alpha1.PackageInstallSpec{
			PackageRef: &kapppkgiv1alpha1.PackageRef{
				RefName:          pkgName,
				VersionSelection: &versions.VersionSelectionSemver{Constraints: "1.0.0"},
			},
		},
	}

	_, err = controllerutil.CreateOrPatch(ctx, k8sClient, pkgi, nil)
	Expect(err).ShouldNot(HaveOccurred())

	installedPkg := &kapppkgiv1alpha1.PackageInstall{}
	key = client.ObjectKey{Namespace: namespace, Name: pkgiName}
	Expect(k8sClient.Get(ctx, key, installedPkg)).Should(Succeed())
}
