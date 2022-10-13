// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestCliCorePkgConfigSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cli/core/pkg/config Suite")
}

var (
	configFile configapi.ClientConfig
)
var _ = Describe("config bom test cases", func() {
	Context("when config file is empty", func() {
		BeforeEach(func() {
			configFile = configapi.ClientConfig{}
		})
		It("should initialize ClientOptions", func() {
			addCompatibilityFile(&configFile, "tkg-compatibility")
			Expect(configFile.ClientOptions).NotTo(BeNil())
			Expect(configFile.ClientOptions.CLI).NotTo(BeNil())
			isMissing := AddCompatibilityFileIfMissing(&configFile)
			Expect(isMissing).To(BeFalse())
		})
		It("should initialize bom repo", func() {
			addBomRepo(&configFile, "projects.registry.vmware.com/tkg")
			Expect(configFile.ClientOptions.CLI).NotTo(BeNil())
		})
		It("should return true", func() {
			isMissing := AddBomRepoIfMissing(&configFile)
			Expect(isMissing).To(BeTrue())
		})
	})
})
