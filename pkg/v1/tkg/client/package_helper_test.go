// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"sigs.k8s.io/yaml"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

const bootstrapObject = "../fakes/config/clusterbootstrap.yaml"
const kappControllerObj = "../fakes/config/kapp-controller_object.yaml"

var (
	fakeMgtClusterClient *fakes.ClusterClient
	fakeWcClusterClient  *fakes.ClusterClient
	timeout              time.Duration
	clusterBootstrap     *runtanzuv1alpha3.ClusterBootstrap
	pkg_kapp             *kapppkgv1alpha1.Package
	pkg_cni              *kapppkgv1alpha1.Package
	pkg_csi              *kapppkgv1alpha1.Package
	pkg_cpi              *kapppkgv1alpha1.Package
)

const pkg_kappObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
spec:
    refname: kapp-controller.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`
const pkg_cniObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
spec:
    refname: antrea.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`
const pkg_csiObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
spec:
    refname: vsphere-pv-csi.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`

const pkg_cpiObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
spec:
    refname: vsphere-cpi.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`

func init() {
	timeout = time.Duration(1)
	bs, _ := os.ReadFile(bootstrapObject)
	clusterBootstrap = &runtanzuv1alpha3.ClusterBootstrap{}
	//Expect(yaml.Unmarshal(bs, clusterBootstrap)).To(Succeed(), "Failed to convert the cluster bootstrap input file to yaml")
	yaml.Unmarshal(bs, clusterBootstrap)
	pkg_kapp = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkg_kappObjStr), pkg_kapp)

	pkg_cni = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkg_cniObjStr), pkg_cni)

	pkg_csi = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkg_csiObjStr), pkg_csi)

	pkg_cpi = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkg_cpiObjStr), pkg_cpi)
}

var _ = Describe("unit tests for monitor addon's packages installation", func() {
	Context("get bootstrap object for workload cluster", func() {
		When("bootstrap object exists should not return any error", func() {
			BeforeEach(func() {
				setFakeClientAndCalls()
				fakeMgtClusterClient.GetResourceReturns(nil)
			})
			It("should not return error", func() {
				_, err := GetClusterBootstrap(fakeMgtClusterClient, "cluster1", "namespace1")
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("bootstrap object not exists should return an error", func() {
			resourceNotExists := "resource not exists"
			BeforeEach(func() {
				setFakeClientAndCalls()
				fakeMgtClusterClient.GetResourceReturns(errors.New(resourceNotExists))
			})
			It("should return error", func() {
				_, err := GetClusterBootstrap(fakeMgtClusterClient, "cluster1", "namespace1")
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("get packages from bootstrap object and monitor packages installation", func() {
		When("package installation successful should not return error", func() {
			BeforeEach(func() {
				setFakeClientAndCalls()
			})
			It("should not return error", func() {
				packages, err := GetCorePackagesFromClusterBootstrap(fakeMgtClusterClient, fakeWcClusterClient, clusterBootstrap, constants.CorePackagesNamespaceInTKGM, clusterBootstrap.Name)
				Expect(err).NotTo(HaveOccurred())
				err = MonitorAddonsCorePackageInstallation(fakeMgtClusterClient, fakeWcClusterClient, packages, timeout)
				pkg, ns, _ := fakeMgtClusterClient.WaitForPackageInstallArgsForCall(0)
				Expect(pkg).To(Equal(packages[0].ObjectMeta.Name))
				Expect(ns).To(Equal(packages[0].ObjectMeta.Namespace))
				pkg, ns, _ = fakeWcClusterClient.WaitForPackageInstallArgsForCall(0)
				Expect([]string{packages[1].ObjectMeta.Name, packages[2].ObjectMeta.Name, packages[3].ObjectMeta.Name}).Should(ContainElements(pkg))
				Expect(ns).To(Equal(packages[1].ObjectMeta.Namespace))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("GetPackage returns error then GetCorePackagesFromClusterBootstrap also should return error", func() {
			errorStr := fmt.Sprintf(clusterclient.ErrUnableToGetPackage, "kapp-controller.tanzu.vmware.com.0.38.4+vmware.1-tkg.1-zshippable", clusterBootstrap.Namespace)
			BeforeEach(func() {
				setFakeClientAndCalls()
				err := errors.New(errorStr)
				fakeMgtClusterClient.GetPackageReturnsOnCall(0, pkg_kapp, err)
			})
			It("should return error", func() {
				_, err := GetCorePackagesFromClusterBootstrap(fakeMgtClusterClient, fakeWcClusterClient, clusterBootstrap, constants.CorePackagesNamespaceInTKGM, clusterBootstrap.Name)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errorStr))
			})
		})

		When("package installation not successful because of MC packages error, should return error", func() {
			packageNotFound := "package not found"
			BeforeEach(func() {
				setFakeClientAndCalls()
				fakeMgtClusterClient.WaitForPackageInstallReturns(errors.New(packageNotFound))
			})
			It("should return error", func() {
				packages, err := GetCorePackagesFromClusterBootstrap(fakeMgtClusterClient, fakeWcClusterClient, clusterBootstrap, constants.CorePackagesNamespaceInTKGM, clusterBootstrap.Name)
				Expect(err).NotTo(HaveOccurred())
				err = MonitorAddonsCorePackageInstallation(fakeMgtClusterClient, fakeWcClusterClient, packages, timeout)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(packageNotFound))
			})
		})
		When("package installation not successful because of WC packages error, should return error", func() {
			packageNotFound := "package not found"
			BeforeEach(func() {
				setFakeClientAndCalls()
				fakeWcClusterClient.WaitForPackageInstallReturns(errors.New(packageNotFound))
			})
			It("should return error", func() {
				packages, err := GetCorePackagesFromClusterBootstrap(fakeMgtClusterClient, fakeWcClusterClient, clusterBootstrap, constants.CorePackagesNamespaceInTKGM, clusterBootstrap.Name)
				Expect(err).NotTo(HaveOccurred())
				err = MonitorAddonsCorePackageInstallation(fakeMgtClusterClient, fakeWcClusterClient, packages, timeout)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(packageNotFound))
			})
		})
	})

})

func setFakeClientAndCalls() {
	fakeMgtClusterClient = &fakes.ClusterClient{}
	fakeWcClusterClient = &fakes.ClusterClient{}
	fakeMgtClusterClient.WaitForPackageInstallReturns(nil)
	fakeMgtClusterClient.GetPackageReturnsOnCall(0, pkg_kapp, nil)
	fakeWcClusterClient.WaitForPackageInstallReturns(nil)
	fakeWcClusterClient.GetPackageReturnsOnCall(0, pkg_cni, nil)
	fakeWcClusterClient.GetPackageReturnsOnCall(1, pkg_csi, nil)
	fakeWcClusterClient.GetPackageReturnsOnCall(2, pkg_cpi, nil)
}
