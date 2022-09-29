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
	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

const (
	bootstrapObject = "../fakes/config/clusterbootstrap.yaml"
	packageNotFound = "package not found"
)

//nolint:goimports
var (
	fakeMgtClusterClient *fakes.ClusterClient
	fakeWcClusterClient  *fakes.ClusterClient
	timeout              time.Duration
	clusterBootstrap     *runtanzuv1alpha3.ClusterBootstrap
	pkgKapp              *kapppkgv1alpha1.Package
	pkgCni               *kapppkgv1alpha1.Package
	pkgCsi               *kapppkgv1alpha1.Package
	pkgCpi               *kapppkgv1alpha1.Package
)

const pkgKappObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
spec:
    refname: kapp-controller.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`
const pkgCniObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
spec:
    refname: antrea.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`
const pkgCsiObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
spec:
    refname: vsphere-pv-csi.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`

const pkgCpiObjStr = `apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
spec:
    refname: vsphere-cpi.tanzu.vmware.com
    version: 0.38.4+vmware.1-tkg.2-zshippable`

func init() {
	timeout = time.Duration(1)
	bs, _ := os.ReadFile(bootstrapObject)
	clusterBootstrap = &runtanzuv1alpha3.ClusterBootstrap{}
	//Expect(yaml.Unmarshal(bs, clusterBootstrap)).To(Succeed(), "Failed to convert the cluster bootstrap input file to yaml")
	yaml.Unmarshal(bs, clusterBootstrap) //nolint:errcheck
	pkgKapp = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkgKappObjStr), pkgKapp) //nolint:errcheck

	pkgCni = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkgCniObjStr), pkgCni) //nolint:errcheck

	pkgCsi = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkgCsiObjStr), pkgCsi) //nolint:errcheck

	pkgCpi = &kapppkgv1alpha1.Package{}
	yaml.Unmarshal([]byte(pkgCpiObjStr), pkgCpi) //nolint:errcheck
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
				fakeMgtClusterClient.GetPackageReturnsOnCall(0, pkgKapp, err)
			})
			It("should return error", func() {
				_, err := GetCorePackagesFromClusterBootstrap(fakeMgtClusterClient, fakeWcClusterClient, clusterBootstrap, constants.CorePackagesNamespaceInTKGM, clusterBootstrap.Name)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errorStr))
			})
		})

		When("package installation not successful because of MC packages error, should return error", func() {
			packageNotFound := packageNotFound
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
			packageNotFound := packageNotFound
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
	fakeMgtClusterClient.GetPackageReturnsOnCall(0, pkgKapp, nil)
	fakeWcClusterClient.WaitForPackageInstallReturns(nil)
	fakeWcClusterClient.GetPackageReturnsOnCall(0, pkgCni, nil)
	fakeWcClusterClient.GetPackageReturnsOnCall(1, pkgCsi, nil)
	fakeWcClusterClient.GetPackageReturnsOnCall(2, pkgCpi, nil)
}
