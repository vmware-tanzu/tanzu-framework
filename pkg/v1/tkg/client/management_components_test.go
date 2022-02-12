// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

var _ = Describe("Unit tests for GetUserConfigVariableValueMap", func() {
	var (
		err                      error
		tkgClient                *TkgClient
		configFilePath           string
		configFileData           string
		userProviderConfigValues map[string]string
		rw                       tkgconfigreaderwriter.TKGConfigReaderWriter
	)

	sampleConfigFileData1 := `
#@data/values
#@overlay/match-child-defaults missing_ok=True
---
ABC:
PQR: ""
`
	sampleConfigFileData2 := ``

	BeforeEach(func() {
		rw, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile("", "../fakes/config/config.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		configFilePath = writeConfigFileData(configFileData)
		userProviderConfigValues, err = tkgClient.GetUserConfigVariableValueMap(configFilePath, rw)
	})

	Context("When only one data value is provided by user", func() {
		BeforeEach(func() {
			configFileData = sampleConfigFileData1
			rw.Set("ABC", "abc-value")
		})
		It("returns userProviderConfigValues with ABC", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(1))
			Expect(userProviderConfigValues["ABC"]).To(Equal("abc-value"))
		})
	})

	Context("When all data value is provided by user", func() {
		BeforeEach(func() {
			configFileData = sampleConfigFileData1
			rw.Set("ABC", "abc-value")
			rw.Set("PQR", "pqr-value")
			rw.Set("TEST", "test-value")
		})
		It("returns userProviderConfigValues with ABC and PQR", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(2))
			Expect(userProviderConfigValues["ABC"]).To(Equal("abc-value"))
			Expect(userProviderConfigValues["PQR"]).To(Equal("pqr-value"))
		})
	})

	Context("When no config variables are defined in config default", func() {
		BeforeEach(func() {
			configFileData = sampleConfigFileData2
			rw.Set("TEST", "test-value")
		})
		It("returns empty map", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(0))
		})
	})
})

func writeConfigFileData(configconfigFileData string) string {
	tmpFile, _ := utils.CreateTempFile("", "")
	_ = utils.WriteToFile(tmpFile, []byte(configconfigFileData))
	return tmpFile
}
