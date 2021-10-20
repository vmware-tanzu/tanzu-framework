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
	secretlib "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/secret/test/lib"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

type PackagePluginConfig struct {
	UseExistingCluster        bool   `json:"use-existing-cluster"`
	Namespace                 string `json:"namespace"`
	PackageName               string `json:"package-name"`
	PackageVersion            string `json:"package-version"`
	PackageVersionUpdate      string `json:"package-version-update"`
	RegistryServer            string `json:"registry-server"`
	RegistryUsername          string `json:"registry-username"`
	RegistryPassword          string `json:"registry-password"`
	RepositoryName            string `json:"repository-name"`
	RepositoryURL             string `json:"repository-url"`
	RepositoryURLNoTag        string `json:"repository-url-no-tag"`
	RepositoryURLPrivate      string `json:"repository-url-private"`
	RepositoryURLPrivateNoTag string `json:"repository-url-private-no-tag"`
	RepositoryOriginalTag     string `json:"repository-original-tag"`
	RepositoryLatestTag       string `json:"repository-latest-tag"`
	RepositoryPrivateTag      string `json:"repository-private-tag"`
	ClusterNameMC             string `json:"mc-cluster-name"`
	ClusterNameWLC            string `json:"wlc-cluster-name"`
	KubeConfigPathMC          string `json:"mc-kubeconfig-Path"`
	KubeConfigPathWLC         string `json:"wlc-kubeconfig-Path"`
	WithValueFile             bool   `json:"with-value-file"`
}

type registrySecretOutput struct {
	Name      string `json:"name"`
	Registry  string `json:"registry"`
	Exported  string `json:"exported"`
	Age       string `json:"age"`
	Namespace string `json:"namespace"`
}

type repositoryOutput struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Status     string `json:"status"`
	Namespace  string `json:"namespace"`
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
	config                      = &PackagePluginConfig{}
	configPath                  string
	tkgCfgDir                   string
	err                         error
	packagePlugin               packagelib.PackagePlugin
	secretPlugin                secretlib.SecretPlugin
	result                      packagelib.PackagePluginResult
	resultImgPullSecret         secretlib.SecretPluginResult
	clusterCreationTimeout      = 30 * time.Minute
	pollInterval                = 20 * time.Second
	pollTimeout                 = 10 * time.Minute
	testImgPullSecretName       = "test-secret"
	testNamespace               = "test-ns"
	testRegistry                = "projects-stg.registry.vmware.com"
	testRegistryUsername        = "robot$tkgprivate+tkgprivate"
	testRegistryPassword        = "JUBFa3AW5SD6S6CBchFH7yNFIaF3LbVL"
	testRepoName                = "carvel-test"
	testRepoURL                 = "projects-stg.registry.vmware.com/tkg/test-packages/test-repo:v1.0.0"
	testRepoURLNoTag            = "projects-stg.registry.vmware.com/tkg/test-packages/test-repo"
	testRepoURLPrivate          = "projects-stg.registry.vmware.com/tkgprivate/example-pkg-repo@sha256:a80e9b512b9eff76ab638cce50a3c4541a12673d9b698103314f32c93f1deb61"
	testRepoURLPrivateNoTag     = "projects-stg.registry.vmware.com/tkgprivate/example-pkg-repo"
	testRepoOriginalTag         = "v1.0.0"
	testRepoLatestTag           = "v1.1.0"
	testRepoPrivateTag          = "sha256:a80e9b512b9eff76ab638cce50a3c4541a12673d9b698103314f32c93f1deb61"
	testPkgInstallName          = "test-pkg"
	testPkgName                 = "pkg.test.carvel.dev"
	testPkgVersion              = "2.0.0"
	testPkgVersionUpdate        = "3.0.0-rc.1"
	imgPullSecretOptions        tkgpackagedatamodel.RegistrySecretOptions
	pkgAvailableOptions         tkgpackagedatamodel.PackageAvailableOptions
	pkgOptions                  tkgpackagedatamodel.PackageOptions
	repoOptions                 tkgpackagedatamodel.RepositoryOptions
	expectedPkgOutput           packageInstalledOutput
	expectedPkgOutputUpdate     packageInstalledOutput
	expectedRepoOutput          repositoryOutput
	expectedRepoOutputLatestTag repositoryOutput
	expectedRepoOutputPrivate   repositoryOutput
	imgPullSecretOutput         []registrySecretOutput
	pkgOutput                   []packageInstalledOutput
	pkgAvailableOutput          []packageAvailableOutput
	repoOutput                  []repositoryOutput
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
				Edition:                "tkg",
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
				Edition:           "tkg",
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
			config.Namespace = testNamespace
		}

		if config.RegistryServer == "" {
			config.RegistryServer = testRegistry
		}

		if config.RegistryUsername == "" {
			config.RegistryUsername = testRegistryUsername
		}

		if config.RegistryPassword == "" {
			config.RegistryPassword = testRegistryPassword
		}

		if config.RepositoryName == "" {
			config.RepositoryName = testRepoName
		}

		if config.RepositoryURL == "" {
			config.RepositoryURL = testRepoURL
		}

		if config.RepositoryURLNoTag == "" {
			config.RepositoryURLNoTag = testRepoURLNoTag
		}

		if config.RepositoryURLPrivate == "" {
			config.RepositoryURLPrivate = testRepoURLPrivate
		}

		if config.RepositoryURLPrivateNoTag == "" {
			config.RepositoryURLPrivate = testRepoURLPrivateNoTag
		}

		if config.RepositoryOriginalTag == "" {
			config.RepositoryOriginalTag = testRepoOriginalTag
		}

		if config.RepositoryLatestTag == "" {
			config.RepositoryLatestTag = testRepoLatestTag
		}

		if config.RepositoryPrivateTag == "" {
			config.RepositoryPrivateTag = testRepoPrivateTag
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

		imgPullSecretOptions = tkgpackagedatamodel.RegistrySecretOptions{
			SecretName: testImgPullSecretName,
			Namespace:  constants.DefaultNamespace,
			SkipPrompt: true,
		}
		pkgOptions = tkgpackagedatamodel.PackageOptions{
			CreateNamespace: true,
			Namespace:       config.Namespace,
			PackageName:     config.PackageName,
			PkgInstallName:  testPkgInstallName,
			Version:         config.PackageVersion,
			SkipPrompt:      true,
			Wait:            true,
			PollInterval:    pollInterval,
			PollTimeout:     pollTimeout,
		}
		repoOptions = tkgpackagedatamodel.RepositoryOptions{
			CreateNamespace: true,
			Namespace:       config.Namespace,
			RepositoryName:  config.RepositoryName,
			SkipPrompt:      true,
			Wait:            true,
			PollInterval:    pollInterval,
			PollTimeout:     pollTimeout,
		}

		expectedRepoOutput = repositoryOutput{
			Name:       config.RepositoryName,
			Repository: config.RepositoryURLNoTag,
			Tag:        config.RepositoryOriginalTag,
			Status:     "Reconcile succeeded",
			Namespace:  config.Namespace,
		}

		expectedRepoOutputLatestTag = repositoryOutput{
			Name:       config.RepositoryName,
			Repository: config.RepositoryURLNoTag,
			Tag:        "(>0.0.0)",
			Status:     "Reconcile succeeded",
			Namespace:  config.Namespace,
		}

		expectedRepoOutputPrivate = repositoryOutput{
			Name:       config.RepositoryName,
			Repository: config.RepositoryURLPrivateNoTag,
			Tag:        config.RepositoryPrivateTag,
			Status:     "Reconcile succeeded",
			Namespace:  config.Namespace,
		}

		expectedPkgOutput = packageInstalledOutput{
			Name:           testPkgInstallName,
			PackageName:    config.PackageName,
			PackageVersion: config.PackageVersion,
			Status:         "Reconcile succeeded",
		}

		expectedPkgOutputUpdate = packageInstalledOutput{
			Name:           testPkgInstallName,
			PackageName:    config.PackageName,
			PackageVersion: config.PackageVersionUpdate,
			Status:         "Reconcile succeeded",
		}
	})

	Context("testing package plugin on management cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.KubeConfigPathMC, pollInterval, pollTimeout, "json", "", 0)
			secretPlugin = secretlib.NewSecretPlugin(config.KubeConfigPathMC, pollInterval, pollTimeout, "json", 0)
		})
		It("should pass all checks on management cluster", func() {
			cleanup()
			testHelper()
			log.Info("Successfully finished package plugin integration tests on management cluster")
		})
	})

	Context("testing package plugin on workload cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.KubeConfigPathWLC, pollInterval, pollTimeout, "json", "", 0)
			secretPlugin = secretlib.NewSecretPlugin(config.KubeConfigPathWLC, pollInterval, pollTimeout, "json", 0)
		})
		It("should pass all checks on workload cluster", func() {
			cleanup()
			testHelper()
			log.Info("Successfully finished package plugin integration tests on workload cluster")
		})
	})
})

func cleanup() {
	By("cleanup previous package installation")
	result = packagePlugin.DeleteInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("cleanup previous package repository installation")
	result = packagePlugin.DeleteRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("cleanup previous secret installation")
	resultImgPullSecret = secretPlugin.DeleteRegistrySecret(&imgPullSecretOptions)
	if resultImgPullSecret.Stderr != nil {
		Expect(resultImgPullSecret.Stderr.String()).ShouldNot(ContainSubstring(fmt.Sprintf(tkgpackagedatamodel.SecretGenAPINotAvailable, tkgpackagedatamodel.SecretGenGVR)))
	}
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())
}

func testHelper() {
	By("trying to update package repository with a private URL")
	repoOptions.RepositoryURL = config.RepositoryURLPrivate
	repoOptions.CreateRepository = true
	repoOptions.PollTimeout = 20 * time.Second
	result = packagePlugin.UpdateRepository(&repoOptions)
	Expect(result.Error).To(HaveOccurred())

	By("add registry secret")
	imgPullSecretOptions.Username = testRegistryUsername
	imgPullSecretOptions.PasswordInput = testRegistryPassword
	imgPullSecretOptions.Server = testRegistry
	resultImgPullSecret = secretPlugin.AddRegistrySecret(&imgPullSecretOptions)
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())

	By("update registry secret to export the secret from default namespace to all namespaces")
	t := true
	imgPullSecretOptions.Export = tkgpackagedatamodel.TypeBoolPtr{ExportToAllNamespaces: &t}
	resultImgPullSecret = secretPlugin.UpdateRegistrySecret(&imgPullSecretOptions)
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())

	By("list registry secret")
	resultImgPullSecret = secretPlugin.ListRegistrySecret(&imgPullSecretOptions)
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(resultImgPullSecret.Stdout.Bytes(), &imgPullSecretOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(imgPullSecretOutput)).To(BeNumerically(">=", 1))

	By("wait for the private package repository reconciliation")
	repoOptions.RepositoryURL = config.RepositoryURLPrivate
	repoOptions.CreateRepository = true
	repoOptions.PollInterval = pollInterval
	repoOptions.PollTimeout = pollTimeout
	result = packagePlugin.CheckRepositoryAvailable(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("list package repository")
	repoOptions.AllNamespaces = true
	result = packagePlugin.ListRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &repoOutput)
	Expect(err).ToNot(HaveOccurred())

	By("get package repository")
	repoOutput = []repositoryOutput{{Namespace: testNamespace}}
	result = packagePlugin.GetRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &repoOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(repoOutput)).To(BeNumerically("==", 1))
	Expect(repoOutput[0]).To(Equal(expectedRepoOutputPrivate))

	By("update package repository with a new URL without tag")
	repoOptions.RepositoryURL = config.RepositoryURLNoTag
	repoOptions.CreateRepository = true
	repoOptions.CreateNamespace = true
	repoOptions.PollInterval = pollInterval
	repoOptions.PollTimeout = pollTimeout
	result = packagePlugin.UpdateRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("wait for package repository reconciliation")
	result = packagePlugin.CheckRepositoryAvailable(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package repository")
	repoOutput = []repositoryOutput{{Namespace: testNamespace}}
	result = packagePlugin.GetRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())
	err = json.Unmarshal(result.Stdout.Bytes(), &repoOutput)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(repoOutput)).To(BeNumerically("==", 1))
	Expect(repoOutput[0]).To(Equal(expectedRepoOutputLatestTag))

	By("update package repository with a new URL")
	repoOptions.RepositoryURL = config.RepositoryURL
	result = packagePlugin.UpdateRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("wait for package repository reconciliation")
	result = packagePlugin.CheckRepositoryAvailable(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("get package repository")
	repoOutput = []repositoryOutput{{Namespace: testNamespace}}
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
	pkgOptions.PollInterval = pollInterval
	pkgOptions.PollTimeout = pollTimeout
	pkgOptions.Version = config.PackageVersion
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
	Expect(pkgOutput[0]).To(Equal(expectedPkgOutputUpdate))

	By("delete package install")
	result = packagePlugin.DeleteInstalledPackage(&pkgOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("delete package repository")
	repoOptions.PollInterval = pollInterval
	result = packagePlugin.DeleteRepository(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("wait for package repository deletion")
	result = packagePlugin.CheckRepositoryDeleted(&repoOptions)
	Expect(result.Error).ToNot(HaveOccurred())

	By("delete registry secret")
	resultImgPullSecret = secretPlugin.DeleteRegistrySecret(&imgPullSecretOptions)
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())
}

func TestTkgpackageclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkgpackageclient Suite")
}
