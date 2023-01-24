// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
)

var _ = Describe("SetClusterClass", func() {
	var (
		config *fakes.TKGConfigReaderWriter
	)

	BeforeEach(func() {
		config = &fakes.TKGConfigReaderWriter{}
	})

	JustBeforeEach(func() {
		client.SetClusterClass(config)
	})

	Context("User provides a cluster class name", func() {
		BeforeEach(func() {
			config.GetReturns("my-cluster-class", nil)
		})
		It("should apply the user provided cluster class name", func() {
			Expect(config.GetCallCount()).To(Equal(1))
			Expect(config.SetCallCount()).To(Equal(1))
			varName, varValue := config.SetArgsForCall(0)
			Expect(varName).To(Equal("CLUSTER_CLASS"))
			Expect(varValue).To(Equal("my-cluster-class"))
		})
	})

	Context("Determine cluster class name by provider", func() {
		BeforeEach(func() {
			config.GetReturnsOnCall(0, "", nil)
			config.GetReturnsOnCall(1, "vsphere", nil)
		})
		It("should set the name based on the provider", func() {
			Expect(config.GetCallCount()).To(Equal(2))
			Expect(config.SetCallCount()).To(Equal(1))
			varName, varValue := config.SetArgsForCall(0)
			Expect(varName).To(Equal("CLUSTER_CLASS"))
			// TODO: fix below with TKG-13296.
			Expect(varValue).To(Equal(fmt.Sprintf("tkg-vsphere-default-%s", constants.DefaultClusterClassVersion)))
		})
	})
})

type MockFeatureFlag struct {
	FeatureFlags map[string]bool
}

func (m *MockFeatureFlag) IsConfigFeatureActivated(featurePath string) (bool, error) {
	if val, ok := m.FeatureFlags[featurePath]; ok {
		return val, nil
	}
	return false, errors.Errorf("missing key %s\n", featurePath)
}

var _ = Describe("ensureAllowLegacyClusterConfiguration", func() {
	var (
		err               error
		tkgClient         *client.TkgClient
		featureFlagClient *MockFeatureFlag
		value             string
		testingDir        string
		values            []string
	)

	BeforeEach(func() {
		testingDir = helper.CreateTempTestingDirectory()
		featureFlagClient = &MockFeatureFlag{map[string]bool{}}
		tkgClient, err = client.CreateTKGClientOptsMutator("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Second, func(o client.Options) client.Options {
			o.FeatureFlagClient = featureFlagClient
			return o
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach((func() {
		helper.DeleteTempTestingDirectory(testingDir)
	}))

	Context("when ALLOW_LEGACY_CLUSTER is not previously set", func() {
		It("ALLOW_LEGACY_CLUSTER should keep consistent with FeatureFlagAllowLegacyCluster (false)", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = false
			_ = tkgClient.SetAllowLegacyClusterConfiguration()
			value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAllowLegacyCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("false"))
		})

		It("ALLOW_LEGACY_CLUSTER should keep consistent with FeatureFlagAllowLegacyCluster (true)", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = true
			_ = tkgClient.SetAllowLegacyClusterConfiguration()
			value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAllowLegacyCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("true"))
		})
	})

	Context("when ALLOW_LEGACY_CLUSTER is explicitly overridden with a valid value", func() {
		values = []string{"true", "false"}
		It("Retain the value", func() {
			for _, v := range values {
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableAllowLegacyCluster, v)
				_ = tkgClient.SetAllowLegacyClusterConfiguration()
				value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAllowLegacyCluster)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(v))
			}
		})
	})

	Context("when ALLOW_LEGACY_CLUSTER is explicitly overridden with an invalid value", func() {
		It("Should return true value", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = true
			tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableAllowLegacyCluster, "invalid value")
			_ = tkgClient.SetAllowLegacyClusterConfiguration()
			value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAllowLegacyCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal("true"))
		})

		It("Should return false value", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = false
			tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableAllowLegacyCluster, "invalid value")
			_ = tkgClient.SetAllowLegacyClusterConfiguration()
			value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAllowLegacyCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal("false"))
		})
	})

})

var _ = Describe("shouldDeployClusterClass", func() {
	var (
		err                    error
		tkgClient              *client.TkgClient
		featureFlagClient      *MockFeatureFlag
		tkgConfigUpdaterClient *fakes.TKGConfigUpdaterClient
		testingDir             string
		isManagementCluster    bool
	)

	BeforeEach(func() {
		testingDir = helper.CreateTempTestingDirectory()
		featureFlagClient = &MockFeatureFlag{map[string]bool{}}
		tkgConfigUpdaterClient = &fakes.TKGConfigUpdaterClient{}
		tkgClient, err = client.CreateTKGClientOptsMutator("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Second, func(o client.Options) client.Options {
			o.FeatureFlagClient = featureFlagClient
			o.TKGConfigUpdater = tkgConfigUpdaterClient
			return o
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach((func() {
		helper.DeleteTempTestingDirectory(testingDir)
	}))

	Context("When cluster is mgmt cluster", func() {
		BeforeEach(func() {
			isManagementCluster = true
			tkgConfigUpdaterClient.GetProvidersChecksumStub = func() (string, error) {
				return "fakeFileSumIsSame", nil
			}
			tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFileStub = func() (string, error) {
				return "fakeFileSumIsDifferent", nil
			}
		})

		It("Should deploy classy based cluster with FeatureFlagPackageBasedCC is enabled", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagPackageBasedCC] = true
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})

		It("Should return false with FeatureFlagPackageBasedCC is disabled", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagPackageBasedCC] = false
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(false))
		})

		It("Should return false when checkOverlay fails", func() {
			tkgConfigUpdaterClient.GetProvidersChecksumStub = func() (string, error) {
				return "fakeFileSumIsSame", errors.Errorf("fake error")
			}
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err.Error()).To(Equal("fake error"))
			Expect(result).To(Equal(false))
		})
	})

	Context("When cluster is workload cluster", func() {
		BeforeEach(func() {
			isManagementCluster = false
			tkgConfigUpdaterClient.GetProvidersChecksumStub = func() (string, error) {
				return "fakeFileSumIsSame", nil
			}
			tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFileStub = func() (string, error) {
				return "fakeFileSumIsDifferent", nil
			}
		})

		It("Should return false and error with AllowLeagcyCluster is false when customization exists", func() {
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err.Error()).To(ContainSubstring("It seems like you have done some customizations to the template overlays"))
			Expect(result).To(Equal(false))
		})

		It("Should return true with AllowLeagcyCluster is false when customization doesn't exist", func() {
			tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFileStub = func() (string, error) {
				return "fakeFileSumIsSame", nil
			}
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})

		It("Should return true and with AllowLeagcyCluster and ForceDeployClassyCluster are true", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = true
			featureFlagClient.FeatureFlags[constants.FeatureFlagForceDeployClusterWithClusterClass] = true
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})

		It("Should return false and with AllowLeagcyCluster is true and ForceDeployClassyCluster is false", func() {
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = true
			featureFlagClient.FeatureFlags[constants.FeatureFlagForceDeployClusterWithClusterClass] = false
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(false))
		})

		It("Should return true when AllowLeagcyCluster is false and SUPPRESS_PROVIDERS_UPDATE is set", func() {
			os.Setenv("SUPPRESS_PROVIDERS_UPDATE", "1")
			featureFlagClient.FeatureFlags[constants.FeatureFlagAllowLegacyCluster] = false
			result, err := tkgClient.ShouldDeployClusterClassBasedCluster(isManagementCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
			os.Unsetenv("SUPPRESS_PROVIDERS_UPDATE")
		})
	})
})
