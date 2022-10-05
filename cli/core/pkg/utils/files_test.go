// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit tests for the files utils", func() {
	Context("Unit tests for path exists", func() {
		It("File does not exist", func() {
			exists := PathExists("/tmp/foo.txt")
			Expect(exists).To(BeFalse())
		})

		It("File exists", func() {
			path, err := os.CreateTemp("/tmp", "bar.txt")
			Expect(err).To(BeNil())
			exists := PathExists(path.Name())
			Expect(exists).To(BeTrue())
			err = os.Remove(path.Name())
			Expect(err).To(BeNil())
		})
	})

	Context("Unit tests for saving a file", func() {
		It("test happy path", func() {
			filePath := "/tmp/testfile"
			fileContent := []byte("Test Content")

			err := SaveFile(filePath, fileContent)
			Expect(err).To(BeNil())

			err = os.Remove(filePath)
			Expect(err).To(BeNil())
		})
	})
})
