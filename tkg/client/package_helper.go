// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// GetClusterBootstrap returns ClusterBootstrap object for the given clusterName in the management cluster
func GetClusterBootstrap(managementClusterClient clusterclient.Client, clusterName, namespace string) (*runtanzuv1alpha3.ClusterBootstrap, error) {
	log.V(3).Infof("getting ClusterBootstrap object for cluster: %v", clusterName)
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	err := managementClusterClient.GetResource(clusterBootstrap, clusterName, namespace, nil, &clusterclient.PollOptions{Interval: clusterclient.CheckResourceInterval, Timeout: clusterclient.PackageInstallTimeout})
	return clusterBootstrap, err
}

// GetCorePackagesFromClusterBootstrap returns addon's core packages details from the given ClusterBootstrap object
func GetCorePackagesFromClusterBootstrap(regionalClusterClient clusterclient.Client, workloadClusterClient clusterclient.Client, clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, corePackagesNamespace, clusterName string) ([]kapppkgv1alpha1.Package, error) {
	var packages []kapppkgv1alpha1.Package
	// kapp package is installed in management cluster, in namespace in which workload cluster created
	if isPackageExists(clusterBootstrap.Spec.Kapp) {
		pkg, err := getPackage(regionalClusterClient, clusterBootstrap.Spec.Kapp.RefName, clusterBootstrap.Namespace, clusterName)
		if err != nil {
			return packages, err
		}
		packages = append(packages, *pkg)
	}

	// Core packages CNI, CSI and CPI (other than Kapp) installed in workload cluster
	if isPackageExists(clusterBootstrap.Spec.CNI) {
		pkg, err := getPackage(workloadClusterClient, clusterBootstrap.Spec.CNI.RefName, corePackagesNamespace, clusterName)
		if err != nil {
			return packages, err
		}
		packages = append(packages, *pkg)
	}
	if isPackageExists(clusterBootstrap.Spec.CSI) {
		pkg, err := getPackage(workloadClusterClient, clusterBootstrap.Spec.CSI.RefName, corePackagesNamespace, clusterName)
		if err != nil {
			return packages, err
		}
		packages = append(packages, *pkg)
	}
	if isPackageExists(clusterBootstrap.Spec.CPI) {
		pkg, err := getPackage(workloadClusterClient, clusterBootstrap.Spec.CPI.RefName, corePackagesNamespace, clusterName)
		if err != nil {
			return packages, err
		}
		packages = append(packages, *pkg)
	}
	return packages, nil
}

func isPackageExists(cbPackage *runtanzuv1alpha3.ClusterBootstrapPackage) bool {
	return cbPackage != nil && cbPackage.ValuesFrom != nil && cbPackage.ValuesFrom.ProviderRef != nil
}

// getPackage takes clusterclient, package name, namespace and cluster name, queries for package details and give package object (which has package name and namespace details)
func getPackage(clusterClient clusterclient.Client, packageRefName, packagesNamespace, clusterName string) (*kapppkgv1alpha1.Package, error) {
	pkg := &kapppkgv1alpha1.Package{}
	pkgFromCluster, err := clusterClient.GetPackage(packageRefName, packagesNamespace)
	if err != nil {
		return pkg, err
	}
	pkgName := GeneratePackageInstallName(clusterName, pkgFromCluster.Spec.RefName)
	pkg = &kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: pkgName, Namespace: packagesNamespace}}
	return pkg, nil
}

// GeneratePackageInstallName is the util function to generate the PackageInstall CR name in a consistent manner.
// clusterName is the name of cluster within which all resources associated with this PackageInstall CR is installed.
// It does not necessarily
// mean the PackageInstall CR will be installed in that cluster. I.e., the kapp-controller PackageInstall CR is installed
// in the management cluster but is named after "<workload-cluster-name>-kapp-controller". It indicates that this kapp-controller
// PackageInstall is for reconciling resources in a cluster named "<workload-cluster-name>".
// addonName is the short name of a Tanzu addon with which the PackageInstall CR is associated.
func GeneratePackageInstallName(clusterName, addonName string) string {
	return fmt.Sprintf("%s-%s", clusterName, strings.Split(addonName, ".")[0])
}

// MonitorAddonsCorePackageInstallation monitors addon's core packages (kapp, cni, csi and cpi) and returns error if any while monitoring packages or any packages are not installed successfully. First it monitors kapp package in management cluster then it monitors other core packages in workload cluster.
func MonitorAddonsCorePackageInstallation(regionalClusterClient clusterclient.Client, workloadClusterClient clusterclient.Client, packages []kapppkgv1alpha1.Package, packageInstallTimeout time.Duration) error {
	if len(packages) == 0 {
		return nil
	}
	var corePackagesOnWorkloadCluster, corePackagesOnManagementCluster []kapppkgv1alpha1.Package
	for _, currentPackage := range packages {
		if strings.Contains(currentPackage.ObjectMeta.Name, "kapp-controller") {
			corePackagesOnManagementCluster = append(corePackagesOnManagementCluster, currentPackage)
		} else {
			corePackagesOnWorkloadCluster = append(corePackagesOnWorkloadCluster, currentPackage)
		}
	}
	err := WaitForPackagesInstallation(regionalClusterClient, corePackagesOnManagementCluster, packageInstallTimeout)
	if err != nil {
		return err
	}
	return WaitForPackagesInstallation(workloadClusterClient, corePackagesOnWorkloadCluster, packageInstallTimeout)
}

func WaitForPackagesInstallation(clusterClient clusterclient.Client, packages []kapppkgv1alpha1.Package, packageInstallTimeout time.Duration) error {
	// Start waiting for all packages in parallel using group.Wait
	// Note: As PackageInstall resources are created in the cluster itself
	// we are using currentClusterClient which will point to correct cluster
	group, _ := errgroup.WithContext(context.Background())

	for _, currentPackage := range packages {
		pn := currentPackage.ObjectMeta.Name
		ns := currentPackage.ObjectMeta.Namespace
		log.V(3).Warningf("waiting for package: '%s'", pn)
		group.Go(
			func() error {
				err := clusterClient.WaitForPackageInstall(pn, ns, packageInstallTimeout)
				if err != nil {
					log.V(3).Warningf("failure while waiting for package: '%s' in namespace: '%s'", pn, ns)
				} else {
					log.V(3).Infof("successfully reconciled package: '%s' in namespace: '%s'", pn, ns)
				}
				return err
			})
	}

	err := group.Wait()
	if err != nil {
		return errors.Wrap(err, "failure while waiting for packages to be installed")
	}
	return nil
}
