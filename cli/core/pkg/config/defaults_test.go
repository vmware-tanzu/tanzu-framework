// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("defaults test cases", func() {
	Context("default locations and repositories", func() {
		It("should initialize ClientOptions", func() {
			artLocations := GetTrustedArtifactLocations()
			Expect(artLocations).NotTo(BeNil())
		})
		It("trusted registries should return value", func() {
			DefaultAllowedPluginRepositories = "https://storage.googleapis.com"
			trustedRegis := GetTrustedRegistries()
			Expect(trustedRegis).NotTo(BeNil())
			DefaultAllowedPluginRepositories = ""
		})
	})
})
