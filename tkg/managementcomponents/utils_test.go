// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/managementcomponents"
)

var _ = Describe("Test IsReconciliationError", func() {
	var (
		err                   error
		isReconciliationError bool
	)

	JustBeforeEach(func() {
		isReconciliationError = IsReconciliationError(err)
	})

	Context("When the error is reconciliation timeout error", func() {
		BeforeEach(func() {
			err = errors.New("resource reconciliation timeout: foo bar")
		})
		It("should return true", func() {
			Expect(isReconciliationError).To(BeTrue())
		})
	})

	Context("When the error is reconciliation failed error", func() {
		BeforeEach(func() {
			err = errors.New("'foo' resource reconciliation failed")
		})
		It("should return true", func() {
			Expect(isReconciliationError).To(BeTrue())
		})
	})

	Context("When the error is not a reconciliation error", func() {
		BeforeEach(func() {
			err = errors.New("fake error")
		})
		It("should return false", func() {
			Expect(isReconciliationError).To(BeFalse())
		})
	})
})
