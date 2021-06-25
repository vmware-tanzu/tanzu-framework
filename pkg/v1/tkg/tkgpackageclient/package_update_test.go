// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var _ = Describe("Update Package", func() {
	var (
		ctl     *pkgClient
		crtCtl  *fakes.CRTClusterClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.PackageInstalledOptions{
			PkgInstallName:  testPkgInstallName,
			Namespace:       testNamespaceName,
			Version:         "2.0.0",
			PollInterval:    testPollInterval,
			PollTimeout:     testPollTimeout,
			CreateNamespace: true,
			Install:         false,
		}
		options = opts
	)

	JustBeforeEach(func() {
		ctl = &pkgClient{kappClient: kappCtl}
		err = ctl.UpdatePackageInstall(&options)
	})

	Context("failure in getting the installed package due to GetPackageInstall API error", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(nil, errors.New("failure in GetPackageInstall"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageInstall"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the installed package", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(nil, apierrors.NewNotFound(schema.GroupResource{Resource: "PackageInstall"}, testPkgInstallName))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("package 'test-pkg' is not among the list of installed packages in namespace 'test-ns'"))
		})
		AfterEach(func() { options = opts })
	})

	Context("success in getting the installed package as empty PackageInstall was returned", func() {
		BeforeEach(func() {
			options.Install = true
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(nil, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("package-name is required when install flag is declared"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in installing the not-already-existing package due to GetPackageByName API error", func() {
		BeforeEach(func() {
			options.Install = true
			options.PackageName = testPkgName
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(nil, nil)
			kappCtl.GetPackageMetadataByNameReturns(nil, errors.New("failure in GetPackageByName"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageByName"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in updating the installed package due to GetPackageByName API error", func() {
		BeforeEach(func() {
			options.Version = testPkgVersion
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
			kappCtl.GetPackageMetadataByNameReturns(nil, errors.New("failure in GetPackageByName"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageByName"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in updating the installed package with nil version spec", func() {
		BeforeEach(func() {
			options.Version = testPkgVersion
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to update package 'test-pkg'"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in updating the installed package due to UpdatePackageInstall API error", func() {
		BeforeEach(func() {
			options.Version = testPkgVersion
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			testPkgInstall.Spec.PackageRef = &kappipkg.PackageRef{
				RefName:          testPkgInstallName,
				VersionSelection: &versions.VersionSelectionSemver{},
			}
			kappCtl.UpdatePackageInstallReturns(errors.New("failure in UpdatePackageInstall"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in UpdatePackageInstall"))
		})
		AfterEach(func() { options = opts })
	})

	Context("success in installing the not-already-existing package", func() {
		BeforeEach(func() {
			options.Install = true
			options.PackageName = testPkgName
			options.Version = testPkgVersion
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.GetPackageInstallReturns(nil, nil)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() { options = opts })
	})

	Context("success in updating the installed package", func() {
		BeforeEach(func() {
			options.Version = testPkgVersion
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			testPkgInstall.Spec.PackageRef = &kappipkg.PackageRef{
				RefName:          testPkgInstallName,
				VersionSelection: &versions.VersionSelectionSemver{},
			}
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() { options = opts })
	})
})
