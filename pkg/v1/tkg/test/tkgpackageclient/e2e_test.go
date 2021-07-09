// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	packagelib "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/package/test/lib"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

type E2EConfig struct {
	CreateCluster     bool   `json:"create-cluster"`
	RepositoryURL     string `json:"repository-url"`
	ClusterNameMC     string `json:"mc-cluster-name"`
	ClusterNameWLC    string `json:"wlc-cluster-name"`
	kubeConfigPathMC  string `json:"mc-kubeconfig-Path"`
	kubeConfigPathWLC string `json:"wlc-kubeconfig-Path"`
}

var (
	config                 = &E2EConfig{}
	configPath             string
	tkgCfgDir              string
	err                    error
	packagePlugin          packagelib.PackagePlugin
	result                 packagelib.PackagePluginResult
	clusterCreationTimeout = 30 * time.Minute
	pollInterval           = 10 * time.Second
	pollTimeout            = 3 * time.Minute
	testRepoName           = "test-repo"
	testPkgName            = "pkg.test.carvel.dev"
	testPkgInstallName     = "test-pkg"
	testVersionOne         = "1.0.0"
	testVersionThree       = "3.0.0-rc.1"
	pkgAvailableOptions    = tkgpackagedatamodel.PackageAvailableOptions{}
	pkgOptions             = tkgpackagedatamodel.PackageOptions{
		CreateNamespace: true,
		PackageName:     testPkgName,
		PkgInstallName:  testPkgInstallName,
		Version:         testVersionOne,
	}
	repoOptions = tkgpackagedatamodel.RepositoryOptions{
		CreateNamespace: true,
		RepositoryName:  testRepoName,
	}
)

var _ = Describe("Package plugin integration test", func() {
	var (
		homeDir    string
		currentDir string
	)

	BeforeSuite(func() {
		defaultCfgPath := path.Join(currentDir, "config/e2e_config.yaml")
		flag.StringVar(&configPath, "e2e-config", defaultCfgPath, "path to the e2e config file")

		configData, err := ioutil.ReadFile(configPath)
		Expect(err).NotTo(HaveOccurred())
		err = yaml.Unmarshal(configData, config)
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())

		homeDir, err = os.UserHomeDir()
		Expect(err).NotTo(HaveOccurred())
		currentDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		tkgCfgDir = filepath.Join(homeDir, ".tkg")
		err = os.MkdirAll(tkgCfgDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		if config.ClusterNameMC == "" {
			config.ClusterNameMC = fmt.Sprintf(framework.TkgDefaultClusterPrefix + "mc-test-pkg-plugin-2")
		}

		if config.ClusterNameWLC == "" {
			config.ClusterNameWLC = fmt.Sprintf(framework.TkgDefaultClusterPrefix + "wlc-test-pkg-plugin-2")
		}

		if config.kubeConfigPathMC == "" {
			config.kubeConfigPathMC = filepath.Join(homeDir, ".kube-tkg/config")
		}

		if config.kubeConfigPathWLC == "" {
			config.kubeConfigPathWLC = filepath.Join(tkgCfgDir, config.ClusterNameWLC+".kubeconfig")
		}

		if config.CreateCluster {
			By(fmt.Sprintf("Creating managemnet cluster %q", config.ClusterNameMC))
			cli, err := tkgctl.New(tkgctl.Options{
				ConfigDir: tkgCfgDir,
				LogOptions: tkgctl.LoggingOptions{
					Verbosity: framework.TkgDefaultLogLevel,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			err = cli.Init(tkgctl.InitRegionOptions{
				ClusterConfigFile:      filepath.Join(tkgCfgDir, "cluster-config.yaml"),
				Plan:                   "dev",
				ClusterName:            config.ClusterNameMC,
				InfrastructureProvider: "docker",
				Timeout:                clusterCreationTimeout,
				CniType:                "calico",
			})
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("Creating workload cluster %q", config.ClusterNameWLC))
			tkgCtlClient, err := tkgctl.New(tkgctl.Options{
				ConfigDir: tkgCfgDir,
				LogOptions: tkgctl.LoggingOptions{
					Verbosity: framework.TkgDefaultLogLevel,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			options := framework.CreateClusterOptions{
				ClusterName: config.ClusterNameWLC,
				Namespace:   constants.DefaultNamespace,
				Plan:        "dev",
			}
			config.kubeConfigPathWLC, err = framework.GetTempClusterConfigFile("", &options)
			Expect(err).To(BeNil())

			err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
				ClusterConfigFile: config.kubeConfigPathWLC,
			})
			Expect(err).To(BeNil())
		}
	})

	Context("testing package plugin on management cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.kubeConfigPathMC, pollInterval, pollTimeout, "", "", 0)
		})

		It("should pass all checks on management cluster", func() {
			testHelper()
		})
	})

	Context("testing package plugin on workload cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.kubeConfigPathWLC, pollInterval, pollTimeout, "", "", 0)
		})

		It("should pass all checks on workload cluster", func() {
			testHelper()
		})
	})
})

func testHelper() {
	By("Adding package repository")
	repoOptions.RepositoryURL = config.RepositoryURL
	repoOptions.CreateRepository = true
	result = packagePlugin.UpdateRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("wait for package repository reconciliation")
	result = packagePlugin.CheckRepositoryAvailable(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("list package repository")
	result = packagePlugin.ListRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package repository")
	result = packagePlugin.GetRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("list package available")
	pkgAvailableOptions.AllNamespaces = true
	result = packagePlugin.ListAvailablePackage(&pkgAvailableOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package available")
	pkgAvailableOptions.AllNamespaces = false
	result = packagePlugin.GetAvailablePackage(testPkgName, &pkgAvailableOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("create package install")
	pkgOptions.PollInterval = pollInterval
	pkgOptions.PollTimeout = pollTimeout
	result = packagePlugin.CreateInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("list package install")
	pkgOptions.AllNamespaces = true
	result = packagePlugin.ListInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("update package install")
	pkgOptions.Version = testVersionThree
	result = packagePlugin.UpdateInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package install")
	result = packagePlugin.GetInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("delete package install")
	pkgOptions.PollInterval = pollInterval
	pkgOptions.PollTimeout = pollTimeout
	result = packagePlugin.DeleteInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("delete package repository")
	result = packagePlugin.DeleteRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("wait for package repository deletion")
	result = packagePlugin.CheckRepositoryDeleted(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())
}

func TestTkgpackageclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkgpackageclient Suite")
}
