// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

var _ = Describe("Test image utils", func() {
	Context("check tag in image URL", func() {

		It("should have error if image url isn't valid", func() {
			// case 1
			repository, tag, err := parseRegistryImageURL("sftp://user:passwd@example.com/foo/bar:latest")
			Expect(err).To(HaveOccurred())
			Expect(repository).To(Equal(""))
			Expect(tag).To(Equal(""))
		})

		It("should give the correct tag when tag is specified", func() {
			// case 1
			repository, tag, err := parseRegistryImageURL("foo/bar:1.1")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("docker.io/foo/bar"))
			Expect(tag).To(Equal("1.1"))

			// case 2
			repository, tag, err = parseRegistryImageURL("http://localhost.localdomain:5000/foo/bar:latest")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("localhost.localdomain:5000/foo/bar"))
			Expect(tag).To(Equal("latest"))
		})

		It("should give the empty tag when tag is not specified", func() {
			// case 1
			repository, tag, err := parseRegistryImageURL("foo/bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("docker.io/foo/bar"))
			Expect(tag).To(Equal(""))

			// case 2
			repository, tag, err = parseRegistryImageURL("http://localhost.localdomain:5000/foo/bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("localhost.localdomain:5000/foo/bar"))
			Expect(tag).To(Equal(""))
		})
	})

	Context("get current repository and tag", func() {

		It("should get tag from URL when tagselection is not specified", func() {
			// case 1
			pkgr := &kappipkg.PackageRepository{
				TypeMeta:   metav1.TypeMeta{APIVersion: tkgpackagedatamodel.DefaultAPIVersion, Kind: tkgpackagedatamodel.KindPackageRepository},
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: testNamespaceName},
				Spec: kappipkg.PackageRepositorySpec{Fetch: &kappipkg.PackageRepositoryFetch{
					ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: "localhost.localdomain:5000/foo/bar"},
				}},
			}
			repository, tag, err := GetCurrentRepositoryAndTagInUse(pkgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(repository).To(Equal("localhost.localdomain:5000/foo/bar"))
			Expect(tag).To(Equal(tkgpackagedatamodel.LatestReleaseTag))

			// case 2
			pkgr = &kappipkg.PackageRepository{
				TypeMeta:   metav1.TypeMeta{APIVersion: tkgpackagedatamodel.DefaultAPIVersion, Kind: tkgpackagedatamodel.KindPackageRepository},
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: testNamespaceName},
				Spec: kappipkg.PackageRepositorySpec{Fetch: &kappipkg.PackageRepositoryFetch{
					ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: "projects-stg.registry.vmware.com/tkg/test-packages/test-repo:v1.1.0"},
				}},
			}
			repository, tag, err = GetCurrentRepositoryAndTagInUse(pkgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(repository).To(Equal("projects-stg.registry.vmware.com/tkg/test-packages/test-repo"))
			Expect(tag).To(Equal("v1.1.0"))

			// case 3
			pkgr = &kappipkg.PackageRepository{
				TypeMeta:   metav1.TypeMeta{APIVersion: tkgpackagedatamodel.DefaultAPIVersion, Kind: tkgpackagedatamodel.KindPackageRepository},
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: testNamespaceName},
				Spec: kappipkg.PackageRepositorySpec{Fetch: &kappipkg.PackageRepositoryFetch{
					ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: "foo/bar:latest"},
				}},
			}
			repository, tag, err = GetCurrentRepositoryAndTagInUse(pkgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(repository).To(Equal("docker.io/foo/bar"))
			Expect(tag).To(Equal("latest"))
		})
	})
})
