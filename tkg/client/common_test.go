// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	"fmt"
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

	Context("when ALLOW_LEGACY_CLUSTER is explicitly overridden", func() {
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

})
