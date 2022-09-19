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
	"sort"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/sets"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/util/resolution"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/util/topology"
	clusterdata "github.com/vmware-tanzu/tanzu-framework/tkg/tkr/webhook/cluster/tkr-resolver/cluster"
)

type E2ETKRResolverValidationForClusterCRUDSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
	Plan            string
	Namespace       string
}

func E2ETKRResolverValidationForClusterCRUDSpec(context context.Context, inputGetter func() E2ETKRResolverValidationForClusterCRUDSpecInput) { //nolint:funlen
	var (
		err                             error
		input                           E2ETKRResolverValidationForClusterCRUDSpecInput
		tkgCtlClient                    tkgctl.TKGClient
		logsDir                         string
		clusterName                     string
		namespace                       string
		mcProxy                         *framework.ClusterProxy
		mcContextName                   string
		options                         framework.CreateClusterOptions
		clusterConfigFile               string
		tkrVersionsSet                  sets.StringSet
		oldTKR                          *runv1.TanzuKubernetesRelease
		defaultTKR                      *runv1.TanzuKubernetesRelease
		mngClient                       client.Client
		clusterResources                []ClusterResource
		infrastructureName              string
		mngDynamicClient                dynamic.Interface
		mngAggregatedAPIResourcesClient client.Client
		mngDiscoveryClient              discovery.DiscoveryInterface
		wlcClient                       client.Client
	)
	const (
		waitTimeout     = time.Minute * 15
		pollingInterval = time.Second * 10
	)

	BeforeEach(func() { //nolint:dupl
		namespace = constants.DefaultNamespace
		input = inputGetter()
		if input.Namespace != "" {
			namespace = input.Namespace
		}

		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName = mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)

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
		mngkubeConfigFileName := input.E2EConfig.ManagementClusterName + ".kubeconfig"
		mngtempFilePath := filepath.Join(os.TempDir(), mngkubeConfigFileName)
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: input.E2EConfig.ManagementClusterName,
			Namespace:   "tkg-system",
			ExportFile:  mngtempFilePath,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Get k8s client for management cluster %q", clusterName))
		mngClient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, err = GetClients(context, mngtempFilePath)
		Expect(err).NotTo(HaveOccurred())

	})
	AfterEach(func() {
		os.Remove(clusterConfigFile)
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
			return clusterResourcesDeleted(context, mngClient, clusterResources)
		}, resourceDeletionWaitTimeout, pollingInterval).Should(BeTrue())
	})

	It("Should create workload cluster with default kubernetes version and verify infra machine images are resolved correctly", func() {
		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))

		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Validating the TKR data after cluster %q is created", clusterName))
		verifyTKRData(context, mcProxy, options.ClusterName, options.Namespace)

		By("Test successful !")
	})

	It("Should upgrade the cluster and verify infra machine images are resolved correctly", func() {
		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))

		tkrVersionsSet, oldTKR, defaultTKR = getAvailableTKRs(context, mcProxy, input.E2EConfig.TkgConfigDir)

		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
			TkrVersion:        oldTKR.Spec.Version,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		kubeConfigFileName := clusterName + ".kubeconfig"
		tempKubeConfigFilePath := filepath.Join(os.TempDir(), kubeConfigFileName)
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			ExportFile:  tempKubeConfigFilePath,
		})
		Expect(err).To(BeNil())

		// validate k8s version of workload cluster
		By(fmt.Sprintf("Validating the kubernetes version after cluster %q is created", clusterName))
		validateKubernetesVersion(clusterName, oldTKR.Spec.Kubernetes.Version, tempKubeConfigFilePath)

		By(fmt.Sprintf("Validating the TKR data after cluster %q is created", clusterName))
		verifyTKRData(context, mcProxy, options.ClusterName, options.Namespace)

		By(fmt.Sprintf("Validating the 'updatesAvailable' condition is true and lists upgradable TKR version"))
		validateUpdatesAvailableCondition(context, mcProxy, options.ClusterName, options.Namespace, tkrVersionsSet)

		By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, tempKubeConfigFilePath, ""), 2)

		By(fmt.Sprintf("Get k8s client for workload cluster %q", clusterName))
		wlcClient, _, _, _, err = GetClients(context, tempKubeConfigFilePath)
		Expect(err).NotTo(HaveOccurred())

		// verify addons are deployed successfully
		By(fmt.Sprintf("Verify addon packages on management cluster %q matches clusterBootstrap info on management cluster %q", input.E2EConfig.ManagementClusterName, input.E2EConfig.ManagementClusterName))
		err = CheckClusterCB(context, mngClient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, "", "", infrastructureName, true, false)
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Verify addon packages on workload cluster %q matches clusterBootstrap info on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
		err = CheckClusterCB(context, mngClient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, clusterName, namespace, infrastructureName, false, false)
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Get management cluster resources created by addons-manager for workload cluster %q on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
		clusterResources, err = GetManagementClusterResources(context, mngClient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, namespace, clusterName, infrastructureName)
		Expect(err).NotTo(HaveOccurred())

		input.E2EConfig.TkrVersion = defaultTKR.Spec.Version
		Expect(input.E2EConfig.TkrVersion).ToNot(BeEmpty(), "config variable 'kubernetes_version' not set")
		By(fmt.Sprintf("Upgrading workload cluster %q to k8s version %q", clusterName, input.E2EConfig.TkrVersion))
		err = tkgCtlClient.UpgradeCluster(tkgctl.UpgradeClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			TkrVersion:  input.E2EConfig.TkrVersion,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Validating the kubernetes version after cluster %q is upgraded", clusterName))
		validateKubernetesVersion(clusterName, defaultTKR.Spec.Kubernetes.Version, tempKubeConfigFilePath)

		By(fmt.Sprintf("Validating the TKR data after cluster %q is upgraded", clusterName))
		verifyTKRData(context, mcProxy, options.ClusterName, options.Namespace)

		// verify addons are deployed successfully after cluster upgrade
		By(fmt.Sprintf("Verify addon packages on workload cluster %q match clusterBootstrap info on management cluster %q after cluster upgrade", clusterName, input.E2EConfig.ManagementClusterName))
		err = CheckClusterCB(context, mngClient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, clusterName, namespace, infrastructureName, false, false)
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Get management cluster resources created by addons-manager for workload cluster %q on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
		clusterResources, err = GetManagementClusterResources(context, mngClient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, namespace, clusterName, infrastructureName)
		Expect(err).NotTo(HaveOccurred())

		By("Test successful !")
	})
}

func validateUpdatesAvailableCondition(ctx context.Context, mcProxy *framework.ClusterProxy, clusterName, namespace string, tkrVersions sets.StringSet) {
	cluster := mcProxy.GetCluster(ctx, clusterName, namespace)
	Expect(cluster).ToNot(BeNil(), fmt.Sprintf("failed to get cluster '%s' in namespace: '%s'", clusterName, namespace))
	Expect(conditions.IsTrue(cluster, runv1.ConditionUpdatesAvailable)).To(BeTrue())

	updatesSet := topology.AvailableUpgrades(cluster)
	Expect(len(updatesSet)).ToNot(BeZero(), "Cluster 'updatesAvailable condition should list at least one TKR version ")

	updates := updatesSet.Slice()
	for i := range updates {
		Expect(tkrVersions.Has(updates[i])).To(BeTrue(), fmt.Sprintf("cluster's available update version '%s' doesn't match with available TKR versions :%v", updates[i], tkrVersions))
	}

}

func verifyTKRData(ctx context.Context, mcProxy *framework.ClusterProxy, clusterName, namespace string) {
	By("Initializing the TKR resolver")
	tkrResolver := initializeTKRResolver(ctx, mcProxy)
	cluster := mcProxy.GetCluster(ctx, clusterName, namespace)
	Expect(cluster).ToNot(BeNil(), fmt.Sprintf("failed to get cluster '%s' in namespace: '%s'", clusterName, namespace))
	clusterClass := mcProxy.GetClusterClass(ctx, cluster.Spec.Topology.Class, namespace)
	Expect(clusterClass).ToNot(BeNil(), fmt.Sprintf("failed to get clusterclass '%s' in namespace: '%s'", cluster.Spec.Topology.Class, namespace))
	query, err := resolution.ConstructQuery(cluster.Spec.Topology.Version, cluster, clusterClass)
	Expect(err).ToNot(HaveOccurred())
	result := tkrResolver.Resolve(*query)
	Expect(cluster.Labels[runv1.LabelTKR]).To(Equal(result.ControlPlane.TKRName),
		fmt.Sprintf("cluster's TKR label:'%s' doesn't match with resolved TKR label '%s'", cluster.Labels[runv1.LabelTKR], result.ControlPlane.TKRName))
	validateControlPlanceTKRDataFields(cluster, result)

	for i := range cluster.Spec.Topology.Workers.MachineDeployments {
		validateMDTKRDataFields(cluster, i, result)
	}

}

func initializeTKRResolver(context context.Context, mcProxy *framework.ClusterProxy) resolver.CachingResolver {
	tkrResolver := resolver.New()
	tkrs := mcProxy.GetTKRs(context)
	Expect(len(tkrs)).ToNot(Equal(0), "TKRs are not available in the management cluster")
	for i := range tkrs {
		tkrResolver.Add(tkrs[i])
	}

	osImages := mcProxy.GetOSImages(context)
	Expect(len(osImages)).ToNot(Equal(0), "OSImage resources are not available in the management cluster")
	for i := range osImages {
		tkrResolver.Add(osImages[i])
	}
	return tkrResolver
}

func validateControlPlanceTKRDataFields(cluster *clusterv1.Cluster, result data.Result) {
	var tkrData clusterdata.TKRData
	err := topology.GetVariable(cluster, "TKR_DATA", &tkrData)
	Expect(err).To(BeNil(), "failed to get the TKR_DATA variable for cluster")

	osImageResolved := tkrData[cluster.Spec.Topology.Version].Labels[runv1.LabelOSImage]
	Expect(osImageResolved).ToNot(BeEmpty(), fmt.Sprintf("cluster TKR_DATA[%s] is missing label:'%s' ", cluster.Spec.Topology.Version, runv1.LabelOSImage))
	k8sVersionResolved := cluster.Spec.Topology.Version
	tkrNameResolved := tkrData[cluster.Spec.Topology.Version].Labels[runv1.LabelTKR]
	Expect(tkrNameResolved).ToNot(BeEmpty(), fmt.Sprintf("cluster's TKR_DATA[%s] is missing label: '%s'", cluster.Spec.Topology.Version, runv1.LabelTKR))
	osImageRefFromTKRDATA := tkrData[k8sVersionResolved].OSImageRef
	osImageRefFromResolvedResult := result.ControlPlane.OSImagesByTKR[tkrNameResolved][osImageResolved].Spec.Image.Ref
	Expect(osImageRefFromTKRDATA).To(BeEquivalentTo(osImageRefFromResolvedResult),
		fmt.Sprintf("cluster's TKR_DATA[%s]'s osImgeRef doesn't match with osImageRef of the OSImage CR : '%s'", cluster.Spec.Topology.Version, tkrNameResolved))
}

func validateMDTKRDataFields(cluster *clusterv1.Cluster, mdIndex int, result data.Result) {
	var tkrData clusterdata.TKRData
	err := topology.GetMDVariable(cluster, mdIndex, "TKR_DATA", &tkrData)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to get the TKR_DATA variable for MD[%v]", mdIndex))

	osImageResolved := tkrData[cluster.Spec.Topology.Version].Labels[runv1.LabelOSImage]
	Expect(osImageResolved).ToNot(BeEmpty(), fmt.Sprintf("MD[%d]'s TKR_DATA[%s] is missing label:'%s' ", mdIndex, cluster.Spec.Topology.Version, runv1.LabelOSImage))

	k8sVersionResolved := cluster.Spec.Topology.Version
	tkrNameResolved := tkrData[cluster.Spec.Topology.Version].Labels[runv1.LabelTKR]
	Expect(tkrNameResolved).ToNot(BeEmpty(), fmt.Sprintf("MD[%d]'s TKR_DATA[%s] is missing label: '%s'", mdIndex, cluster.Spec.Topology.Version, runv1.LabelTKR))

	osImageRefFromTKRDATA := tkrData[k8sVersionResolved].OSImageRef
	osImageRefFromResolvedResult := result.MachineDeployments[mdIndex].OSImagesByTKR[tkrNameResolved][osImageResolved].Spec.Image.Ref
	Expect(osImageRefFromTKRDATA).To(BeEquivalentTo(osImageRefFromResolvedResult),
		fmt.Sprintf("MD[%d]'s TKR_DATA[%s]'s osImgeRef doesn't match with osImageRef of the OSImage CR : '%s'", mdIndex, cluster.Spec.Topology.Version, tkrNameResolved))
}
func getTKRsForUpgrade(defaultTKRVersion string, tkrs []*runv1.TanzuKubernetesRelease) (newTKR, oldTKR *runv1.TanzuKubernetesRelease) {
	var defaultTKR *runv1.TanzuKubernetesRelease
	for i := range tkrs {
		if tkrs[i].Spec.Version == defaultTKRVersion {
			defaultTKR = tkrs[i]
		}
	}

	defaultTKRVersionSemantic, _ := version.ParseSemantic(defaultTKR.Spec.Version)
	sortedCompatibleTKRs := getSortedCompatibleTKRs(tkrs)
	for i := range sortedCompatibleTKRs {
		tkrVersionSemantic, _ := version.ParseSemantic(sortedCompatibleTKRs[i].Spec.Version)
		if tkrVersionSemantic.LessThan(defaultTKRVersionSemantic) && differByOneMinorVersion(defaultTKRVersionSemantic, tkrVersionSemantic) {
			return defaultTKR, sortedCompatibleTKRs[i]
		}
	}
	return defaultTKR, nil
}

func getTKRVersions(tkrs []*runv1.TanzuKubernetesRelease) sets.StringSet {
	tkrVersions := sets.StringSet{}
	for i := range tkrs {
		tkrVersions.Add(tkrs[i].Spec.Version)
	}
	return tkrVersions
}

func differByOneMinorVersion(v1, v2 *version.Version) bool {
	return (v1.Major() == v2.Major()) && (v1.Minor()-v2.Minor() == 1)
}

func getSortedCompatibleTKRs(tkrs []*runv1.TanzuKubernetesRelease) []*runv1.TanzuKubernetesRelease {
	compatibleTKRs := filterCompatibleTKRs(tkrs)
	sort.Sort(byTKRVersion(compatibleTKRs))
	return compatibleTKRs
}

func filterCompatibleTKRs(tkrs []*runv1.TanzuKubernetesRelease) []*runv1.TanzuKubernetesRelease {
	var compatibleTKRs []*runv1.TanzuKubernetesRelease
	for i := range tkrs {
		if _, exists := tkrs[i].Labels[runv1.LabelIncompatible]; !exists {
			compatibleTKRs = append(compatibleTKRs, tkrs[i])
		}
	}
	return compatibleTKRs
}

type byTKRVersion []*runv1.TanzuKubernetesRelease

func (s byTKRVersion) Len() int {
	return len(s)
}
func (s byTKRVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byTKRVersion) Less(i, j int) bool {
	v1, _ := version.ParseSemantic(s[i].Spec.Version) // TKRs are expected to be following semantic version
	v2, _ := version.ParseSemantic(s[j].Spec.Version)
	return v1.LessThan(v2)
}
