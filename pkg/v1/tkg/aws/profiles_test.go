// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aws_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/aws"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
)

var _ = Describe("ListCredentialProfiles", func() {
	var (
		err      error
		fileName string
		result   []string
	)

	JustBeforeEach(func() {
		result, err = aws.ListCredentialProfiles(fileName)
	})

	Context("when the filename is specified", func() {
		BeforeEach(func() {
			fileName = "../fakes/config/aws_credentials"
		})
		It("should return the correct list of profile names", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(len(result)).To(Equal(2))
			Expect(result).To(ConsistOf([]string{constants.DefaultNamespace, "test"}))
		})
	})
})
