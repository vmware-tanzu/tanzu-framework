// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/carvelhelpers"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cli/core/pkg/carvelhelpers Suite")
}

var _ = Describe("Unit tests for CarvelPackageProcessor", func() {
	var (
		configBytes        []byte
		err                error
		packageDownloadDir string
		image              string
		outputFilePath     string
	)

	JustBeforeEach(func() {
		configBytes, err = CarvelPackageProcessor(packageDownloadDir, image)
	})

	Context("When processing test package1 which includes .imgpkg dir", func() {
		BeforeEach(func() {
			packageDownloadDir = "./test/package1/input"
			outputFilePath = "./test/package1/output.yaml"
		})
		It("should not return an error", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(configBytes)).To(Equal(readFile(outputFilePath)))
		})
	})

	Context("When processing test package2 which does not include .imgpkg dir", func() {
		BeforeEach(func() {
			packageDownloadDir = "./test/package2/input"
			outputFilePath = "./test/package2/output.yaml"
		})
		It("should not return an error", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(configBytes)).To(Equal(readFile(outputFilePath)))
		})
	})
})

func readFile(path string) string {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

var _ = Describe("Unit tests for CarvelPackageProcessor", func() {
	var (
		configBytes []byte
		err         error
		image       string
	)

	JustBeforeEach(func() {
		configBytes, err = ProcessCarvelPackage(image, "")
	})

	Context("When in-correct image passed", func() {
		BeforeEach(func() {
			image = "image"
		})
		It("should not return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get resource files from discovery"))
			Expect(configBytes).To(BeNil())
		})
	})

})
