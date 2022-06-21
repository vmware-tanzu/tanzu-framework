// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/tools/clientcmd"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	addonutil "github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	waitTimeout     = time.Second * 90
	pollingInterval = time.Second * 2
)

// create cluster client from kubeconfig
func createClientFromKubeconfig(exportFile string, scheme *runtime.Scheme) (client.Client, error) {
	config, err := clientcmd.LoadFromFile(exportFile)
	Expect(err).ToNot(HaveOccurred(), "Failed to load cluster Kubeconfig file from %q", exportFile)

	rawConfig, err := clientcmd.Write(*config)
	Expect(err).ToNot(HaveOccurred(), "Failed to create raw config ")

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(rawConfig)
	Expect(err).ToNot(HaveOccurred(), "Failed to create rest config ")

	client, err := client.New(restConfig, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred(), "Failed to create a cluster client")

	return client, nil
}

// getClusterBootstrap gets ClusterBootstrap resource with the provided object key
func getClusterBootstrap(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterName}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, clusterBootstrap)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(clusterBootstrap).ShouldNot(BeNil())
	return clusterBootstrap
}

// getPackageInstall get PackageInstall resource with the provided object key
func getPackageInstall(ctx context.Context, k8sClient client.Client, namespace, pkgiName string) *kapppkgiv1alpha1.PackageInstall {
	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	objKey := client.ObjectKey{Namespace: namespace, Name: pkgiName}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, pkgInstall)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(pkgInstall).ShouldNot(BeNil())

	return pkgInstall
}

func getPackageDetailsFromCBS(CBSRefName string) (string, string, string, error) {
	pkgShortName := strings.Split(CBSRefName, ".")[0]

	pkgName := strings.Join(strings.Split(CBSRefName, ".")[0:4], ".")

	pkgVersion := strings.Join(strings.Split(CBSRefName, ".")[4:], ".")

	return pkgShortName, pkgName, pkgVersion, nil
}

func checkClusterCBS(ctx context.Context, mccl, wccl client.Client, mcClusterName, wcClusterName string) error {
	var err error

	// Get ClusterBootstrap and return error if not found
	clusterBootstrap := getClusterBootstrap(ctx, mccl, constants.TkgNamespace, mcClusterName)

	// packageInstall name for for both management and workload clusters should follow the <cluster name>-<addon short name>
	// packageInstall name and version should match info in clusterBootstrap for all packages, format is <package name>.<package version>
	// get package details from clusterBootstrap
	cniPkgShortName, cniPkgName, cniPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.CNI.RefName)
	Expect(err).NotTo(HaveOccurred())
	csiPkgShortName, csiPkgName, csiPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.CSI.RefName)
	Expect(err).NotTo(HaveOccurred())
	cpiPkgShortName, cpiPkgName, cpiPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.CPI.RefName)
	Expect(err).NotTo(HaveOccurred())
	kappPkgShortName, kappPkgName, kappPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.Kapp.RefName)
	Expect(err).NotTo(HaveOccurred())
	msPkgShortName, msPkgName, msPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[0].RefName)
	Expect(err).NotTo(HaveOccurred())
	scPkgShortName, scPkgName, scPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[1].RefName)
	Expect(err).NotTo(HaveOccurred())
	pdPkgShortName, pdPkgName, pdPkgVersion, err := getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[1].RefName)
	Expect(err).NotTo(HaveOccurred())

	// verify clusterBootstrap details matches package install on workload cluster
	verifyPackageInstall(ctx, wccl, wcClusterName, cniPkgShortName, cniPkgName, cniPkgVersion)
	verifyPackageInstall(ctx, wccl, wcClusterName, csiPkgShortName, csiPkgName, csiPkgVersion)
	verifyPackageInstall(ctx, wccl, wcClusterName, cpiPkgShortName, cpiPkgName, cpiPkgVersion)
	verifyPackageInstall(ctx, wccl, wcClusterName, kappPkgShortName, kappPkgName, kappPkgVersion)
	verifyPackageInstall(ctx, wccl, wcClusterName, msPkgShortName, msPkgName, msPkgVersion)
	verifyPackageInstall(ctx, wccl, wcClusterName, scPkgShortName, scPkgName, scPkgVersion)
	verifyPackageInstall(ctx, wccl, wcClusterName, pdPkgShortName, pdPkgName, pdPkgVersion)

	return nil
}

func verifyPackageInstall(ctx context.Context, wccl client.Client, clusterName, pkgShortName, pkgName, pkgVersion string) {
	pkgiName := addonutil.GeneratePackageInstallName(clusterName, pkgShortName)
	pkgi := getPackageInstall(ctx, wccl, constants.TkgNamespace, pkgiName)

	// check package install reconcile status is succeed
	Expect(pkgi.Status.GenericStatus.Conditions[0].Type).Should(Equal(kappctrl.ReconcileSucceeded))
	Expect(pkgi.Status.GenericStatus.Conditions[0].Status).Should(Equal(corev1.ConditionTrue))

	// Verify package name match between clusterBootstrap and packageInstall
	Expect(pkgName).Should(Equal(pkgi.Spec.PackageRef.RefName))

	// Verify package version match between clusterBootstrap and packageInstall
	Expect(pkgVersion).Should(Equal(pkgi.Spec.PackageRef.VersionSelection.Constraints))
}
