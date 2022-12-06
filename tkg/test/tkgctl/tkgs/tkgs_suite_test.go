// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package tkgs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

var (
	// path to the e2e config file
	e2eConfigPath string
	// path to store test artifacts
	artifactsFolder string
	// config read from e2eConfigPath to be used in the tests
	e2eConfig                      *framework.E2EConfig
	logsDir                        string
	err                            error
	deleteClusterOptions           tkgctl.DeleteClustersOptions
	clusterOptions                 tkgctl.CreateClusterOptions
	tkgctlOptions                  tkgctl.Options
	tkgctlClient                   tkgctl.TKGClient
	isClusterClassFeatureActivated bool
	isTKCAPIFeatureActivated       bool
)

func TestE2E(t *testing.T) {
	if folder, ok := os.LookupEnv("E2E_ARTIFACTS"); ok {
		artifactsFolder = folder
	}
	if artifactsFolder == "" {
		artifactsFolder = filepath.Join(os.TempDir(), "artifacts")
	}
	if configPath, ok := os.LookupEnv("E2E_CONFIG"); ok {
		e2eConfigPath = configPath
	}

	RegisterFailHandler(Fail)
	junitPath := filepath.Join(artifactsFolder, "junit", fmt.Sprintf("junit.e2e_suite.%d.xml", config.GinkgoConfig.ParallelNode))
	junitReporter := reporters.NewJUnitReporter(junitPath)
	RunSpecsWithDefaultAndCustomReporters(t, "tkgctl-tkgs-e2e", []Reporter{junitReporter})
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Before all parallel nodes

	Expect(e2eConfigPath).To(BeAnExistingFile(), "e2e config file is either not set or invalid")
	Expect(os.MkdirAll(artifactsFolder, 0o700)).To(Succeed(), "can't create artifacts directory %q", artifactsFolder)

	logsDir := filepath.Join(artifactsFolder, "logs")
	Expect(os.MkdirAll(logsDir, 0o700)).To(Succeed(), "can't create logs directory %q", logsDir)

	By(fmt.Sprintf("loading the e2e test configuration from %q", e2eConfigPath))
	e2eConfig = framework.LoadE2EConfig(context.TODO(), framework.E2EConfigInput{ConfigPath: e2eConfigPath})
	Expect(e2eConfigPath).ToNot(BeNil(), "failed to load e2e config from %s", e2eConfigPath)

	Expect(ValidateTKGSConf(e2eConfig)).To(Succeed(), "e2e test configuration is not valid")

	logsDir = filepath.Join(artifactsFolder, "logs")
	tkgctlOptions = tkgctl.Options{
		ConfigDir:   e2eConfig.TkgConfigDir,
		KubeConfig:  e2eConfig.TKGSKubeconfigPath,
		KubeContext: e2eConfig.TKGSKubeconfigContext,
		LogOptions: tkgctl.LoggingOptions{
			File:      filepath.Join(logsDir, "tkgs-create-wc.log"),
			Verbosity: e2eConfig.TkgCliLogLevel,
		},
	}
	clusterOptions = tkgctl.CreateClusterOptions{
		Edition:    "tkg",
		SkipPrompt: true,
		Plan:       "dev",
	}
	tkgctlClient, err = tkgctl.New(tkgctlOptions)
	Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("failed to connect cluster with given input kube config file:%v and context:%v, reason: %v", e2eConfig.TKGSKubeconfigPath, e2eConfig.TKGSKubeconfigContext, err))
	isTKGS, err := tkgctlClient.IsPacificRegionalCluster()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(isTKGS).To(Equal(true), fmt.Sprintf("the input kube config file:%v with given context:%v is not TKGS cluster", e2eConfig.TKGSKubeconfigPath, e2eConfig.TKGSKubeconfigContext))

	featureGateHelper := tkgctlClient.FeatureGateHelper()
	isClusterClassFeatureActivated, _ = featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
	isTKCAPIFeatureActivated, _ = featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.TKCAPIFeature, constants.TKGSTKCAPINamespace)
	By(fmt.Sprintf("in the tkgs cluster the %v feature is %v, and the %v feature is %v", constants.ClusterClassFeature, isClusterClassFeatureActivated, constants.TKCAPIFeature, isTKCAPIFeatureActivated))

	return []byte(
		strings.Join([]string{
			artifactsFolder,
			e2eConfigPath,
		}, ","),
	)
}, func(data []byte) {
	// Before each parallel node
	parts := strings.Split(string(data), ",")
	Expect(parts).To(HaveLen(2))
	artifactsFolder = parts[0]
	e2eConfigPath = parts[1]
	e2eConfig = framework.LoadE2EConfig(context.TODO(), framework.E2EConfigInput{ConfigPath: e2eConfigPath})
	Expect(e2eConfigPath).ToNot(BeNil(), "failed to load e2e config from %s", e2eConfigPath)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {})

// ValidateTKGSConf validate the configuration in the e2e config file
func ValidateTKGSConf(c *framework.E2EConfig) error {
	if c.InfrastructureName == "" || c.InfrastructureName != constants.InfrastructureProviderTkgs {
		return errors.Errorf("config variable '%v' value is '%v', it should be 'tkgs' for tkgs test suite", "infrastructure_name", c.InfrastructureName)
	}
	return nil
}

func getDeleteClustersOptions(e2eConfig *framework.E2EConfig) tkgctl.DeleteClustersOptions {
	return tkgctl.DeleteClustersOptions{
		ClusterName: e2eConfig.WorkloadClusterOptions.ClusterName,
		Namespace:   e2eConfig.WorkloadClusterOptions.Namespace,
		SkipPrompt:  true,
	}
}

// createClusterConfigFile return temporary cluster config file
// it creates the temporary cluster config file by taking user inputs from the input config file
func createClusterConfigFile(e2eConfig *framework.E2EConfig) string {
	options := framework.CreateClusterOptions{}
	clusterConfigFile, err := framework.GetTempClusterConfigFile(e2eConfig.TkgClusterConfigPath, &options)
	Expect(err).To(BeNil())
	err = e2eConfig.SaveWorkloadClusterOptions(clusterConfigFile)
	Expect(err).To(BeNil())
	return clusterConfigFile
}
