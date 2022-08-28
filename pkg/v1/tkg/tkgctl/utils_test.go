// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"

	"os"
)

var _ = Describe("Unit tests for ceip", func() {
	var (
		tkgctlClient *tkgctl
	)

	BeforeEach(func() {
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("../fakes/config/config.yaml", "../fakes/config/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		tkgctlClient = &tkgctl{
			tkgConfigReaderWriter: tkgConfigReaderWriter,
		}
	})

	Context("CEIP value is set to true in the config file", func() {
		BeforeEach(func() {
			tkgctlClient.tkgConfigReaderWriter.Set(constants.ConfigVariableEnableCEIPParticipation, "true")
		})

		It("When build edition is tce", func() {
			ceipOptinStatus := tkgctlClient.setCEIPOptinBasedOnConfigAndBuildEdition("tce")
			Expect(ceipOptinStatus).To(Equal("true"))
		})
		It("When build edition is tkg", func() {
			ceipOptinStatus := tkgctlClient.setCEIPOptinBasedOnConfigAndBuildEdition("tkg")
			Expect(ceipOptinStatus).To(Equal("true"))
		})

		AfterEach(func() {
			err := os.Unsetenv(constants.ConfigVariableEnableCEIPParticipation)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("CEIP value is set to false in the config file", func() {
		BeforeEach(func() {
			tkgctlClient.tkgConfigReaderWriter.Set(constants.ConfigVariableEnableCEIPParticipation, "false")
		})

		It("When build edition is tce", func() {
			ceipOptinStatus := tkgctlClient.setCEIPOptinBasedOnConfigAndBuildEdition("tce")
			Expect(ceipOptinStatus).To(Equal("false"))
		})
		It("When build edition is tkg", func() {
			ceipOptinStatus := tkgctlClient.setCEIPOptinBasedOnConfigAndBuildEdition("tkg")
			Expect(ceipOptinStatus).To(Equal("false"))
		})

		AfterEach(func() {
			err := os.Unsetenv(constants.ConfigVariableEnableCEIPParticipation)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("CEIP value is not set in the config file", func() {
		BeforeEach(func() {
			err := os.Unsetenv(constants.ConfigVariableEnableCEIPParticipation)
			Expect(err).ToNot(HaveOccurred())
		})

		It("When build edition is tce", func() {
			ceipOptinStatus := tkgctlClient.setCEIPOptinBasedOnConfigAndBuildEdition("tce")
			Expect(ceipOptinStatus).To(Equal("false"))
		})
		It("When build edition is tkg", func() {
			ceipOptinStatus := tkgctlClient.setCEIPOptinBasedOnConfigAndBuildEdition("tkg")
			Expect(ceipOptinStatus).To(Equal("true"))
		})
	})
})
