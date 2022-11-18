// Package config_e2e_test provides config command specific E2E test cases
package config_e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/test/e2e/framework"
)

var _ = framework.CLICoreDescribe("[Tests:E2E][Feature:Command-Config]", func() {
	var (
		tf *framework.Framework
	)
	BeforeEach(func() {
		tf = framework.NewFramework()
	})
	Context("config feature flag operations", func() {
		When("new config flag set with value", func() {
			It("should set flag and unset flag successfully", func() {
				randomFlagName := "e2e-test-" + framework.RandomString(4)
				randomFeatureFlagPath := "features.global." + randomFlagName
				flagVal := "true"
				err := tf.Config.ConfigSetFeatureFlag(randomFeatureFlagPath, flagVal)
				Expect(err).To(BeNil())

				val, err := tf.Config.ConfigGetFeatureFlag(randomFeatureFlagPath)
				Expect(err).To(BeNil())
				Expect(val).Should(Equal("true"))

				err = tf.Config.ConfigUnsetFeature(randomFeatureFlagPath)
				Expect(err).To(BeNil())

				val, err = tf.Config.ConfigGetFeatureFlag(randomFeatureFlagPath)
				Expect(err).To(BeNil())
				Expect(val).Should(Equal(""))
			})
		})
		When("config init called when config files not exists", func() {
			It("should initialize configuration successfully", func() {
				// delete config files
				err := tf.Config.DeleteCLIConfigurationFiles()
				Expect(err).To(BeNil())
				// call init
				err = tf.Config.ConfigInit()
				Expect(err).To(BeNil())
				// should create config files
				Expect(tf.Config.IsCLIConfigurationFilesExists()).To(BeTrue())

				// set feature flag
				randomFlagName := "e2e-test-" + framework.RandomString(4)
				randomFeatureFlagPath := "features.global." + randomFlagName
				flagVal := "true"
				err = tf.Config.ConfigSetFeatureFlag(randomFeatureFlagPath, flagVal)
				Expect(err).To(BeNil())

				val, err := tf.Config.ConfigGetFeatureFlag(randomFeatureFlagPath)
				Expect(err).To(BeNil())
				Expect(val).Should(Equal("true"))

				// call init
				err = tf.Config.ConfigInit()
				Expect(err).To(BeNil())
				// second run of init should not remove the existing feature flag
				val, err = tf.Config.ConfigGetFeatureFlag(randomFeatureFlagPath)
				Expect(err).To(BeNil())
				Expect(val).Should(Equal("true"))

				// unset the feature flag
				err = tf.Config.ConfigUnsetFeature(randomFeatureFlagPath)
				Expect(err).To(BeNil())
			})
		})
	})
})
