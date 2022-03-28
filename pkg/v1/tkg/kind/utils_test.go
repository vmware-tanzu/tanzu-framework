// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kind

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	kindClusterProxy KindClusterProxy
)

var _ = Describe("Kind Client", func() {
	BeforeEach(func() {
		kindClusterProxy = KindClusterProxy{
			options: &KindClusterOptions{DefaultImageRepo: "test-repo/tkg"},
		}
	})
	Describe("Only image repository hostname should be used", func() {
		Context("Custom Image Repository not set", func() {
			It("Default hostname from the imagerepos in the TKG BOM should be returned", func() {
				hostName := kindClusterProxy.ResolveHostname("")
				Expect(hostName).To(Equal("test-repo"))
			})
		})

		Context("Custom Image Repository is set", func() {
			It("Default hostname from the imagerepos in the TKG BOM should be returned", func() {
				customImageRepo := "test-custom-repo/tkg"
				hostName := kindClusterProxy.ResolveHostname(customImageRepo)
				Expect(hostName).To(Equal("test-custom-repo"))
			})
		})
	})
})
