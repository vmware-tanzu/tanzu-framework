// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package features

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Feature Flag Client Suite")
}

var _ = Describe("Feature flag client", func() {
	var (
		featureClient Client
		featureFlags  map[string]string
		err           error
		isEnabled     bool
		flags         map[string]string
		filePath      string
		flagName      string
	)

	const (
		validFeaturesJSONFilePath     = "../fakes/config/features/features_valid.json"
		invalidFeaturesJSONFilePath   = "../fakes/config/features/features_invalid.json"
		duplicateFeaturesJSONFilePath = "../fakes/config/features/features_duplicate.json"
		featuresWritesJSONFilePath    = "../fakes/config/features/features_write.json"
		featureOne                    = "feature1"
		featureTwo                    = "feature2"
	)
	Describe("GetFeatureFlags", func() {
		JustBeforeEach(func() {
			featureClient = &client{
				featureFlagConfigPath: filePath,
			}
			featureFlags, err = featureClient.GetFeatureFlags()
		})

		Context("with valid json", func() {
			BeforeEach(func() {
				filePath = validFeaturesJSONFilePath
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should contain feature flags", func() {
				Expect(len(featureFlags)).To(Equal(2))
				Expect(featureFlags[featureOne]).To(Equal("true"))
				Expect(featureFlags[featureTwo]).To(Equal("http://google.com"))
			})
		})
		Context("with invalid json", func() {
			BeforeEach(func() {
				filePath = invalidFeaturesJSONFilePath
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal feature flag json data"))
			})
		})
		Context("with duplicate json", func() {
			BeforeEach(func() {
				filePath = duplicateFeaturesJSONFilePath
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should contain no duplicate flag", func() {
				Expect(len(featureFlags)).To(Equal(1))
			})
		})
	})

	Describe("IsFeatureFlagEnabled", func() {
		JustBeforeEach(func() {
			featureClient = &client{
				featureFlagConfigPath: filePath,
			}
			isEnabled, err = featureClient.IsFeatureFlagEnabled(flagName)
		})

		Context("with valid json", func() {
			BeforeEach(func() {
				filePath = validFeaturesJSONFilePath
				flagName = featureOne
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should contain feature flags", func() {
				Expect(isEnabled).To(BeTrue())
			})
		})
		Context("with invalid json", func() {
			BeforeEach(func() {
				filePath = invalidFeaturesJSONFilePath
				flagName = featureOne
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal feature flag json data"))
			})
		})
		Context("with duplicate json", func() {
			BeforeEach(func() {
				filePath = duplicateFeaturesJSONFilePath
				flagName = featureOne
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return second of duplicates", func() {
				Expect(isEnabled).To(BeTrue())
			})
		})
		Context("when flag doesn't exist in file", func() {
			BeforeEach(func() {
				filePath = validFeaturesJSONFilePath
				flagName = "feature_does_not_exist"
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return false", func() {
				Expect(isEnabled).To(BeFalse())
			})
		})
	})

	Describe("WriteFeatureFlags", func() {
		JustBeforeEach(func() {
			featureClient = &client{
				featureFlagConfigPath: filePath,
			}
			err = featureClient.WriteFeatureFlags(flags)
			featureFlags, _ = featureClient.GetFeatureFlags()
		})

		Context("with valid json", func() {
			BeforeEach(func() {
				filePath = featuresWritesJSONFilePath
				flags = map[string]string{
					featureOne: "true",
					featureTwo: "false",
				}
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should contain feature flags", func() {
				Expect(len(featureFlags)).To(Equal(2))
			})
		})
		Context("with data in map", func() {
			BeforeEach(func() {
				filePath = featuresWritesJSONFilePath
				flags = map[string]string{
					featureOne: "http://google.com",
					featureTwo: "antrea",
				}
			})
			It("should not return error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should contain data", func() {
				Expect(len(featureFlags)).To(Equal(2))
				Expect(featureFlags[featureOne]).To(Equal("http://google.com"))
				Expect(featureFlags[featureTwo]).To(Equal("antrea"))
			})
		})
		Context("with empty map", func() {
			BeforeEach(func() {
				filePath = featuresWritesJSONFilePath
				flags = map[string]string{}
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should contain nothing", func() {
				Expect(len(featureFlags)).To(Equal(0))
			})
		})
		Context("with data already in file", func() {
			BeforeEach(func() {
				filePath = "../fakes/config/features/features_overwrite.json"
				flags = map[string]string{
					featureOne: "true",
					featureTwo: "false",
				}
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should overwrite existing data", func() {
				Expect(len(featureFlags)).To(Equal(2))
				Expect(featureFlags["overwriteMe"]).To(Not(BeNil()))
			})
		})
		Context("when file doesn't exist", func() {
			BeforeEach(func() {
				filePath = "../fakes/config/features/features_new.json"
				flags = map[string]string{
					featureOne: "true",
					featureTwo: "false",
				}
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should contain data", func() {
				Expect(len(featureFlags)).To(Equal(2))
			})
			AfterEach(func() {
				os.Remove("../fakes/config/features/features_new.json")
			})
		})
	})
})
