// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("PackageAvailableList", func() {
	var (
		kappClient *fakes.KappClient
	)

	Context("getPackageLatestVersion()", func() {
		var (
			incorrectVersion = "incorrect-version-format"
			pkgName          = "cert-manager.tanzu.vmware.com"
			pkgNamespace     = defaultString
		)

		BeforeEach(func() {
			kappClient = &fakes.KappClient{}
		})

		It("should return the latest package version if no errors", func() {
			// tanzu package version does not include leading `v`
			pkgV1 := "1.1.0+vmware.1-tkg.2-zshippable"
			pkgV2 := "1.1.0+vmware.2-tkg.2-zshippable"
			pkgV3 := "1.1.0+vmware.2-tkg.3-zshippable"
			pkgV4 := "1.1.1+vmware.1-tkg.2-zshippable"
			pkgV5 := "1.1.2+vmware.1-tkg.2-zshippable"
			pkgV6 := "1.2.0+vmware.1-tkg.2-zshippable"

			objMeta := metav1.ObjectMeta{
				Name:      pkgName,
				Namespace: pkgNamespace,
			}

			kappClient.ListPackagesReturns(&v1alpha1.PackageList{
				Items: []v1alpha1.Package{
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV1,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV2,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV3,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV4,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV5,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV6,
						},
					},
				},
			}, nil)

			latestVersion, err := getPackageLatestVersion(pkgName, pkgNamespace, kappClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(latestVersion).To(Equal(pkgV6))
		})

		It("should still return the latest package version even if there is one package version which does not"+
			" follow the semver standard", func() {
			// tanzu package version does not include leading `v`
			pkgV1 := "1.2.1+vmware.1-tkg.2-zshippable"
			pkgV3 := "1.3.0+vmware.1-tkg.2-zshippable"

			objMeta := metav1.ObjectMeta{
				Name:      pkgName,
				Namespace: pkgNamespace,
			}

			kappClient.ListPackagesReturns(&v1alpha1.PackageList{
				Items: []v1alpha1.Package{
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV1,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: incorrectVersion,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: pkgV3,
						},
					},
				},
			}, nil)

			latestVersion, err := getPackageLatestVersion(pkgName, pkgNamespace, kappClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(latestVersion).To(Equal(pkgV3))
		})

		It("should still return empty version if there errors to parse all package versions", func() {
			objMeta := metav1.ObjectMeta{
				Name:      pkgName,
				Namespace: pkgNamespace,
			}

			kappClient.ListPackagesReturns(&v1alpha1.PackageList{
				Items: []v1alpha1.Package{
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: incorrectVersion,
						},
					},
					{
						ObjectMeta: objMeta,
						Spec: v1alpha1.PackageSpec{
							RefName: pkgName,
							Version: incorrectVersion,
						},
					},
				},
			}, nil)

			latestVersion, err := getPackageLatestVersion(pkgName, pkgNamespace, kappClient)
			Expect(err).To(HaveOccurred())
			Expect(latestVersion).To(BeEmpty())
		})

		It("should return empty version if there is an error to list packages", func() {
			kappClient.ListPackagesReturns(&v1alpha1.PackageList{
				Items: []v1alpha1.Package{},
			}, fmt.Errorf("dummy error to list packages"))

			latestVersion, err := getPackageLatestVersion(pkgName, pkgNamespace, kappClient)
			Expect(err).To(HaveOccurred())
			Expect(latestVersion).To(BeEmpty())
		})
	})
})
