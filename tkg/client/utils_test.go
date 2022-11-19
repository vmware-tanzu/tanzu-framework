// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

type MockFeatureFlagClient struct {
	FeatureValues map[string]bool
}

func (m *MockFeatureFlagClient) IsConfigFeatureActivated(featurePath string) (bool, error) {
	if val, ok := m.FeatureValues[featurePath]; ok {
		return val, nil
	}
	return false, errors.Errorf("missing key %s\n", featurePath)
}

var _ = Describe("Utils", func() {
	var (
		tempKubeConfigPath string
		err                error
		contextName        string
		testingDir         string
	)

	BeforeEach((func() {
		testingDir = fakehelper.CreateTempTestingDirectory()
	}))

	AfterEach((func() {
		fakehelper.DeleteTempTestingDirectory(testingDir)
	}))

	Describe("DeleteContextFromKubeConfig tests", func() {
		BeforeEach(func() {
			f, err := os.CreateTemp("", "yaml")
			Expect(err).ToNot(HaveOccurred())
			tempKubeConfigPath = f.Name()
			copyFile("../fakes/config/kubeconfig/config1.yaml", tempKubeConfigPath)
		})
		AfterEach(func() {
			_ = utils.DeleteFile(tempKubeConfigPath)
		})

		JustBeforeEach(func() {
			err = DeleteContextFromKubeConfig(tempKubeConfigPath, contextName)
		})
		Context("When context to be deleted is not present in the kubeconfig file", func() {
			BeforeEach(func() {
				contextName = "fake-nonexisting-context"
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("When context to be deleted is present in the kubeconfig file", func() {
			BeforeEach(func() {
				contextName = "queen-anne-context"
			})
			It("should not return error and delete the context and cluster from kubeconfig file", func() {
				Expect(err).ToNot(HaveOccurred())
				config, err1 := clientcmd.LoadFromFile(tempKubeConfigPath)
				Expect(err1).ToNot(HaveOccurred())
				_, ok := config.Contexts[contextName]
				Expect(ok).To(Equal(false))
				_, ok = config.Clusters["pig-cluster"]
				Expect(ok).To(Equal(false))
			})
		})
		Context("When context to be deleted is present in the kubeconfig file and is current context", func() {
			BeforeEach(func() {
				contextName = "federal-context"
			})
			It("should not return error and delete the context,cluster and also set the current-context to empty string", func() {
				Expect(err).ToNot(HaveOccurred())
				config, err1 := clientcmd.LoadFromFile(tempKubeConfigPath)
				Expect(err1).ToNot(HaveOccurred())
				_, ok := config.Contexts[contextName]
				Expect(ok).To(Equal(false))
				_, ok = config.Clusters["horse-cluster"]
				Expect(ok).To(Equal(false))
				Expect(config.CurrentContext).To(Equal(""))
			})
		})
	})

	Describe("Set machine counts", func() {
		var (
			err       error
			tkgClient *TkgClient
		)

		BeforeEach(func() {
			tkgClient, err = createTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Second)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("When plan is prod", func() {
			plan := "prod"
			It("Get default MC machine counts", func() {
				defaultCPCount, defaultWorkerCount := tkgClient.getMachineCountForMC(plan)
				Expect(defaultCPCount).To(Equal(3))
				Expect(defaultWorkerCount).To(Equal(3))
			})

			It("Override default MC machine counts", func() {
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableControlPlaneMachineCount, "5")
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount, "7")
				defaultCPCount, defaultWorkerCount := tkgClient.getMachineCountForMC(plan)
				Expect(defaultCPCount).To(Equal(5))
				Expect(defaultWorkerCount).To(Equal(7))
			})
		})

		Describe("When plan is dev", func() {
			plan := "dev"
			It("Get default MC machine counts", func() {
				defaultCPCount, defaultWorkerCount := tkgClient.getMachineCountForMC(plan)
				Expect(defaultCPCount).To(Equal(1))
				Expect(defaultWorkerCount).To(Equal(1))
			})

			It("Override default MC machine counts", func() {
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableControlPlaneMachineCount, "5")
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount, "7")
				defaultCPCount, defaultWorkerCount := tkgClient.getMachineCountForMC(plan)
				Expect(defaultCPCount).To(Equal(5))
				Expect(defaultWorkerCount).To(Equal(7))
			})

			It("Use default default machine counts if overrides are even", func() {
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableControlPlaneMachineCount, "4")
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount, "6")
				defaultCPCount, defaultWorkerCount := tkgClient.getMachineCountForMC(plan)
				Expect(defaultCPCount).To(Equal(1))
				Expect(defaultWorkerCount).To(Equal(6))
			})
		})
	})

	Describe("GetCCPlanFromLegacyPlan", func() {
		It("when dev plan is used", func() {
			plan, err := getCCPlanFromLegacyPlan(constants.PlanDev)
			Expect(err).ToNot(HaveOccurred())
			Expect(plan).To(Equal(constants.PlanDevCC))
		})
		It("when prod plan is used", func() {
			plan, err := getCCPlanFromLegacyPlan(constants.PlanProd)
			Expect(err).ToNot(HaveOccurred())
			Expect(plan).To(Equal(constants.PlanProdCC))
		})
		It("when devcc plan is used", func() {
			plan, err := getCCPlanFromLegacyPlan(constants.PlanDevCC)
			Expect(err).ToNot(HaveOccurred())
			Expect(plan).To(Equal(constants.PlanDevCC))
		})
		It("when prodcc plan is used", func() {
			plan, err := getCCPlanFromLegacyPlan(constants.PlanProdCC)
			Expect(err).ToNot(HaveOccurred())
			Expect(plan).To(Equal(constants.PlanProdCC))
		})
		It("when random plan is used", func() {
			_, err := getCCPlanFromLegacyPlan("random")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown plan 'random'"))
		})
	})

	Describe("ensureClusterTopologyConfiguration", func() {
		var (
			err               error
			tkgClient         *TkgClient
			featureFlagClient *MockFeatureFlagClient
			value             string
		)

		BeforeEach(func() {
			featureFlagClient = &MockFeatureFlagClient{map[string]bool{}}
			tkgClient, err = createTKGClientOpts("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Second, func(o Options) Options {
				o.FeatureFlagClient = featureFlagClient
				return o
			})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when feature flag is set to enable CC use", func() {
			BeforeEach(func() {
				featureFlagClient.FeatureValues[constants.FeatureFlagPackageBasedCC] = true
			})

			It("The cluster topology configuration is always set to true", func() {
				tkgClient.ensureClusterTopologyConfiguration()
				value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterTopology)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal("true"))

				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterTopology, "false")
				tkgClient.ensureClusterTopologyConfiguration()
				value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterTopology)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal("true"))

				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterTopology, "true")
				tkgClient.ensureClusterTopologyConfiguration()
				value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterTopology)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal("true"))
			})
		})

		Context("when feature flag is set to not enable CC use", func() {
			BeforeEach(func() {
				featureFlagClient.FeatureValues[constants.FeatureFlagPackageBasedCC] = false
			})

			Context("when CLUSTER_TOPOLOGY is explicitly overridden", func() {
				It("The retains the value", func() {
					var value string //nolint:govet
					tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterTopology, "false")
					tkgClient.ensureClusterTopologyConfiguration()
					value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterTopology)
					Expect(err).NotTo(HaveOccurred())
					Expect(value).To(Equal("false"))

					tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterTopology, "true")
					tkgClient.ensureClusterTopologyConfiguration()
					value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterTopology)
					Expect(err).NotTo(HaveOccurred())
					Expect(value).To(Equal("true"))
				})
			})

			Context("when CLUSTER_TOPOLOGY is not previously set", func() {
				It("The cluster topology configuration is set to false", func() {
					tkgClient.ensureClusterTopologyConfiguration()
					value, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterTopology)
					Expect(err).NotTo(HaveOccurred())
					Expect(value).To(Equal("false"))
				})
			})
		})

	})

})
