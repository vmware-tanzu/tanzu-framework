// Package config_e2e_test provides config command specific E2E test cases
package cli_e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/test/e2e/framework"
)

var _ = framework.CLICoreDescribe("[Tests:E2E][Feature:Command-init-version]", func() {
	var (
		tf *framework.Framework
	)
	BeforeEach(func() {
		tf = framework.NewFramework()
	})
	Context("tests for tanzu init and version commands", func() {
		When("init command executed", func() {
			It("should initialize cli successfully", func() {
				err := tf.CliInit()
				Expect(err).To(BeNil())
			})
		})
		When("version command executed", func() {
			It("should return version info", func() {
				version, err := tf.CliVersion()
				Expect(version).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})
})
