// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var _ = Describe("List Packages", func() {
	var (
		ctl     *pkgClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.PackageListOptions{
			Available:     false,
			AllNamespaces: false,
			ListInstalled: false,
			Namespace:     testNamespaceName,
		}
		options         = opts
		pkgMetadataList *kapppkg.PackageMetadataList
		packageInstalls *kappipkg.PackageInstallList
		packageVersions *kapppkg.PackageList
		pkgInstallList  = &kappipkg.PackageInstallList{
			TypeMeta: metav1.TypeMeta{Kind: "PackageInstallList"},
			Items:    []kappipkg.PackageInstall{*testPkgInstall},
		}
		packageMetadataList = &kapppkg.PackageMetadataList{
			TypeMeta: metav1.TypeMeta{Kind: "PackageList"},
			Items: []kapppkg.PackageMetadata{{
				TypeMeta:   metav1.TypeMeta{Kind: "PackageMetadata"},
				ObjectMeta: metav1.ObjectMeta{Name: testPkgInstallName, Namespace: testNamespaceName}},
			},
		}
	)

	Context("failure in listing available packages due to ListPackageMetadata API error", func() {
		BeforeEach(func() {
			options.Available = true
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageMetadataReturns(nil, errors.New("failure in ListPackageMetadata"))
			ctl = &pkgClient{kappClient: kappCtl}
			pkgMetadataList, err = ctl.ListPackageMetadata(&options)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackageMetadata"))
			Expect(pkgMetadataList).To(BeNil())
		})
		AfterEach(func() { options = opts })
	})

	Context("success in listing available packages", func() {
		BeforeEach(func() {
			options.Available = true
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageMetadataReturns(packageMetadataList, nil)
			ctl = &pkgClient{kappClient: kappCtl}
			pkgMetadataList, err = ctl.ListPackageMetadata(&options)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(pkgMetadataList).NotTo(BeNil())
			Expect(pkgMetadataList).To(Equal(packageMetadataList))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in listing installed packages due to ListPackageInstalls API error", func() {
		BeforeEach(func() {
			options.ListInstalled = true
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageInstallsReturns(nil, errors.New("failure in ListPackageInstalls"))
			ctl = &pkgClient{kappClient: kappCtl}
			packageInstalls, err = ctl.ListPackageInstalls(&options)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackageInstalls"))
			Expect(packageInstalls).To(BeNil())
		})
		AfterEach(func() { options = opts })
	})

	Context("success in listing installed packages", func() {
		BeforeEach(func() {
			options.ListInstalled = true
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageInstallsReturns(pkgInstallList, nil)
			ctl = &pkgClient{kappClient: kappCtl}
			packageInstalls, err = ctl.ListPackageInstalls(&options)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(packageInstalls).NotTo(BeNil())
			Expect(packageInstalls).To(Equal(pkgInstallList))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in listing package versions due to ListPackages API error", func() {
		BeforeEach(func() {
			options.Available = true
			options.PackageName = testPkgInstallName
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackagesReturns(nil, errors.New("failure in ListPackages"))
			ctl = &pkgClient{kappClient: kappCtl}
			packageVersions, err = ctl.ListPackages(&options)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackages"))
			Expect(packageVersions).To(BeNil())
		})
		AfterEach(func() { options = opts })
	})

	Context("success in listing package versions", func() {
		BeforeEach(func() {
			options.Available = true
			options.PackageName = testPkgInstallName
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			ctl = &pkgClient{kappClient: kappCtl}
			packageVersions, err = ctl.ListPackages(&options)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(packageVersions).NotTo(BeNil())
			Expect(packageVersions).To(Equal(testPkgVersionList))
		})
		AfterEach(func() { options = opts })
	})
})
