// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework/exec"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

var (
	// path to the e2e config file
	e2eConfigPath string

	// path to store test artifacts
	artifactsFolder string

	// config read from e2eConfigPath to be used in the tests
	e2eConfig *framework.E2EConfig

	clusterName string
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
	RunSpecsWithDefaultAndCustomReporters(t, "tkgctl-docker-e2e", []Reporter{junitReporter})
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

	logLocation := filepath.Join(artifactsFolder, "logs")
	hackCmd := exec.NewCommand(
		exec.WithCommand("../../scripts/cc_hack.sh"),
		exec.WithStdout(GinkgoWriter),
	)

	fmt.Println("Executing the hack script")
	out, cmdErr, err := hackCmd.Run(context.Background())
	fmt.Println(string(out))
	fmt.Println(string(cmdErr))
	Expect(err).To(BeNil())

	cli, err := tkgctl.New(tkgctl.Options{
		ConfigDir: e2eConfig.TkgConfigDir,
		LogOptions: tkgctl.LoggingOptions{
			File:      filepath.Join(logLocation, "before_suite.log"),
			Verbosity: e2eConfig.TkgCliLogLevel,
		},
	})
	Expect(err).To(BeNil())

	// save config variables from e2e config to the tkg config file
	err = e2eConfig.SaveTkgConfigVariables()
	Expect(err).To(BeNil())

	timeout, err := time.ParseDuration(e2eConfig.DefaultTimeout)
	Expect(err).To(BeNil())

	// create management cluster
	if !e2eConfig.UseExistingCluster {
		err := cli.Init(tkgctl.InitRegionOptions{
			Plan:                        e2eConfig.ManagementClusterOptions.Plan,
			ClusterName:                 e2eConfig.ManagementClusterName,
			InfrastructureProvider:      e2eConfig.InfrastructureName,
			Timeout:                     timeout,
			Size:                        e2eConfig.ManagementClusterOptions.Size,
			DeployTKGonVsphere7:         e2eConfig.ManagementClusterOptions.DeployTKGonVsphere7,
			EnableTKGSOnVsphere7:        e2eConfig.ManagementClusterOptions.EnableTKGSOnVsphere7,
			VsphereControlPlaneEndpoint: e2eConfig.ManagementClusterOptions.Endpoint,
			CniType:                     "antrea",
			Edition:                     "tkg",
		})

		Expect(err).To(BeNil())
	}

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

	timeout, err := time.ParseDuration(e2eConfig.DefaultTimeout)
	Expect(err).To(BeNil())
	
	logLocation := filepath.Join(artifactsFolder, "logs")
	cli, err := tkgctl.New(tkgctl.Options{
		ConfigDir: e2eConfig.TkgConfigDir,
		LogOptions: tkgctl.LoggingOptions{
			File:      filepath.Join(logLocation, "after_suite.log"),
			Verbosity: e2eConfig.TkgCliLogLevel,
		},
	})
	Expect(err).To(BeNil())

	err = cli.DeleteRegion(tkgctl.DeleteRegionOptions{
		ClusterName: e2eConfig.ManagementClusterName,
		Force:       true,
		SkipPrompt:  true,
		Timeout:     timeout,
	})

	Expect(err).To(BeNil())
})
