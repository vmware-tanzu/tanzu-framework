// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package net_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	netutil "github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/net"
)

func TestNetUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Net util suite")
}

var _ = Describe("Overriding ports in an endpoint", func() {
	Context("when a port is provided", func() {
		It("should override the port", func() {
			endpoint := "https://foo.com:1234"
			result, err := netutil.SetPort(endpoint, 443)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("https://foo.com:443"))
		})
	})
	Context("when no scheme is provided", func() {
		It("should override the port", func() {
			endpoint := "foo.com"
			result, err := netutil.SetPort(endpoint, 443)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("https://foo.com:443"))
		})
	})
	Context("when no scheme is provided, but a port is set", func() {
		It("should override the port", func() {
			endpoint := "foo.com:6443"
			result, err := netutil.SetPort(endpoint, 443)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("https://foo.com:443"))
		})
	})
	Context("when a port with the same value as the overridden value is in the endpoint", func() {
		It("should preserve the port", func() {
			endpoint := "https://foo.com:443"
			result, err := netutil.SetPort(endpoint, 443)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("https://foo.com:443"))
		})
	})
})
