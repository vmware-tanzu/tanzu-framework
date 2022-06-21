// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Addon util test cases", func() {
	Context("ParseStringForLabel", func() {
		var (
			label  string
			result string
		)
		BeforeEach(func() {
		})
		When("label be of short size and no trailing non-alphanumeric characters", func() {
			BeforeEach(func() {
				label = "foobar1.example.com.1.17.2"
				result = ParseStringForLabel(label)
			})
			It("should parse label", func() {
				Expect(result).To(Equal(label))
			})
		})

		When("label be of short size but with trailing non-alphanumeric characters", func() {
			BeforeEach(func() {
				label = "foobar1.example.com.1.17.2."
				result = ParseStringForLabel(label)
			})
			It("should parse label", func() {
				Expect(result).To(Equal(label[:len(label)-1]))
			})
		})
		When("label be of long size and no non-alphanumeric characters at index 63", func() {
			BeforeEach(func() {
				label = "foobar1.very.long.name.for.a.label.exceeding.63.characters.example.com.1.17.2"
				result = ParseStringForLabel(label)
			})
			It("should parse label", func() {
				Expect(result).To(Equal("foobar1.very.long.name.for.a.label.exceeding.63.characters.exam"))
			})
		})
		When("label be of long size and with non-alphanumeric trailing '-' characters just before index 62", func() {
			BeforeEach(func() {
				label = "foobar1.very.long.name.for.a.label.exceeding.63.characters.----ple.com.1.17.2"
				result = ParseStringForLabel(label)
			})
			It("should parse label", func() {
				Expect(result).To(Equal("foobar1.very.long.name.for.a.label.exceeding.63.characters"))
			})
		})

		When("label be of long size and with non-alphanumeric trailing '.' character just before index 62", func() {
			BeforeEach(func() {
				label = "foobar1.very.long.name.for.a.label.exceeding.63.characters.....ple.com.1.17.2"
				result = ParseStringForLabel(label)
			})
			It("should parse label", func() {
				Expect(result).To(Equal("foobar1.very.long.name.for.a.label.exceeding.63.characters"))
			})
		})
		When("label be of long size and with non-alphanumeric trailing '_' character just before index 62", func() {
			BeforeEach(func() {
				label = "foobar1.very.long.name.for.a.label.exceeding.63.characters.____ple.com.1.17.2--"
				result = ParseStringForLabel(label)
			})
			It("should parse label", func() {
				Expect(result).To(Equal("foobar1.very.long.name.for.a.label.exceeding.63.characters"))
			})
		})
	})
})
