// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	addonutil "github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

const (
	getResourceTimeout  = time.Minute * 1
	waitForReadyTimeout = time.Minute * 10
	pollingInterval     = time.Second * 30
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

func checkClusterCB(ctx context.Context, mccl, wccl client.Client, mcClusterName, wcClusterName, infrastructureName string) error {
	log.Infof("Verify addons on workload cluster %s with management cluster %s", wcClusterName, mcClusterName)

	// get ClusterBootstrap and return error if not found
	clusterBootstrap := getClusterBootstrap(ctx, mccl, constants.TkgNamespace, mcClusterName)

	// verify cni package is installed on the workload cluster
	cniPkgShortName, cniPkgName, cniPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.CNI.RefName)
	verifyPackageInstall(ctx, wccl, addonutil.GeneratePackageInstallName(wcClusterName, cniPkgShortName), cniPkgName, cniPkgVersion)

	// verify the remote kapp-controller package is installed on the management cluster
	kappPkgShortName, kappPkgName, kappPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.Kapp.RefName)
	verifyPackageInstall(ctx, mccl, addonutil.GeneratePackageInstallName(wcClusterName, kappPkgShortName), kappPkgName, kappPkgVersion)

	// verify csi and cpi package is installed on the workload cluster if in vSphere environment
	if infrastructureName == "vsphere" {
		csiPkgShortName, csiPkgName, csiPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.CSI.RefName)
		verifyPackageInstall(ctx, wccl, addonutil.GeneratePackageInstallName(wcClusterName, csiPkgShortName), csiPkgName, csiPkgVersion)

		cpiPkgShortName, cpiPkgName, cpiPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.CPI.RefName)
		verifyPackageInstall(ctx, wccl, addonutil.GeneratePackageInstallName(wcClusterName, cpiPkgShortName), cpiPkgName, cpiPkgVersion)
	}

	// loop over additional packages list in clusterBootstrap spec to check package information
	if clusterBootstrap.Spec.AdditionalPackages != nil {
		// validate additional packages
		for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
			pkgShortName, pkgName, pkgVersion := getPackageDetailsFromCBS(additionalPkg.RefName)

			// TODO: temporarily skip verifying tkg-storageclass due to install failure issue.
			//		 this should be removed once the issue is resolved.
			if pkgShortName == "tkg-storageclass" {
				continue
			}
			verifyPackageInstall(ctx, wccl, addonutil.GeneratePackageInstallName(wcClusterName, pkgShortName), pkgName, pkgVersion)
		}
	}

	return nil
}

func verifyPackageInstall(ctx context.Context, c client.Client, pkgInstallName, pkgName, pkgVersion string) {
	log.Infof("Check PackageInstall %s for package %s of version %s", pkgInstallName, pkgName, pkgVersion)

	// verify the package is successfully deployed and its name and version match with the clusterBootstrap CR
	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	objKey := client.ObjectKey{Namespace: constants.TkgNamespace, Name: pkgInstallName}
	Eventually(func() bool {
		if err := c.Get(ctx, objKey, pkgInstall); err != nil {
			return false
		}
		if len(pkgInstall.Status.GenericStatus.Conditions) == 0 {
			return false
		}
		if pkgInstall.Status.GenericStatus.Conditions[0].Type != kappctrl.ReconcileSucceeded {
			return false
		}
		if pkgInstall.Status.GenericStatus.Conditions[0].Status != corev1.ConditionTrue {
			return false
		}
		if pkgInstall.Spec.PackageRef.RefName != pkgName {
			return false
		}
		if pkgInstall.Spec.PackageRef.VersionSelection.Constraints != pkgVersion {
			return false
		}
		return true
	}, waitForReadyTimeout, pollingInterval).Should(BeTrue(), "Package %s is not deployed successfully", pkgName)
}

func getPackageDetailsFromCBS(CBSRefName string) (pkgShortName, pkgName, pkgVersion string) {
	pkgShortName = strings.Split(CBSRefName, ".")[0]
	pkgName = strings.Join(strings.Split(CBSRefName, ".")[0:4], ".")
	pkgVersion = strings.Join(strings.Split(CBSRefName, ".")[4:], ".")
	return
}

func getClusterBootstrap(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterName}

	Eventually(func() error {
		return k8sClient.Get(ctx, objKey, clusterBootstrap)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(clusterBootstrap).ShouldNot(BeNil())
	return clusterBootstrap
}
