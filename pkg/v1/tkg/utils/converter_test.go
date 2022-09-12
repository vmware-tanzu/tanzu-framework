// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

var _ = Describe("Test Converter function Tests", func() {
	Context("Test conversion with string types", func() {
		It("Should return the expected output", func() {
			Expect(Convert("abc")).To(Equal("abc"))
			Expect(Convert("pqr")).To(Equal("pqr"))
			Expect(Convert("a")).To(Equal("a"))
		})
	})

	Context("Test conversion with int types", func() {
		It("Should return the expected output", func() {
			Expect(Convert("1")).To(Equal(uint64(1)))
			Expect(Convert("22")).To(Equal(uint64(22)))
			Expect(Convert("100")).To(Equal(uint64(100)))
		})
	})

	Context("Test conversion with boolean types", func() {
		It("Should return the expected output", func() {
			Expect(Convert("true")).To(Equal(true))
			Expect(Convert("false")).To(Equal(false))
			Expect(Convert("True")).To(Equal(true))
			Expect(Convert("False")).To(Equal(false))
		})
	})

	Context("Test conversion with null value", func() {
		It("Should return the expected output", func() {
			Expect(Convert("null")).To(BeNil())
		})
	})

	Context("Test conversion with float value", func() {
		It("Should return the expected output", func() {
			Expect(Convert("1.2")).To(Equal(1.2))
			Expect(Convert("100.212")).To(Equal(100.212))
		})
	})
})
