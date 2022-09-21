// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	packagelib "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/package/test/lib"
	secretlib "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/secret/test/lib"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
	"github.com/vmware-tanzu/tanzu-framework/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/tkg/test/framework/exec"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type PackagePluginConfig struct {
	UseExistingCluster        bool   `json:"use-existing-cluster"`
	Namespace                 string `json:"namespace"`
	PackageName               string `json:"package-name"`
	PackageVersion            string `json:"package-version"`
	PackageVersionUpdate      string `json:"package-version-update"`
	PrivateRegistryServer     string `json:"private-registry-server"`
	PrivateRegistryUsername   string `json:"private-registry-username"`
	PrivateRegistryPassword   string `json:"private-registry-password"`
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
	ClusterNamespaceWLC       string `json:"wlc-cluster-namespace"`
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
	privateTestRegistry         = "registry-svc.registry.svc.cluster.local:443"
	privateTestRegistryUsername = "testuser"
	privateTestRegistryPassword = "testpassword"
	testRepoName                = "carvel-test"
	testRepoURL                 = "projects-stg.registry.vmware.com/tkg/test-packages/test-repo:v1.0.0"
	testRepoURLNoTag            = "projects-stg.registry.vmware.com/tkg/test-packages/test-repo"
	testRepoURLPrivate          = "registry-svc.registry.svc.cluster.local:443/secret-test/test-repo@sha256:e07483e2140fa427d9875aee9055d72efc49a732f3a3fb2c9651d9f39159315a"
	testRepoURLPrivateNoTag     = "registry-svc.registry.svc.cluster.local:443/secret-test/test-repo"
	testRepoOriginalTag         = "v1.0.0"
	testRepoLatestTag           = "v1.1.0"
	testRepoPrivateTag          = "sha256:e07483e2140fa427d9875aee9055d72efc49a732f3a3fb2c9651d9f39159315a"
	testPkgInstallName          = "test-pkg"
	testPkgName                 = "pkg.test.carvel.dev"
	testPkgVersion              = "2.0.0"
	testPkgVersionUpdate        = "3.0.0-rc.1"
	imgPullSecretOptions        packagedatamodel.RegistrySecretOptions
	pkgAvailableOptions         packagedatamodel.PackageAvailableOptions
	pkgOptions                  packagedatamodel.PackageOptions
	repoOptions                 packagedatamodel.RepositoryOptions
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

		configData, err := os.ReadFile(configPath)
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
			config.ClusterNameMC = fmt.Sprintf(packagedatamodel.DefaultClusterPrefix + "mc-test-pkg-plugin")
		}

		if config.ClusterNameWLC == "" {
			config.ClusterNameWLC = fmt.Sprintf(packagedatamodel.DefaultClusterPrefix + "wlc-test-pkg-plugin")
		}

		if config.ClusterNamespaceWLC == "" {
			config.ClusterNamespaceWLC = "default"
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
					Verbosity: packagedatamodel.DefaultLogLevel,
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
					Verbosity: packagedatamodel.DefaultLogLevel,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			options := framework.CreateClusterOptions{
				ClusterName: config.ClusterNameWLC,
				Namespace:   packagedatamodel.DefaultNamespace,
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
				Namespace:   packagedatamodel.DefaultNamespace,
				ExportFile:  config.KubeConfigPathWLC,
			})
			Expect(err).NotTo(HaveOccurred())

			log.Info("Finished creating management and workload clusters")
		}

		if config.Namespace == "" {
			config.Namespace = testNamespace
		}

		if config.PrivateRegistryServer == "" {
			config.PrivateRegistryServer = privateTestRegistry
		}

		if config.PrivateRegistryUsername == "" {
			config.PrivateRegistryUsername = privateTestRegistryUsername
		}

		if config.PrivateRegistryPassword == "" {
			config.PrivateRegistryPassword = privateTestRegistryPassword
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

		pkgAvailableOptions = packagedatamodel.PackageAvailableOptions{
			Namespace: config.Namespace,
		}

		imgPullSecretOptions = packagedatamodel.RegistrySecretOptions{
			SecretName: testImgPullSecretName,
			Namespace:  packagedatamodel.DefaultNamespace,
			SkipPrompt: true,
		}
		pkgOptions = packagedatamodel.PackageOptions{
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
		repoOptions = packagedatamodel.RepositoryOptions{
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
			setUpPrivateRegistry(config.KubeConfigPathMC, config.ClusterNameMC)
			testHelper()
			log.Info("Successfully finished package plugin integration tests on management cluster")
		})
	})

	Context("testing package plugin on workload cluster", func() {
		BeforeEach(func() {
			packagePlugin = packagelib.NewPackagePlugin(config.KubeConfigPathWLC, pollInterval, pollTimeout, "json", "", 0)
			secretPlugin = secretlib.NewSecretPlugin(config.KubeConfigPathWLC, pollInterval, pollTimeout, "json", 0)
		})
		AfterEach(func() {
			paused := false
			pauseKappControllerPackage(config.ClusterNameWLC, config.ClusterNamespaceWLC, config.KubeConfigPathMC, config.ClusterNameMC, paused)
		})
		It("should pass all checks on workload cluster", func() {
			cleanup()
			paused := true
			pauseKappControllerPackage(config.ClusterNameWLC, config.ClusterNamespaceWLC, config.KubeConfigPathMC, config.ClusterNameMC, paused)
			setUpPrivateRegistry(config.KubeConfigPathWLC, config.ClusterNameWLC)
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
		Expect(resultImgPullSecret.Stderr.String()).ShouldNot(ContainSubstring(fmt.Sprintf(packagedatamodel.SecretGenAPINotAvailable, packagedatamodel.SecretGenAPIName, packagedatamodel.SecretGenAPIVersion)))
	}
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())
}

func setUpPrivateRegistry(kubeconfigPath, clusterName string) {
	By("set up private registry in cluster")
	kubeCtx := clusterName + "-admin@" + clusterName
	registryNamespace := "registry"

	command := exec.NewCommand(
		exec.WithCommand("kubectl"),
		exec.WithArgs("patch", "configmap", "kapp-controller-config", "-n", "tkg-system", "--type", "merge", "-p", fmt.Sprintf("{\"data\":{\"dangerousSkipTLSVerify\":\"registry-svc.%s.svc.cluster.local\"}}", registryNamespace), "--context", kubeCtx, "--kubeconfig", kubeconfigPath),
		exec.WithStdout(GinkgoWriter),
	)
	err := command.RunAndRedirectOutput(context.Background())
	Expect(err).ToNot(HaveOccurred())

	command = exec.NewCommand(
		exec.WithCommand("kubectl"),
		exec.WithArgs("apply", "-f", "config/assets/registry-namespace.yml", "--context", kubeCtx, "--kubeconfig", kubeconfigPath),
		exec.WithStdout(GinkgoWriter),
	)
	err = command.RunAndRedirectOutput(context.Background())
	Expect(err).ToNot(HaveOccurred())

	backOff := wait.Backoff{
		Steps:    10,
		Duration: 15 * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}

	err = retry.OnError(
		backOff,
		func(err error) bool {
			return err != nil
		},
		func() error {
			command = exec.NewCommand(
				exec.WithCommand("kubectl"),
				exec.WithArgs("apply", "-f", "config/assets/registry-contents.yml", "--context", kubeCtx, "--kubeconfig", kubeconfigPath),
				exec.WithStdout(GinkgoWriter),
			)
			return command.RunAndRedirectOutput(context.Background())
		})
	Expect(err).ToNot(HaveOccurred())

	command = exec.NewCommand(
		exec.WithCommand("kubectl"),
		exec.WithArgs("apply", "-f", "config/assets/htpasswd-auth.yml", "--context", kubeCtx, "--kubeconfig", kubeconfigPath),
		exec.WithStdout(GinkgoWriter),
	)
	err = command.RunAndRedirectOutput(context.Background())
	Expect(err).ToNot(HaveOccurred())

	command = exec.NewCommand(
		exec.WithCommand("kubectl"),
		exec.WithArgs("apply", "-f", "config/assets/certs-for-skip-tls.yml", "--context", kubeCtx, "--kubeconfig", kubeconfigPath),
		exec.WithStdout(GinkgoWriter),
	)
	err = command.RunAndRedirectOutput(context.Background())
	Expect(err).ToNot(HaveOccurred())

	command = exec.NewCommand(
		exec.WithCommand("kubectl"),
		exec.WithArgs("apply", "-f", "config/assets/registry.yml", "--context", kubeCtx, "--kubeconfig", kubeconfigPath),
		exec.WithStdout(GinkgoWriter),
	)
	err = command.RunAndRedirectOutput(context.Background())
	Expect(err).ToNot(HaveOccurred())

	By("restart kapp-controller pod to pick up configmap changes")
	proxy := framework.NewClusterProxy(clusterName, kubeconfigPath, kubeCtx)
	clientSet := proxy.GetClientSet()

	var podList *corev1.PodList
	kappNamespace := "tkg-system"
	podList, err = clientSet.CoreV1().Pods(kappNamespace).List(context.Background(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())

	for i := range podList.Items {
		pod := &podList.Items[i]
		if strings.Contains(pod.Name, "kapp-controller") {
			err = clientSet.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
		}
	}

	By("make sure registry pod is running")
	err = retry.OnError(
		backOff,
		func(err error) bool {
			return err != nil
		},
		func() error {
			podList, err := clientSet.CoreV1().Pods(registryNamespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			for _, pod := range podList.Items {
				if pod.Status.Phase != corev1.PodRunning {
					return errors.New("registry pod is not running")
				}
			}
			return nil
		})
	Expect(err).ToNot(HaveOccurred())

	By("make sure package API is available after kapp-controller restart")
	err = retry.OnError(
		backOff,
		func(err error) bool {
			return err != nil
		},
		func() error {
			result = packagePlugin.ListRepository(&repoOptions)
			if result.Error != nil {
				return result.Error
			}
			return nil
		})
	Expect(err).ToNot(HaveOccurred())
}

func pauseKappControllerPackage(clusterName, clusterNamespace, mgmtClusterKubeconfigPath, mgmtClusterName string, pause bool) {
	By("pausing kapp controller app")
	mgmtClusterKubeCtx := mgmtClusterName + "-admin@" + mgmtClusterName

	command := exec.NewCommand(
		exec.WithCommand("kubectl"),
		exec.WithArgs("patch", "app", fmt.Sprintf("%s-kapp-controller", clusterName), "-n", clusterNamespace, "--type", "merge", "-p", fmt.Sprintf("{\"spec\":{\"paused\":%t}}", pause), "--context", mgmtClusterKubeCtx, "--kubeconfig", mgmtClusterKubeconfigPath),
		exec.WithStdout(GinkgoWriter),
	)
	err := command.RunAndRedirectOutput(context.Background())
	Expect(err).ToNot(HaveOccurred())
}

func testHelper() {

	By("trying to update package repository with a private URL")
	repoOptions.RepositoryURL = config.RepositoryURLPrivate
	repoOptions.CreateRepository = true
	repoOptions.PollTimeout = 30 * time.Second
	result = packagePlugin.UpdateRepository(&repoOptions)
	Expect(result.Error).To(HaveOccurred())

	By("add registry secret")
	imgPullSecretOptions.Username = privateTestRegistryUsername
	imgPullSecretOptions.PasswordInput = privateTestRegistryPassword
	imgPullSecretOptions.Server = privateTestRegistry
	resultImgPullSecret = secretPlugin.AddRegistrySecret(&imgPullSecretOptions)
	Expect(resultImgPullSecret.Error).ToNot(HaveOccurred())

	By("update registry secret to export the secret from default namespace to all namespaces")
	t := true
	imgPullSecretOptions.Export = packagedatamodel.TypeBoolPtr{ExportToAllNamespaces: &t}
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

func TestE2EPackageClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E PackageClient Suite")
}
