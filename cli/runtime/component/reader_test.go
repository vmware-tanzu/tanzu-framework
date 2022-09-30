// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestComponent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Suite")
}

var _ = Describe("Unit testing the reader", func() {
	Context("Test reading from file", func() {
		It("test happy path", func() {
			file, err := os.CreateTemp("/tmp", "foo")
			Expect(err).ToNot(HaveOccurred())

			testContent := "Test content"
			_, err = file.WriteString(testContent)
			Expect(err).ToNot(HaveOccurred())

			contents, err := ReadInput(file.Name())
			Expect(err).ToNot(HaveOccurred())

			Expect(string(contents)).To(Equal(testContent))

			err = os.Remove(file.Name())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
