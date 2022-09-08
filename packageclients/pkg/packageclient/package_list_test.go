// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/fakes"
	. "github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packageclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

var _ = Describe("List Packages", func() {
	var (
		ctl     PackageClient
		kappCtl *fakes.KappClient
		err     error
		opts    = packagedatamodel.PackageOptions{
			AllNamespaces: false,
			Namespace:     testNamespaceName,
		}
		options       = opts
		optsAvailable = packagedatamodel.PackageAvailableOptions{
			Namespace:     testNamespaceName,
			AllNamespaces: false,
			ValuesSchema:  false,
		}
		optionsAvailable = optsAvailable
		pkgMetadataList  *kapppkg.PackageMetadataList
		packageInstalls  *kappipkg.PackageInstallList
		packageVersions  *kapppkg.PackageList
		pkgInstallList   = &kappipkg.PackageInstallList{
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
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageMetadataReturns(nil, errors.New("failure in ListPackageMetadata"))
			ctl, err = NewPackageClientWithKappClient(kappCtl)
			Expect(err).NotTo(HaveOccurred())
			pkgMetadataList, err = ctl.ListPackageMetadata(&optionsAvailable)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackageMetadata"))
			Expect(pkgMetadataList).To(BeNil())
		})
		AfterEach(func() { optionsAvailable = optsAvailable })
	})

	Context("success in listing available packages", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageMetadataReturns(packageMetadataList, nil)
			ctl, err = NewPackageClientWithKappClient(kappCtl)
			Expect(err).NotTo(HaveOccurred())
			pkgMetadataList, err = ctl.ListPackageMetadata(&optionsAvailable)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(pkgMetadataList).NotTo(BeNil())
			Expect(pkgMetadataList).To(Equal(packageMetadataList))
		})
		AfterEach(func() { optionsAvailable = optsAvailable })
	})

	Context("failure in listing installed packages due to ListPackageInstalls API error", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageInstallsReturns(nil, errors.New("failure in ListPackageInstalls"))
			ctl, err = NewPackageClientWithKappClient(kappCtl)
			Expect(err).NotTo(HaveOccurred())
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
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageInstallsReturns(pkgInstallList, nil)
			ctl, err = NewPackageClientWithKappClient(kappCtl)
			Expect(err).NotTo(HaveOccurred())
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
			optionsAvailable.PackageName = testPkgInstallName
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackagesReturns(nil, errors.New("failure in ListPackages"))
			ctl, err = NewPackageClientWithKappClient(kappCtl)
			Expect(err).NotTo(HaveOccurred())
			packageVersions, err = ctl.ListPackages(&optionsAvailable)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackages"))
			Expect(packageVersions).To(BeNil())
		})
		AfterEach(func() { optionsAvailable = optsAvailable })
	})

	Context("success in listing package versions", func() {
		BeforeEach(func() {
			optionsAvailable.PackageName = testPkgInstallName
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			ctl, err = NewPackageClientWithKappClient(kappCtl)
			Expect(err).NotTo(HaveOccurred())
			packageVersions, err = ctl.ListPackages(&optionsAvailable)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(packageVersions).NotTo(BeNil())
			Expect(packageVersions).To(Equal(testPkgVersionList))
		})
		AfterEach(func() { optionsAvailable = optsAvailable })
	})
})
