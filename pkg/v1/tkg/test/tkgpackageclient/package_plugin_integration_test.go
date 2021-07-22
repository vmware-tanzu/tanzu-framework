// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	packagelib "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/package/test/lib"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

type PackagePluginConfig struct {
	UseExistingCluster   bool   `json:"use-existing-cluster"`
	Namespace            string `json:"namespace"`
	PackageName          string `json:"package-name"`
	PackageVersion       string `json:"package-version"`
	PackageVersionUpdate string `json:"package-version-update"`
	RepositoryName       string `json:"repository-name"`
	RepositoryURL        string `json:"repository-url"`
	ClusterNameMC        string `json:"mc-cluster-name"`
	ClusterNameWLC       string `json:"wlc-cluster-name"`
	KubeConfigPathMC     string `json:"mc-kubeconfig-Path"`
	KubeConfigPathWLC    string `json:"wlc-kubeconfig-Path"`
	WithValueFile        bool   `json:"with-value-file"`
}

type repositoryOutput struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
}

type packageInstalledOutput struct {
	Name           string `json:"name"`
	PackageName    string `json:"package-name"`
	PackageVersion string `json:"package-version"`
	Status         string `json:"status"`
}

type packageAvailableOutput struct {
	Name string `json:"name"`
}

var (
	config                 = &PackagePluginConfig{}
	configPath             string
	tkgCfgDir              string
	err                    error
	packagePlugin          packagelib.PackagePlugin
	result                 packagelib.PackagePluginResult
	clusterCreationTimeout = 30 * time.Minute
	pollInterval           = 15 * time.Second
	pollTimeout            = 10 * time.Minute
	standardRepoName       = "tanzu-standard"
	standardNamespace      = "tanzu-package-repo-global"
	standardRepoURL        = "projects-stg.registry.vmware.com/tkg/packageplugin/standard/repo:v1.4.0-zshippable"
	testPkgInstallName     = "test-pkg"
	testPkgName            = "fluent-bit.tanzu.vmware.com"
	testPkgVersion         = "1.7.5+vmware.1-tkg.1-zshippable"
	testPkgVersionUpdate   = "1.7.5+vmware.1-tkg.1-zshippable"
	pkgAvailableOptions    tkgpackagedatamodel.PackageAvailableOptions
	pkgOptions             tkgpackagedatamodel.PackageOptions
	repoOptions            tkgpackagedatamodel.RepositoryOptions
	repoOutput             []repositoryOutput
	expectedRepoOutput     repositoryOutput
	pkgOutput              []packageInstalledOutput
	expectedPkgOutput      packageInstalledOutput
	pkgAvailableOutput     []packageAvailableOutput
)

var _ = Describe("Package plugin integration test", func() {
	var (
		homeDir           string
		clusterConfigFile string
	)

	BeforeSuite(func() {
		flag.StringVar(&configPath, "package-plugin-config", "config/package_plugin_config.yaml", "path to the package plugin config file")

		configData, err := ioutil.ReadFile(configPath)
		Expect(err).NotTo(HaveOccurred())
		err = yaml.Unmarshal(configData, config)
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())

		homeDir, err = os.UserHomeDir()
		Expect(err).NotTo(HaveOccurred())
		tkgCfgDir = filepath.Join(homeDir, ".tkg")
		err = os.MkdirAll(tkgCfgDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		if config.ClusterNameMC == "" {
			config.ClusterNameMC = fmt.Sprintf(framework.TkgDefaultClusterPrefix + "mc-test-pkg-plugin")
		}

		if config.ClusterNameWLC == "" {
			config.ClusterNameWLC = fmt.Sprintf(framework.TkgDefaultClusterPrefix + "wlc-test-pkg-plugin")
		}

		if config.KubeConfigPathMC == "" {
			config.KubeConfigPathMC = filepath.Join(homeDir, ".kube-tkg/config")
		}

		if config.KubeConfigPathWLC == "" {
			config.KubeConfigPathWLC = filepath.Join(tkgCfgDir, config.ClusterNameWLC+".kubeconfig")
		}

		if !config.UseExistingCluster {
			By(fmt.Sprintf("Creating management cluster %q", config.ClusterNameMC))
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
			clusterConfigFile, err = framework.GetTempClusterConfigFile("", &options)
			Expect(err).To(BeNil())
			defer os.Remove(clusterConfigFile)

			err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
				ClusterConfigFile: clusterConfigFile,
			})
			Expect(err).To(BeNil())

			err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: config.ClusterNameWLC,
				Namespace:   constants.DefaultNamespace,
				ExportFile:  config.KubeConfigPathWLC,
			})
			Expect(err).NotTo(HaveOccurred())

			log.Info("Finished creating management and workload clusters")
		}

		if config.Namespace == "" {
			config.Namespace = standardNamespace
		}

		if config.RepositoryName == "" {
			config.RepositoryName = standardRepoName
		}

		if config.RepositoryURL == "" {
			config.RepositoryURL = standardRepoURL
		}

		if config.PackageName == "" {
			config.PackageName = testPkgName
		}

		if config.PackageVersion == "" {
			config.PackageVersion = testPkgVersion
		}

		if config.PackageVersionUpdate == "" {
			config.PackageVersionUpdate = testPkgVersionUpdate
		}

		pkgAvailableOptions = tkgpackagedatamodel.PackageAvailableOptions{
			Namespace: config.Namespace,
		}

		pkgOptions = tkgpackagedatamodel.PackageOptions{
			CreateNamespace: true,
			Namespace:       config.Namespace,
			PackageName:     config.PackageName,
			PkgInstallName:  testPkgInstallName,
			Version:         config.PackageVersion,
		}
		repoOptions = tkgpackagedatamodel.RepositoryOptions{
			CreateNamespace: true,
			Namespace:       config.Namespace,
			RepositoryName:  config.RepositoryName,
		}

		expectedRepoOutput = repositoryOutput{
			Name:       config.RepositoryName,
			Repository: config.RepositoryURL,
			Status:     "Reconcile succeeded",
		}

		expectedPkgOutput = packageInstalledOutput{
			Name:           testPkgInstallName,
			PackageName:    config.PackageName,
			PackageVersion: config.PackageVersion,
			Status:         "Reconcile succeeded",
		}
	})

	Context("testing package plugin on management cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.KubeConfigPathMC, pollInterval, pollTimeout, "json", "", 0)
		})

		It("should pass all checks on management cluster", func() {
			testHelper()
			log.Info("Successfully finished package plugin integration tests on management cluster")
		})
	})

	Context("testing package plugin on workload cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.KubeConfigPathWLC, pollInterval, pollTimeout, "json", "", 0)
		})

		It("should pass all checks on workload cluster", func() {
			testHelper()
			log.Info("Successfully finished package plugin integration tests on workload cluster")
		})
	})
})

func testHelper() {
	By("list package repository")
	repoOptions.AllNamespaces = true
	result = packagePlugin.ListRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &repoOutput)
	Expect(err).ToNot(HaveOccurred())

	By("update package repository")
	repoOptions.RepositoryURL = config.RepositoryURL
	repoOptions.CreateRepository = true
	repoOptions.CreateNamespace = true
	result = packagePlugin.UpdateRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("wait for package repository reconciliation")
	result = packagePlugin.CheckRepositoryAvailable(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package repository")
	result = packagePlugin.GetRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &repoOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(repoOutput)).To(BeNumerically("==", 1))
	Expect(repoOutput[0]).To(Equal(expectedRepoOutput))

	By("list package available without packagename argument")
	pkgAvailableOptions.AllNamespaces = true
	result = packagePlugin.ListAvailablePackage("", &pkgAvailableOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &pkgAvailableOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(pkgAvailableOutput)).To(BeNumerically(">=", 1))
	Expect(pkgAvailableOutput).To(ContainElement(packageAvailableOutput{Name: config.PackageName}))

	By("list package available with packagename argument")
	pkgAvailableOptions.AllNamespaces = true
	result = packagePlugin.ListAvailablePackage(config.PackageName, &pkgAvailableOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &pkgAvailableOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(pkgAvailableOutput).To(ContainElement(packageAvailableOutput{Name: config.PackageName}))

	By("get package available packagename format")
	pkgAvailableOptions.AllNamespaces = false
	result = packagePlugin.GetAvailablePackage(config.PackageName, &pkgAvailableOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &pkgAvailableOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(pkgAvailableOutput)).To(BeNumerically("==", 1))
	Expect(pkgAvailableOutput[0].Name).To(Equal(config.PackageName))

	By("get package available packagename/packageversion format")
	result = packagePlugin.GetAvailablePackage(config.PackageName+"/"+config.PackageVersion, &pkgAvailableOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &pkgAvailableOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(pkgAvailableOutput)).To(BeNumerically("==", 1))
	Expect(pkgAvailableOutput[0].Name).To(Equal(config.PackageName))

	By("create package install")
	pkgOptions.Wait = true
	pkgOptions.PollInterval = pollInterval
	pkgOptions.PollTimeout = pollTimeout
	if config.WithValueFile {
		pkgOptions.ValuesFile = "config/values.yaml"
	}
	result = packagePlugin.CreateInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("list package install")
	pkgOptions.AllNamespaces = true
	result = packagePlugin.ListInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &pkgOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(pkgOutput)).To(BeNumerically(">=", 1))
	Expect(pkgOutput).To(ContainElement(expectedPkgOutput))

	By("update package install")
	pkgOptions.Wait = true
	pkgOptions.Version = config.PackageVersionUpdate
	if config.WithValueFile {
		pkgOptions.ValuesFile = "config/values_update.yaml"
	}
	result = packagePlugin.UpdateInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package install")
	result = packagePlugin.GetInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &pkgOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(pkgOutput)).To(BeNumerically("==", 1))
	Expect(pkgOutput[0]).To(Equal(expectedPkgOutput))

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
