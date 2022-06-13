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

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
)

const clusterName = "tkg-cli-wc-e2e-testing"

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
	e2eConfigPath = "/Users/cpamuluri/tkg/tasks/uTKG-Testing/TKGS-tests/tanzu-framework/pkg/v1/tkg/test/config/tkgs.yaml"
	RegisterFailHandler(Fail)
	junitPath := filepath.Join(artifactsFolder, "junit", fmt.Sprintf("junit.e2e_suite.%d.xml", config.GinkgoConfig.ParallelNode))
	junitReporter := reporters.NewJUnitReporter(junitPath)
	RunSpecsWithDefaultAndCustomReporters(t, "tkgctl-tkgs-e2e", []Reporter{junitReporter})
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Before all parallel nodes

	Expect(e2eConfigPath).To(BeAnExistingFile(), "e2e config file is either not set or invalid")
	Expect(os.MkdirAll(artifactsFolder, 0o700)).To(Succeed(), "Can't create artifacts directory %q", artifactsFolder)

	logsDir := filepath.Join(artifactsFolder, "logs")
	Expect(os.MkdirAll(logsDir, 0o700)).To(Succeed(), "Can't create logs directory %q", logsDir)

	By(fmt.Sprintf("Loading the e2e test configuration from %q", e2eConfigPath))
	e2eConfig = framework.LoadE2EConfig(context.TODO(), framework.E2EConfigInput{ConfigPath: e2eConfigPath})
	Expect(e2eConfigPath).ToNot(BeNil(), "Failed to load e2e config from %s", e2eConfigPath)


	
	// save config variables from e2e config to the tkg config file
	//err := e2eConfig.SaveTkgConfigVariables()
	//Expect(err).To(BeNil())

	// validate management cluster is TKGS cluster

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
	Expect(e2eConfigPath).ToNot(BeNil(), "Failed to load e2e config from %s", e2eConfigPath)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	// After all parallel nodes
})
