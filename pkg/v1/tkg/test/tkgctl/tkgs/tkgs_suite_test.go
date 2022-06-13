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

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

const CLI_CLUSTERCLASS_FLAG = "features.global.package-based-lcm-beta"
const TKGS_SETTINGS_FILE = "/Users/cpamuluri/tkg/tasks/uTKG-Testing/TKGS-tests/tkgs-inegration-tests/tanzu-framework/pkg/v1/tkg/test/config/config.yaml"

var (
	// path to the e2e config file
	e2eConfigPath string
	// path to store test artifacts
	artifactsFolder string
	// config read from e2eConfigPath to be used in the tests
	e2eConfig *framework.E2EConfig
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

func createClusterConfigFile(e2eConfig *framework.E2EConfig) string {
	options := framework.CreateClusterOptions{}
	clusterConfigFile, err := framework.GetTempClusterConfigFile(e2eConfig.TkgClusterConfigPath, &options)
	Expect(err).To(BeNil())
	err = e2eConfig.SaveWorkloadClusterOptions(clusterConfigFile)
	Expect(err).To(BeNil())
	return clusterConfigFile
}
