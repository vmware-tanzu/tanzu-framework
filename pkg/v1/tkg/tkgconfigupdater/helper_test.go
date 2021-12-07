// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	sampleFilePath string
	destDir        string
)

var _ = Describe("Tests while unzipping a file", func() {
	Context("Validating destDir path", func() {
		It("when file path is invalid", func() {
			sampleFilePath = "baz"
			destDir = "foo/../bar"
			err := unzip(sampleFilePath, destDir)
			Expect(err).To(HaveOccurred())
		})
	})
})
