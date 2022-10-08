// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kubeconfig provides kubeconfig access functions.
package kubeconfig

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTkgAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cli/core/pkg/auth/tkg/util/kubeconfig Suite")
}

var (
	kubeconfiFilePath  string
	kubeconfiFilePath2 string
	kubeconfiFilePath3 string
)

const ConfigFilePermissions = 0o600

var _ = Describe("Unit tests for kubeconfig use cases", func() {

	Context("when valid kubeconfig file is provided", func() {
		BeforeEach(func() {
			kubeconfiFilePath = "../../../fakes/config/kubeconfig1.yaml"
			kubeconfiFilePath2 = "../../../fakes/config/kubeconfig2.yaml"
			kubeconfiFilePath3 = "../../../fakes/config/kubeconfig3_temp_rnhwe.yaml"
			deleteTempFile(kubeconfiFilePath3)
		})
		It("should merge with existing kubeconf file", func() {
			copyFile(kubeconfiFilePath2, kubeconfiFilePath3)
			kubeconfFileContent, _ := os.ReadFile(kubeconfiFilePath)
			err := MergeKubeConfigWithoutSwitchContext(kubeconfFileContent, kubeconfiFilePath3)
			Expect(err).To(BeNil())
		})
		It("should merge with existing empty kubeconf file", func() {
			kubeconfFileContent, _ := os.ReadFile(kubeconfiFilePath)
			err := MergeKubeConfigWithoutSwitchContext(kubeconfFileContent, kubeconfiFilePath3)
			Expect(err).To(BeNil())
		})
		It("should return value for default kubeconfig file", func() {
			defKubeConf := getDefaultKubeConfigFile()
			Expect(defKubeConf).ToNot(BeNil())
		})
	})
})

func copyFile(sourceFile, destFile string) {
	input, _ := os.ReadFile(sourceFile)
	_ = os.WriteFile(destFile, input, ConfigFilePermissions)
}

func deleteTempFile(filename string) {
	os.Remove(filename)
}
