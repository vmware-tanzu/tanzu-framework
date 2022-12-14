// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/isolated-cluster/fakes"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/isolated-cluster/imagepushop"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Publish image from tar file")
}

var _ = Describe("pushImageToRepo()", func() {
	pushImage := &imagepushop.PublishImagesFromTarOptions{}

	BeforeEach(func() {
		pushImage.PkgClient = &fakes.ImgpkgClientFake{}
		pushImage.TkgTarFilePath = "./testdata"

	})

	When("publish-images-fromtar.yaml, which contain tar file name and destination repo path, doesn't existed", func() {
		It("should return err", func() {
			err := pushImage.PushImageToRepo()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error while reading testdata/publish-images-fromtar.yaml file"))
		})
	})
	When("publish-images-fromtar.yaml, which contain tar file name and destination repo path, has wrong format", func() {
		It("should return err", func() {
			err := utils.CopyFile("./testdata/publish-images-fromtar_with_error", "./testdata/publish-images-fromtar.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = pushImage.PushImageToRepo()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error while parsing publish-images-fromtar.yaml file"))
			err = utils.DeleteFile("./testdata/publish-images-fromtar.yaml")
			Expect(err).ToNot(HaveOccurred())
		})
	})
	When("PushImageToRepo successful", func() {
		It("should return nil", func() {
			err := utils.CopyFile("./testdata/publish-images-fromtar", "./testdata/publish-images-fromtar.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = pushImage.PushImageToRepo()
			Expect(err).ToNot(HaveOccurred())
			err = utils.DeleteFile("./testdata/publish-images-fromtar.yaml")
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
