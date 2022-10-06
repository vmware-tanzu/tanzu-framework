// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit tests for gcp", func() {
	discover := &GCPDiscovery{
		bucketName:   "foo",
		manifestPath: "bar",
		name:         "gcp-name",
	}
	It("test getting gcp name", func() {
		name := discover.Name()
		Expect(name).To(Equal(discover.name))
	})

	It("test getting gcp type", func() {
		name := discover.Type()
		Expect(name).To(Equal("gcp"))
	})
})
