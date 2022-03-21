// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("SetClusterClass", func() {
	var (
		config *fakes.TKGConfigReaderWriter
	)

	BeforeEach(func() {
		config = &fakes.TKGConfigReaderWriter{}
	})

	JustBeforeEach(func() {
		client.SetClusterClass(config)
	})

	Context("User provides a cluster class name", func() {
		BeforeEach(func() {
			config.GetReturns("my-cluster-class", nil)
		})
		It("should apply the user provided cluster class name", func() {
			Expect(config.GetCallCount()).To(Equal(1))
			Expect(config.SetCallCount()).To(Equal(1))
			varName, varValue := config.SetArgsForCall(0)
			Expect(varName).To(Equal("CLUSTER_CLASS"))
			Expect(varValue).To(Equal("my-cluster-class"))
		})
	})

	Context("Determine cluster class name by provider", func() {
		BeforeEach(func() {
			config.GetReturnsOnCall(0, "", nil)
			config.GetReturnsOnCall(1, "vsphere", nil)
		})
		It("should set the name based on the provider", func() {
			Expect(config.GetCallCount()).To(Equal(2))
			Expect(config.SetCallCount()).To(Equal(1))
			varName, varValue := config.SetArgsForCall(0)
			Expect(varName).To(Equal("CLUSTER_CLASS"))
			Expect(varValue).To(Equal("tkg-vsphere-default"))
		})
	})
})
