// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

var _ = Describe("Unit tests for GetUserConfigVariableValueMap", func() {
	var (
		err                      error
		tkgClient                *TkgClient
		configFilePath           string
		configFileData           string
		userProviderConfigValues map[string]interface{}
		rw                       tkgconfigreaderwriter.TKGConfigReaderWriter
	)

	sampleConfigFileData1 := `
#@data/values
#@overlay/match-child-defaults missing_ok=True
---
ABC:
PQR: ""
Test1:
Test2:
Test3:
Test4:
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
			rw.Set("Test1", "true")
			rw.Set("Test2", "null")
			rw.Set("Test3", "1")
			rw.Set("Test4", "1.2")
		})
		It("returns userProviderConfigValues with ABC", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(5))
			Expect(userProviderConfigValues["ABC"]).To(Equal("abc-value"))
			Expect(userProviderConfigValues["Test1"]).To(Equal(true))
			Expect(userProviderConfigValues["Test2"]).To(BeNil())
			Expect(userProviderConfigValues["Test3"]).To(Equal(uint64(1)))
			Expect(userProviderConfigValues["Test4"]).To(Equal(1.2))
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

var _ = Describe("Unit tests for GetKappControllerConfigValuesFile", func() {
	var (
		err                               error
		kappControllerValuesYttDir        = "../../../../providers/kapp-controller-values"
		inputDataValuesFile               string
		processedKappControllerValuesFile string
		outputKappControllerValuesFile    string
	)

	validateResult := func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(processedKappControllerValuesFile).NotTo(BeEmpty())
		filedata1, err := readFileData(processedKappControllerValuesFile)
		Expect(err).NotTo(HaveOccurred())
		filedata2, err := readFileData(outputKappControllerValuesFile)
		Expect(err).NotTo(HaveOccurred())
		if strings.Compare(filedata1, filedata2) != 0 {
			log.Infof("Processed Output: %v", filedata1)
			log.Infof("Expected  Output: %v", filedata2)
		}
		Expect(filedata1).To(Equal(filedata2))
	}

	JustBeforeEach(func() {
		processedKappControllerValuesFile, err = GetKappControllerConfigValuesFile(inputDataValuesFile, kappControllerValuesYttDir)
	})

	Context("When no config variables are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/kapp-controller-values/testcase1/uservalues.yaml"
			outputKappControllerValuesFile = "test/kapp-controller-values/testcase1/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

	Context("When codedns, provider type and cidr variables are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/kapp-controller-values/testcase2/uservalues.yaml"
			outputKappControllerValuesFile = "test/kapp-controller-values/testcase2/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

	Context("When custom image repository variables are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/kapp-controller-values/testcase3/uservalues.yaml"
			outputKappControllerValuesFile = "test/kapp-controller-values/testcase3/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

})

func writeConfigFileData(configconfigFileData string) string {
	tmpFile, _ := utils.CreateTempFile("", "")
	_ = utils.WriteToFile(tmpFile, []byte(configconfigFileData))
	return tmpFile
}

func readFileData(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	return string(data), err
}
