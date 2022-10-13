// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCliCorePkgArtifactSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cli/core/pkg/artifact Suite")
}

var _ = Describe("Unit tests for local artifact", func() {
	When("url is http/https", func() {
		It("should not return error", func() {
			uriArtifact, err := NewURIArtifact(dummyURL)
			Expect(uriArtifact).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
	})
	When("url is local", func() {
		It("should not return error", func() {
			uriArtifact, err := NewURIArtifact("file://home/user/")
			Expect(uriArtifact).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
	})
	When("url is default", func() {
		It("should not return error", func() {
			uriArtifact, err := NewURIArtifact("/default")
			Expect(uriArtifact).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
	})
	When("url is not valid", func() {
		It("should return error", func() {
			uriArtifact, err := NewURIArtifact("sql://user:pwd/db")
			Expect(uriArtifact).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
	})
})
