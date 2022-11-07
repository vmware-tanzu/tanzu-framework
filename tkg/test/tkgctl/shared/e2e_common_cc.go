// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/sets"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/tkg/test/framework/exec"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type E2ECommonCCSpecInput struct {
	E2EConfig             *framework.E2EConfig
	ArtifactsFolder       string
	Cni                   string
	Plan                  string
	Namespace             string
	IsCustomCB            bool
	DoUpgrade             bool
	CheckAdmissionWebhook bool
}

func E2ECommonCCSpec(ctx context.Context, inputGetter func() E2ECommonCCSpecInput) { //nolint:funlen
	var (
		err                             error
		input                           E2ECommonCCSpecInput
		tkgCtlClient                    tkgctl.TKGClient
		logsDir                         string
		clusterName                     string
		namespace                       string
		mngKubeConfigFileName           string
		mngKubeConfigFile               string
		mngProxy                        *framework.ClusterProxy
		mngContextName                  string
		options                         framework.CreateClusterOptions
		clusterConfigFile               string
		tkrVersionsSet                  sets.StringSet
		oldTKR                          *runv1alpha3.TanzuKubernetesRelease
		defaultTKR                      *runv1alpha3.TanzuKubernetesRelease
		mngClient                       client.Client
		clusterResources                []ClusterResource
		mngDynamicClient                dynamic.Interface
		mngAggregatedAPIResourcesClient client.Client
		mngDiscoveryClient              discovery.DiscoveryInterface
		admissionRegistrationClient     admissionregistrationv1.AdmissionregistrationV1Interface
		infrastructureName              string
		wlcClient                       client.Client
	)

	BeforeEach(func() { //nolint:dupl
		namespace = constants.DefaultNamespace
		input = inputGetter()
		if input.Namespace != "" {
			namespace = input.Namespace
		}

		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc-" + util.RandomString(4) // nolint:gomnd

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})
		Expect(err).To(BeNil())

		mngContextName = input.E2EConfig.ManagementClusterName + "-admin@" + input.E2EConfig.ManagementClusterName
		mngProxy = framework.NewClusterProxy(input.E2EConfig.ManagementClusterName, "", mngContextName)
		mngKubeConfigFileName = input.E2EConfig.ManagementClusterName + ".kubeconfig"
		mngKubeConfigFile = filepath.Join(os.TempDir(), mngKubeConfigFileName)

		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: input.E2EConfig.ManagementClusterName,
			Namespace:   "tkg-system",
			ExportFile:  mngKubeConfigFile,
		})
		Expect(err).To(BeNil())

		options = framework.CreateClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			Plan:        "dev",
			CniType:     input.Cni,
		}

		if input.Plan != "" {
			options.Plan = input.Plan
		}

		if input.E2EConfig.InfrastructureName == "vsphere" {
			if input.Cni == "antrea" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_ANTREA"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
			if input.Cni == "calico" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_CALICO"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
		}
		clusterConfigFile, err = framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		pacificCluster, err := tkgCtlClient.IsPacificRegionalCluster()
		Expect(err).NotTo(HaveOccurred())
		if pacificCluster {
			infrastructureName = "TKGS"
		} else {
			infrastructureName = input.E2EConfig.InfrastructureName
		}

		By(fmt.Sprintf("Get k8s client for management cluster %q", input.E2EConfig.ManagementClusterName))
		mngClient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, admissionRegistrationClient, err = GetClients(ctx, mngKubeConfigFile)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By(fmt.Sprintf("Deleting workload cluster %q", clusterName))
		err = tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())

		// verify addon resources are deleted successfully
		By(fmt.Sprintf("Verify workload cluster %q resources have been deleted", clusterName))
		Eventually(func() bool {
			return clusterResourcesDeleted(ctx, mngClient, clusterResources)
		}, resourceDeletionWaitTimeout, pollingInterval).Should(BeTrue())

		os.Remove(clusterConfigFile)
		os.Remove(mngKubeConfigFile)

		By("Test successful !")
	})

	It("Should verify basic cluster lifecycle operations", func() {
		By(fmt.Sprintf("Generating workload cluster configuration for cluster %q", clusterName))

		err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
		})
		Expect(err).To(BeNil())

		tkrVersionsSet, oldTKR, defaultTKR = getAvailableTKRs(ctx, mngProxy, input.E2EConfig.TkgConfigDir)

		if input.IsCustomCB {
			// in case of customCB, a new configuration should be generated
			// Due to current CB limitation. enable sc will generate a new CB which overwrite below kubectl applied one,
			// need antrea config guru polish current e2e to support sc
			options.OtherConfigs = map[string]string{
				"ENABLE_DEFAULT_STORAGE_CLASS": "false",
			}
			clusterConfigFile, err = framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
			err = exec.KubectlApplyWithArgs(ctx, mngKubeConfigFile, getCustomCBResourceFile(clusterName, namespace, defaultTKR.Name))
			Expect(err).To(BeNil())
		}

		By(fmt.Sprintf("Creating a workload cluster %q with TKR %q", clusterName, oldTKR.Spec.Version))
		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
			TkrVersion:        oldTKR.Spec.Version,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		wlcKubeConfigFileName := clusterName + ".kubeconfig"
		wlcKubeConfigFile := filepath.Join(os.TempDir(), wlcKubeConfigFileName)
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			ExportFile:  wlcKubeConfigFile,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, wlcKubeConfigFile, ""), 2)

		By(fmt.Sprintf("Get k8s client for workload cluster %q", clusterName))
		wlcClient, _, _, _, _, err = GetClients(ctx, wlcKubeConfigFile)
		Expect(err).NotTo(HaveOccurred())

		// verify addons are deployed successfully
		By(fmt.Sprintf("Verify addon packages on management cluster %q matches clusterBootstrap info on management cluster %q", input.E2EConfig.ManagementClusterName, input.E2EConfig.ManagementClusterName))
		err = CheckClusterCB(ctx, mngClient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, "", "", infrastructureName, true, input.IsCustomCB)
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Verify addon packages on workload cluster %q match clusterBootstrap info on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
		err = CheckClusterCB(ctx, mngClient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, clusterName, namespace, infrastructureName, false, input.IsCustomCB)
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Get management cluster resources created by addons-manager for workload cluster %q on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
		clusterResources, err = GetManagementClusterResources(ctx, mngClient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, namespace, clusterName, infrastructureName)
		Expect(err).NotTo(HaveOccurred())

		By(fmt.Sprintf("Validating the kubernetes version after cluster %q is created", clusterName))
		validateKubernetesVersion(clusterName, oldTKR.Spec.Kubernetes.Version, wlcKubeConfigFile)

		By(fmt.Sprintf("Validating the TKR data after cluster %q is created", clusterName))
		verifyTKRData(ctx, mngProxy, options.ClusterName, options.Namespace)

		if input.DoUpgrade {
			By(fmt.Sprintf("Validating the 'updatesAvailable' condition is true and lists upgradable TKR version"))
			validateUpdatesAvailableCondition(ctx, mngProxy, options.ClusterName, options.Namespace, tkrVersionsSet)

			By(fmt.Sprintf("Upgrading workload cluster %q with TKR %q", clusterName, defaultTKR.Spec.Version))
			err = tkgCtlClient.UpgradeCluster(tkgctl.UpgradeClusterOptions{
				ClusterName: clusterName,
				Namespace:   namespace,
				TkrVersion:  defaultTKR.Spec.Version,
				SkipPrompt:  true,
			})
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Validating the kubernetes version after cluster %q is upgraded", clusterName))
			validateKubernetesVersion(clusterName, defaultTKR.Spec.Kubernetes.Version, wlcKubeConfigFile)

			By(fmt.Sprintf("Validating the TKR data after cluster %q is upgraded", clusterName))
			verifyTKRData(ctx, mngProxy, options.ClusterName, options.Namespace)

			// verify addons are deployed successfully after cluster upgrade
			By(fmt.Sprintf("Verify addon packages on workload cluster %q match clusterBootstrap info on management cluster %q after cluster upgrade", clusterName, input.E2EConfig.ManagementClusterName))
			err = CheckClusterCB(ctx, mngClient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, clusterName, namespace, infrastructureName, false, input.IsCustomCB)
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Get management cluster resources created by addons-manager for workload cluster %q on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
			clusterResources, err = GetManagementClusterResources(ctx, mngClient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, namespace, clusterName, infrastructureName)
			Expect(err).NotTo(HaveOccurred())
		}

		if input.CheckAdmissionWebhook {
			checkAdmissionWebhooks(ctx, mngClient, admissionRegistrationClient)
		}
	})
}

// getCustomCBResourceFile return a manifest containing custom ClusterBootstrap and AntreaConfig
func getCustomCBResourceFile(clusterName, namespace string, tkrName string) []byte {
	return []byte(fmt.Sprintf(customAntreaConfigAndCBResource, clusterName, namespace, "kube-system", clusterName, namespace, tkrName, clusterName, namespace, clusterName, clusterName))
}

func getAvailableTKRs(ctx context.Context, mcProxy *framework.ClusterProxy, tkgConfigDir string) (sets.StringSet, *runv1alpha3.TanzuKubernetesRelease, *runv1alpha3.TanzuKubernetesRelease) {
	var (
		tkrs               []*runv1alpha3.TanzuKubernetesRelease
		defaultTKR, oldTKR *runv1alpha3.TanzuKubernetesRelease
	)

	tkgBOMConfigClient := tkgconfigbom.New(tkgConfigDir, nil)
	defaultTKRVersion, err := tkgBOMConfigClient.GetDefaultTKRVersion()
	Expect(err).ToNot(HaveOccurred(), "failed to get the default TKR version")

	Eventually(func() bool {
		tkrs = mcProxy.GetTKRs(ctx)
		defaultTKR, oldTKR = getTKRsForUpgrade(defaultTKRVersion, tkrs)
		return defaultTKR != nil && oldTKR != nil
	}, waitTimeout, pollingInterval).Should(BeTrue(), "failed to get at least 2 TKRs(upgradable) to perform upgrade tests")
	tkrVersions := getTKRVersions(tkrs)

	return tkrVersions, oldTKR, defaultTKR
}

func checkAdmissionWebhooks(ctx context.Context, mngClient client.Client, admissionRegistrationClient admissionregistrationv1.AdmissionregistrationV1Interface) {
	caCert, err := getWebhookCACert(ctx, mngClient, webhookScrtName, systemNamespace)
	Expect(err).NotTo(HaveOccurred())
	Expect(caCert).NotTo(BeEmpty())

	labelMatch, err := labels.NewRequirement(webhookLabelKey, selection.Equals, []string{webhookLabelValue})
	Expect(err).ToNot(HaveOccurred())
	webhookSelectorString := labels.NewSelector().Add(*labelMatch).String()

	caCertExists, err := verifyCABundleInLabeledWebhooks(ctx, admissionRegistrationClient, webhookSelectorString, caCert)
	Expect(err).ToNot(HaveOccurred())
	Expect(caCertExists).To(BeTrue())
}
