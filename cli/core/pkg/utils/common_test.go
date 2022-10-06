// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Unit tests for the common utils", func() {
	testStrings := []string{"foo", "bar", "baz"}

	It("String present in input array", func() {
		present := ContainsString(testStrings, "bar")
		Expect(present).To(BeTrue())
	})

	It("String not present in input array", func() {
		present := ContainsString(testStrings, "foobar")
		Expect(present).To(BeFalse())
	})
})
