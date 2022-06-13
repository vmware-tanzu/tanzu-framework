// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package tkgs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework/exec"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

var _ = Describe("TKGS - Create workload cluster use cases", func() {
	var (
		logsDir              string
		clusterConfigFile    string
		err                  error
		deleteClusterOptions tkgctl.DeleteClustersOptions
		clusterOptions       tkgctl.CreateClusterOptions
		tkgctlOptions        tkgctl.Options
		tkgctlClient         tkgctl.TKGClient
	)
	BeforeEach(func() {
		logsDir = filepath.Join(artifactsFolder, "logs")
		tkgctlOptions = tkgctl.Options{
			ConfigDir:    e2eConfig.TkgConfigDir,
			KubeConfig:   e2eConfig.TKGSKubeconfigPath,
			KubeContext:  e2eConfig.TKGSKubeconfigContext,
			SettingsFile: TKGS_SETTINGS_FILE,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, "tkgs-create-wc.log"),
				Verbosity: e2eConfig.TkgCliLogLevel,
			},
		}
		clusterOptions = tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			SkipPrompt:        true,
		}
	})
	Context("input file is legacy config file - TKC cluster", func() {
		BeforeEach(func() {
			Expect(e2eConfig.TkrVersion).ToNot(BeEmpty(), fmt.Sprintf("the kubernetes_version should not be empty to create legacy TKGS cluster"))
			clusterOptions.TkrVersion = e2eConfig.TkrVersion
		})
		Context("cluster Plan is dev", func() {
			BeforeEach(func() {
				e2eConfig.WorkloadClusterOptions.ClusterPlan = "dev"
				e2eConfig.WorkloadClusterOptions.ClusterName = "tkc-e2e-" + util.RandomString(4)
				deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
				clusterOptions.ClusterConfigFile = createClusterConfigFile(e2eConfig)
			})
			AfterEach(func() {
				defer os.Remove(clusterConfigFile)
			})
			When("cluster class cli feature flag (features.global.package-based-lcm-beta) set true", func() {
				BeforeEach(func() {
					//set the cli feature flag as true -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "true")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})
				It("should create TKC Workload Cluster and delete it", func() {
					By(fmt.Sprintf("creating TKC workload cluster %v in namespace: %v, cli feature flag is enabled", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.CreateCluster(clusterOptions)
					Expect(err).To(BeNil())

					By(fmt.Sprintf("deleting TKC workload cluster %v in namespace: %v", clusterOptions.ClusterName, clusterOptions.Namespace))
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
					Expect(err).To(BeNil())
				})
			})
			When("cluster class cli feature flag (features.global.package-based-lcm-beta) set false", func() {
				BeforeEach(func() {
					//set the cli feature flag as false -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "false")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})
				It("should create TKC Workload Cluster and delete it", func() {
					By(fmt.Sprintf("creating TKC workload cluster %v in namespace: %v, cli feature flag is disabled", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.CreateCluster(clusterOptions)
					Expect(err).To(BeNil())

					By(fmt.Sprintf("deleting TKC workload cluster %v in namespace: %v", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
					Expect(err).To(BeNil())
				})
			})
		})
		Context("Cluster Plan is prod", func() {
			BeforeEach(func() {
				e2eConfig.WorkloadClusterOptions.ClusterPlan = "prod"
				e2eConfig.WorkloadClusterOptions.ClusterName = "tkc-e2e-" + util.RandomString(4)
				deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
				clusterOptions.ClusterConfigFile = createClusterConfigFile(e2eConfig)
			})
			AfterEach(func() {
				defer os.Remove(clusterConfigFile)
			})
			When("cluster class cli feature flag (features.global.package-based-lcm-beta) set true", func() {
				BeforeEach(func() {
					//set the cli feature flag as true -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "true")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})
				It("should create TKC Workload Cluster and delete it", func() {
					By(fmt.Sprintf("creating TKC workload cluster %v in namespace: %v, cli feature flag is enabled", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.CreateCluster(clusterOptions)
					Expect(err).To(BeNil())

					By(fmt.Sprintf("deleting TKC workload cluster %v in namespace: %v ", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
					Expect(err).To(BeNil())
				})
			})
			When("cluster class cli feature flag (features.global.package-based-lcm-beta) set false", func() {
				BeforeEach(func() {
					//set the cli feature flag as false -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "false")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})
				It("should create TKC Workload Cluster and delete it", func() {
					By(fmt.Sprintf("Creating TKC workload cluster %v in namespace: %v, cli feature flag is disabled", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.CreateCluster(clusterOptions)
					Expect(err).To(BeNil())

					By(fmt.Sprintf("Deleting TKC workload cluster %v in namespace: %v", e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace))
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
					Expect(err).To(BeNil())
				})
			})
		})
	})
	Context("input file is Cluster Class based", func() {
		BeforeEach(func() {
			cclusterFile, err := os.ReadFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to read the input cluster class based config file from: %v", e2eConfig.WorkloadClusterOptions.ClusterClassFilePath))
			Expect(cclusterFile).ToNot(BeEmpty(), fmt.Sprintf("the input cluster class based config file should not be empty, file path: %v", e2eConfig.WorkloadClusterOptions.ClusterClassFilePath))
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
		})
		When("cluster class cli feature flag (features.global.package-based-lcm-beta) set true", func() {
			BeforeEach(func() {
				//set the cli feature flag as true -  (features.global.package-based-lcm-beta)
				Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "true")).To(Succeed(), "error while setting CLI feature flag")
				tkgctlClient, err = tkgctl.New(tkgctlOptions)
				Expect(err).To(BeNil())
			})
			It("should create cluster class based workload cluster and delete it", func() {
				By(fmt.Sprintf("creating cluster class based workload cluster, cli feature flag is enabled"))
				err = tkgctlClient.CreateCluster(clusterOptions)
				Expect(err).To(BeNil())

				By(fmt.Sprintf("deleting cluster class based workload cluster"))
				_, ccObject, _ := tkgctl.CheckIfInputFileIsClusterClassBased(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
				err = exec.KubectlWithArgs(context.Background(), e2eConfig.TKGSKubeconfigPath, "--context", e2eConfig.TKGSKubeconfigContext, "delete", "cluster", ccObject.GetName(), "-n", ccObject.GetNamespace())
				Expect(err).To(BeNil())
			})
		})
		When("cluster class cli feature flag (features.global.package-based-lcm-beta) set false", func() {
			BeforeEach(func() {
				//set the cli feature flag as false -  (features.global.package-based-lcm-beta)
				Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "false")).To(Succeed(), "error while setting CLI ClusterClass flag")
				tkgctlClient, err = tkgctl.New(tkgctlOptions)
				Expect(err).To(BeNil())
			})
			It("should return error", func() {
				By(fmt.Sprintf("creating Cluster class based workload cluster, cli feature flag is disabled"))
				err = tkgctlClient.CreateCluster(clusterOptions)
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(constants.ErrorMsgCClassInputFeatureFlagDisabled, config.FeatureFlagPackageBasedLCM)))
			})
		})
	})
})
