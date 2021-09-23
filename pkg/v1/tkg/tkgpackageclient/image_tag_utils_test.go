// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test image utils", func() {
	Context("check tag in image URL", func() {

		It("should have error if image url isn't valid", func() {
			// case 1
			repository, tag, err := ParseImageUrl("sftp://user:passwd@example.com/foo/bar:latest")
			Expect(err).To(HaveOccurred())
			Expect(repository).To(Equal(""))
			Expect(tag).To(Equal(""))
		})

		It("should give the correct tag when tag is specified", func() {
			// case 1
			repository, tag, err := ParseImageUrl("foo/bar:1.1")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("docker.io/foo/bar"))
			Expect(tag).To(Equal("1.1"))

			// case 2
			repository, tag, err = ParseImageUrl("http://localhost.localdomain:5000/foo/bar:latest")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("localhost.localdomain:5000/foo/bar"))
			Expect(tag).To(Equal("latest"))
		})

		It("should give the empty tag when tag is not specified", func() {
			// case 1
			repository, tag, err := ParseImageUrl("foo/bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("docker.io/foo/bar"))
			Expect(tag).To(Equal(""))

			// case 2
			repository, tag, err = ParseImageUrl("http://localhost.localdomain:5000/foo/bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(repository).To(Equal("localhost.localdomain:5000/foo/bar"))
			Expect(tag).To(Equal(""))
		})
	})

	//Context("check tagSelection field in PackageRepository CRD", func() {
	//
	//	It("should find tagSelection", func() {
	//		found, err := checkPackageRepositoryTagSelection()
	//		Expect(err).NotTo(HaveOccurred())
	//		Expect(found).To(Equal(true))
	//	})
	//
	//})
})
