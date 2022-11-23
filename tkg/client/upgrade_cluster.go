// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capav1beta2 "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	capzv1beta1 "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capikubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	capdv1beta1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	upgradePatchInterval = 30 * time.Second
	upgradePatchTimeout  = 5 * time.Minute
	vmTemplateMoidKey    = "vmTemplateMoid"
)

// UpgradeClusterOptions upgrade cluster options
type UpgradeClusterOptions struct {
	ClusterName         string
	Namespace           string
	KubernetesVersion   string
	TkrVersion          string
	Kubeconfig          string
	VSphereTemplateName string
	OSName              string
	OSVersion           string
	OSArch              string
	IsRegionalCluster   bool
	SkipAddonUpgrade    bool
	SkipPrompt          bool
	IsTKGSCluster       bool
	// Tanzu edition (either tce or tkg)
	Edition string
}

type mdInfastructureTemplateInfo struct {
	MDInfrastructureTemplateName      string
	MDInfrastructureTemplateNamespace string
}

// ComponentInfo defines cluster component related metadata used for upgrade
type ComponentInfo struct {
	TkrVersion                         string
	KubernetesVersion                  string
	ImageRepository                    string
	CoreDNSImageRepository             string
	CoreDNSImageTag                    string
	EtcdDataDir                        string
	EtcdImageRepository                string
	EtcdImageTag                       string
	EtcdExtraArgs                      map[string]string
	KCPInfrastructureTemplateName      string
	KCPInfrastructureTemplateNamespace string
	MDInfastructureTemplates           map[string]mdInfastructureTemplateInfo
	VSphereVMTemplateName              string
	VSphereVMTemplateMOID              string
	AwsAMIID                           string
	CAPDImageName                      string
	CAPDImageRepo                      string
	AwsRegionToAMIMap                  map[string][]tkgconfigbom.AMIInfo
	AzureImage                         tkgconfigbom.AzureInfo
	OsInfo                             tkgconfigbom.OSInfo
}

type upgradeStatus string

const (
	upgradeStateInitiated             = "Initiated"
	upgradeStateInfraTemplatesCreated = "InfraTemplatesCreated"
	upgradeStateKCPPatchApplied       = "KCPPatchApplied"
	upgradeStateKCPUpgraded           = "KCPUpgraded"
	upgradeStateMDPatchApplied        = "MDPatchApplied"
	upgradeStateSuccess               = "Success"
)

// ClusterUpgradeInfo defines cluster upgrade metadata used during upgrade process
type ClusterUpgradeInfo struct {
	UpgradeComponentInfo ComponentInfo
	ActualComponentInfo  ComponentInfo

	KCPObjectName      string
	KCPObjectNamespace string
	MDObjects          []capi.MachineDeployment
	ClusterName        string
	ClusterNamespace   string

	UpgradeState upgradeStatus
}

// UpgradeCluster upgrades workload and management clusters k8s version
// Steps:
//  1. Verify k8s version
//  2. Get the Upgrade configuration by reading BOM file to get the ImageTag and ImageRepository information for CoreDNS and Etcd,
//     Read AWS_AMI_ID map from BOM for AWS upgrade scenario. Also use command line argument options to fill the upgrade configuration
//  3. Create InfrastructureMachineTemplates(VSphereMachineTemplate, AWSMachineTemplate, AzureMachineTemplate) required for upgrade
//  4. Patch KCP object to upgrade control-plane nodes
//  5. Wait for k8s version to be updated for the cluster
//  6. Patch MachineDeployment object to upgrade worker nodes
//  7. Wait for k8s version to be updated for all worker nodes
func (c *TkgClient) UpgradeCluster(options *UpgradeClusterOptions) error {
	if options == nil {
		return errors.New("invalid upgrade cluster options nil")
	}

	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	options.Kubeconfig = currentRegion.SourceFilePath

	log.V(4).Info("Creating management cluster client...")
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while upgrading cluster")
	}

	// get the management cluster name and namespace in case of management cluster upgrade
	if options.IsRegionalCluster {
		clusterName, namespace, err := c.getRegionalClusterNameAndNamespace(regionalClusterClient)
		if err != nil {
			return errors.Wrap(err, "unable to get current management cluster information")
		}
		options.ClusterName = clusterName
		options.Namespace = namespace
	}

	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	// Check if Cluster is ClusterClass based cluster or not
	isClusterClassBased, err := regionalClusterClient.IsClusterClassBased(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to determine cluster type")
	}

	// If this is TKGS cluster and using TKC API then upgrade using legacy TKC based approach
	if options.IsTKGSCluster && !isClusterClassBased {
		return c.DoPacificClusterUpgrade(regionalClusterClient, options)
	}

	var currentClusterClient clusterclient.Client
	if options.IsRegionalCluster {
		currentClusterClient = regionalClusterClient
	} else if !options.IsTKGSCluster {
		log.V(4).Info("Creating workload cluster client...")
		currentClusterClient, err = c.getWorkloadClusterClient(options.ClusterName, options.Namespace)
		if err != nil {
			return errors.Wrap(err, "unable to get workload cluster client")
		}
	}

	// If cluster is ClusterClass based cluster upgrade the cluster with different path
	if isClusterClassBased {
		return c.DoClassyClusterUpgrade(regionalClusterClient, currentClusterClient, options)
	}

	log.Warning(constants.YTTBasedClusterWarning)
	return c.DoLegacyClusterUpgrade(regionalClusterClient, currentClusterClient, options)
}

func (c *TkgClient) DoLegacyClusterUpgrade(regionalClusterClient, currentClusterClient clusterclient.Client, options *UpgradeClusterOptions) error {
	log.Info("Verifying kubernetes version...")
	err := c.verifyK8sVersion(currentClusterClient, options.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubernetes version verification failed")
	}

	err = c.addKubernetesReleaseLabel(regionalClusterClient, options)
	if err != nil {
		return errors.Wrapf(err, "unable to patch the cluster object with TanzuKubernetesRelease label")
	}

	// Upgrade/Add certain addons on old clusters during upgrade
	// The addons should upgrade prior to cluster upgrade to account for forward compatibility
	// i.e. some old addons may not run on the nodes with new k8s version
	// We will ensure backward compatibility when shipping packages going forward
	// With package-package-lcm approach addons will be upgraded as part of management package upgrade
	// and we do not need to upgrade addons with below function
	if !options.SkipAddonUpgrade && !config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		err = c.upgradeAddonPreNodeUpgrade(regionalClusterClient, currentClusterClient, options.ClusterName, options.Namespace, options.IsRegionalCluster, options.Edition)
		if err != nil {
			return err
		}
	}

	err = c.DoClusterUpgrade(regionalClusterClient, currentClusterClient, options)
	if err != nil {
		return err
	}

	// Upgrade addon metadata configmaps after the nodes are upgraded
	if !options.SkipAddonUpgrade && !config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		err = c.upgradeAddonPostNodeUpgrade(regionalClusterClient, currentClusterClient, options.ClusterName, options.Namespace, options.IsRegionalCluster, options.Edition)
		if err != nil {
			return err
		}
	}

	if !options.IsRegionalCluster {
		// update autoscaler deployment if enabled
		err = regionalClusterClient.ApplyPatchForAutoScalerDeployment(c.tkgBomClient, options.ClusterName, options.KubernetesVersion, options.Namespace)
		if err != nil {
			return errors.Wrapf(err, "failed to upgrade autoscaler for cluster '%s'", options.ClusterName)
		}
	}

	if options.IsRegionalCluster {
		if !config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
			log.Info("Waiting for additional components to be up and running...")
			if err := c.WaitForAddonsDeployments(regionalClusterClient); err != nil {
				return err
			}
		}
	}

	log.Info("Waiting for packages to be up and running...")
	if err := c.WaitForPackages(regionalClusterClient, currentClusterClient, options.ClusterName, options.Namespace, options.IsRegionalCluster); err != nil {
		log.Warningf("Warning: Cluster is upgraded successfully, but some packages are failing. %v", err)
	}

	return nil
}

// DoPacificClusterUpgrade perform TKGS cluster upgrade
func (c *TkgClient) DoPacificClusterUpgrade(regionalClusterClient clusterclient.Client, options *UpgradeClusterOptions) error {
	if options.IsRegionalCluster {
		return errors.New("upgrading kubernetes on 'Tanzu Kubernetes Cluster service for vSphere' management cluster is not yet supported")
	}
	log.Infof("Patching TanzuKubernetesCluster object with the kubernetes version %s...", options.KubernetesVersion)
	if err := regionalClusterClient.PatchK8SVersionToPacificCluster(options.ClusterName, options.Namespace, options.KubernetesVersion); err != nil {
		return errors.Wrap(err, "failed to update the Kubernetes version for TanzuKubernetesCluster object")
	}
	log.Info("Waiting for the 'Tanzu Kubernetes Cluster service for vSphere' cluster kubernetes version update and it may take a while...")
	if err := regionalClusterClient.WaitForPacificClusterK8sVersionUpdate(options.ClusterName, options.Namespace, options.KubernetesVersion); err != nil {
		return errors.Wrap(err, "failed waiting on updating kubernetes version for 'Tanzu Kubernetes Cluster service for vSphere' cluster")
	}

	return nil
}

// DoClusterUpgrade upgrades cluster
func (c *TkgClient) DoClusterUpgrade(regionalClusterClient clusterclient.Client,
	currentClusterClient clusterclient.Client, options *UpgradeClusterOptions) error {

	if err := c.configureOSOptionsForUpgrade(regionalClusterClient, options); err != nil {
		return errors.Wrap(err, "error configuring os options during upgrade")
	}

	log.Info("Retrieving configuration for upgrade cluster...")
	upgradeClusterConfig, err := c.getUpgradeClusterConfig(options)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve component upgrade info")
	}

	upgradeClusterConfig.UpgradeState = upgradeStateInitiated

	log.Info("Create InfrastructureTemplate for upgrade...")
	err = c.createInfrastructureTemplateForUpgrade(regionalClusterClient, upgradeClusterConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create infrastructure template for upgrade")
	}
	upgradeClusterConfig.UpgradeState = upgradeStateInfraTemplatesCreated

	err = c.applyPatchAndWait(regionalClusterClient, currentClusterClient, upgradeClusterConfig)
	if err != nil {
		return err
	}

	// osInfo annotation format: "ubuntu,20.04,amd64"
	osInfoString := fmt.Sprintf("%s,%s,%s", upgradeClusterConfig.UpgradeComponentInfo.OsInfo.Name, upgradeClusterConfig.UpgradeComponentInfo.OsInfo.Version, upgradeClusterConfig.UpgradeComponentInfo.OsInfo.Arch)
	err = regionalClusterClient.PatchClusterObjectAnnotations(upgradeClusterConfig.ClusterName, upgradeClusterConfig.ClusterNamespace, "osInfo", osInfoString)
	if err != nil {
		return errors.Wrap(err, "error while patching osInfo to the cluster resource")
	}

	return nil
}

func (c *TkgClient) addKubernetesReleaseLabel(regionalClusterClient clusterclient.Client, options *UpgradeClusterOptions) error {
	patchFormat := `
	{
		"metadata": {
			"labels": {
				"tanzuKubernetesRelease": "%s"
			}
		}
	}`
	patchLabel := fmt.Sprintf(patchFormat, utils.GetTkrNameFromTkrVersion(options.TkrVersion))
	pollOptions := clusterclient.PollOptions{Interval: upgradePatchInterval, Timeout: upgradePatchTimeout}
	err := regionalClusterClient.PatchClusterObjectWithPollOptions(options.ClusterName, options.Namespace, patchLabel, &pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to patch the cluster object with TanzuKubernetesRelease label")
	}
	return nil
}

// upgradeAddonPreNodeUpgrade upgrades kapp-controller, addons-manager, tkr-controller and core packageRepository
// before control plane nodes and worker nodes are bumped to new K8S version, to take care of forward compatibility
func (c *TkgClient) upgradeAddonPreNodeUpgrade(regionalClusterClient clusterclient.Client, currentClusterClient clusterclient.Client,
	clusterName string, clusterNamespace string, isRegionalCluster bool, tanzuEdition string) error {

	addonsToBeUpgraded := []string{
		"addons-management/kapp-controller",
	}
	// tanzu-addons-manager and tkr-controller only runs in management cluster
	if isRegionalCluster {
		addonsToBeUpgraded = append(addonsToBeUpgraded,
			"addons-management/tanzu-addons-manager",
			"tkr/tkr-controller",
			"addons-management/core-package-repo",
			"packages/management-package-repo",
			"packages/management-package")
	}
	upgradeClusterMetadataOptions := &UpgradeAddonOptions{
		AddonNames:        addonsToBeUpgraded,
		ClusterName:       clusterName,
		Namespace:         clusterNamespace,
		IsRegionalCluster: isRegionalCluster,
		Edition:           tanzuEdition,
	}

	err := c.DoUpgradeAddon(regionalClusterClient, currentClusterClient, upgradeClusterMetadataOptions, c.GetClusterConfiguration)
	if err != nil {
		return errors.Wrap(err, "failed to update additional addon components")
	}

	return nil
}

// upgradeAddonPostNodeUpgrade upgrades metadata configmaps and core packageRepository after node upgrade
func (c *TkgClient) upgradeAddonPostNodeUpgrade(regionalClusterClient clusterclient.Client, currentClusterClient clusterclient.Client,
	clusterName string, clusterNamespace string, isRegionalCluster bool, tanzuEdition string) error {

	addonsToBeUpgraded := []string{
		"metadata/tkg",
	}
	if tanzuEdition != "tce" {
		addonsToBeUpgraded = append(addonsToBeUpgraded, "addons-management/standard-package-repo")
	}
	upgradeClusterMetadataOptions := &UpgradeAddonOptions{
		AddonNames:        addonsToBeUpgraded,
		ClusterName:       clusterName,
		Namespace:         clusterNamespace,
		IsRegionalCluster: isRegionalCluster,
		Edition:           tanzuEdition,
	}

	err := c.DoUpgradeAddon(regionalClusterClient, currentClusterClient, upgradeClusterMetadataOptions, c.GetClusterConfiguration)
	if err != nil {
		return errors.Wrap(err, "failed to update metadata configmaps after kubernetes upgrade")
	}

	return nil
}

func (c *TkgClient) applyPatchAndWait(regionalClusterClient, currentClusterClient clusterclient.Client, upgradeClusterConfig *ClusterUpgradeInfo) error {
	var err error
	kubernetesVersion := upgradeClusterConfig.UpgradeComponentInfo.KubernetesVersion

	// Ensure Cluster API Provider AWS is running on the control plane before continuing with EC2 instance profile
	if err := currentClusterClient.PatchClusterAPIAWSControllersToUseEC2Credentials(); err != nil {
		return err
	}

	// Clusters deployed with TKG CLI version prior to v1.2 uses `beta.kubernetes.io/os: linux` nodeSelector
	// for `calico-node` daemonset and `calico-kube-controller` deployment.
	// As k8s v1.19.x removed the support for `beta.kubernetes.io/os: linux` node label and it requires nodes
	// to have `kubernetes.io/os: linux` label, we need to patch `calico-node` daemonset's and
	// `calico-kube-controller` deployment's nodeSelector to use new `kubernetes.io/os: linux`
	// Reference: https://github.com/kubernetes/kubernetes/commit/54c0f8b677d0b82258f3b4df6d325cc3c0011661
	if err := currentClusterClient.PatchCalicoNodeDaemonSetWithNewNodeSelector("kubernetes.io/os", "linux"); err != nil {
		return errors.Wrap(err, "unable to patch 'calico-node' daemonset")
	}
	if err := currentClusterClient.PatchCalicoKubeControllerDeploymentWithNewNodeSelector("kubernetes.io/os", "linux"); err != nil {
		return errors.Wrap(err, "unable to patch 'calico-node' daemonset")
	}

	// If user is using custom image repository, update the CoreDNS imageRepository
	// in kubeadm-config ConfigMap before starting control-plane upgrade
	// IMPORTANT: This change is needed to fix the clusters created with v1.0.x and v1.1.x in air-gapped case, where
	// as container images are available inside node, we did not used custom image repository for KCP which user provided,
	// And as 'registry.tkg.vmware.run' is not reachable in air-gapped case during upgrade, we are making this patch
	// before we start actual upgrade process so coredns container image is pullable across nodes.
	if tkgconfighelper.IsCustomRepository(upgradeClusterConfig.UpgradeComponentInfo.ImageRepository) && !tkgconfighelper.SkipImageReferenceUpdateOnUpgrade() {
		log.Info("Configuring cluster for upgrade...")
		log.V(3).Info("Updating coreDNS imageRepository in kubeadm-config ConfigMap...")
		if err = currentClusterClient.PatchCoreDNSImageRepositoryInKubeadmConfigMap(upgradeClusterConfig.UpgradeComponentInfo.ImageRepository); err != nil {
			return errors.Wrap(err, "unable to update the kubeadm configmap with new image repository")
		}
	}

	// Upgrade logic for kapp-controller related changes
	if err := c.handleKappControllerUpgrade(regionalClusterClient, currentClusterClient, upgradeClusterConfig); err != nil {
		return errors.Wrapf(err, "unable to apply upgrade for kapp-controller")
	}

	log.Info("Upgrading control plane nodes...")
	log.Infof("Patching KubeadmControlPlane with the kubernetes version %s...", kubernetesVersion)
	err = c.PatchKubernetesVersionToKubeadmControlPlane(regionalClusterClient, upgradeClusterConfig)
	if err != nil {
		return errors.Wrap(err, "unable to patch kubernetes version to kubeadm control plane")
	}
	upgradeClusterConfig.UpgradeState = upgradeStateKCPPatchApplied

	// If user is using custom image repository, update the kube-proxy imageRepository
	// in kube-proxy daemonset after starting control-plane upgrade
	// Note: kube-proxy daemonset update is done after we patch KCP object because CAPI control-plane controller
	// during reconciliation updates kube-proxy daemonset's imageRepository from KCP.Spec.ClusterConfiguration.ImageRepository
	// if the upgrade process is not started, this will override the kube-proxy daemonset update if done before KCP patch
	// IMPORTANT: This change is needed to fix the clusters created with v1.0.x and v1.1.x in air-gapped case, where
	// as container images are available inside node, we did not used custom image repository for KCP which user provided,
	// And as 'registry.tkg.vmware.run' is not reachable in air-gapped case during upgrade, we are making this patch
	// before we start actual upgrade process so kubeproxy container image is pullable across nodes.
	if tkgconfighelper.IsCustomRepository(upgradeClusterConfig.UpgradeComponentInfo.ImageRepository) && !tkgconfighelper.SkipImageReferenceUpdateOnUpgrade() {
		log.V(3).Info("Updating imageRepository for kube-proxy daemonset...")
		if err := currentClusterClient.PatchImageRepositoryInKubeProxyDaemonSet(upgradeClusterConfig.UpgradeComponentInfo.ImageRepository); err != nil {
			return errors.Wrap(err, "unable to update the kube-proxy daemonset with new image repository")
		}
	}

	log.Info("Waiting for kubernetes version to be updated for control plane nodes")
	err = regionalClusterClient.WaitK8sVersionUpdateForCPNodes(upgradeClusterConfig.ClusterName, upgradeClusterConfig.ClusterNamespace, kubernetesVersion, currentClusterClient)
	if err != nil {
		return errors.Wrap(err, "error waiting for kubernetes version update for kubeadm control plane")
	}
	upgradeClusterConfig.UpgradeState = upgradeStateKCPUpgraded

	log.Info("Upgrading worker nodes...")
	log.Infof("Patching MachineDeployment with the kubernetes version %s...", kubernetesVersion)
	err = c.patchKubernetesVersionToMachineDeployment(regionalClusterClient, upgradeClusterConfig)
	if err != nil {
		return errors.Wrap(err, "unable to patch kubernetes version to kubeadm control plane")
	}
	upgradeClusterConfig.UpgradeState = upgradeStateMDPatchApplied

	log.Info("Waiting for kubernetes version to be updated for worker nodes...")
	err = regionalClusterClient.WaitK8sVersionUpdateForWorkerNodes(upgradeClusterConfig.ClusterName, upgradeClusterConfig.ClusterNamespace, kubernetesVersion, currentClusterClient)
	if err != nil {
		return errors.Wrap(err, "error waiting for kubernetes version update for worker nodes")
	}

	upgradeClusterConfig.UpgradeState = upgradeStateSuccess
	return nil
}

func (c *TkgClient) getUpgradeClusterConfig(options *UpgradeClusterOptions) (*ClusterUpgradeInfo, error) {
	bomConfiguration, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(options.TkrVersion)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read in configuration from BOM file")
	}

	if tkgconfighelper.IsCustomRepository(bomConfiguration.ImageConfig.ImageRepository) {
		log.Infof("Using custom image repository: %s", bomConfiguration.ImageConfig.ImageRepository)
	}

	upgradeInfo := &ClusterUpgradeInfo{}
	upgradeInfo.UpgradeComponentInfo.TkrVersion = bomConfiguration.Release.Version
	upgradeInfo.UpgradeComponentInfo.KubernetesVersion = bomConfiguration.KubeadmConfigSpec.KubernetesVersion
	upgradeInfo.UpgradeComponentInfo.CoreDNSImageTag = bomConfiguration.KubeadmConfigSpec.DNS.ImageTag
	upgradeInfo.UpgradeComponentInfo.EtcdDataDir = bomConfiguration.KubeadmConfigSpec.Etcd.Local.DataDir
	upgradeInfo.UpgradeComponentInfo.EtcdImageTag = bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageTag
	upgradeInfo.UpgradeComponentInfo.EtcdExtraArgs = bomConfiguration.KubeadmConfigSpec.Etcd.Local.ExtraArgs

	upgradeInfo.ClusterName = options.ClusterName
	upgradeInfo.ClusterNamespace = options.Namespace

	upgradeInfo.UpgradeComponentInfo.AwsRegionToAMIMap = bomConfiguration.AMI
	upgradeInfo.UpgradeComponentInfo.VSphereVMTemplateName = options.VSphereTemplateName

	// get the Azure VM image info from TKG config if available and fall back to the image info from BOM if not available in TKG config file
	azureVMImage, err := c.tkgConfigProvidersClient.GetAzureVMImageInfo(upgradeInfo.UpgradeComponentInfo.TkrVersion)
	if err == nil && azureVMImage != nil {
		// TODO: what if error is returned or azureVMImage is nil, handle that case
		upgradeInfo.UpgradeComponentInfo.AzureImage = *azureVMImage
		upgradeInfo.UpgradeComponentInfo.OsInfo = azureVMImage.OSInfo
	}

	// We are hard-coding the assumption that during upgrade imageConfig.ImageRepository should take precedence
	// over whatever is spelled out in the KubeAdmConfigSpec section.
	// This change also implies when imageConfig.ImageRepository differs from kubeadmConfigSpec's repository,
	// we will end up pulling container images from imageConfig.ImageRepository despite the same images
	// associated with the latter are in possibly in the node already.
	// Testcase: When doing management cluster upgrade, it was observed that once the KCP upgrade is complete and
	// before the new worker nodes are up, old worker nodes are trying to pull new coredns and kubeproxy images from
	// `projects.registry.vmware.com/tkg` registry and as during dev cycle, we do not publish container images to this
	// repository but rather we publish it to our staging registry, it is important for us to use staging registry for
	// KCP patch during cluster upgrade workflow.
	upgradeInfo.UpgradeComponentInfo.ImageRepository = bomConfiguration.ImageConfig.ImageRepository
	upgradeInfo.UpgradeComponentInfo.CoreDNSImageRepository = bomConfiguration.ImageConfig.ImageRepository
	upgradeInfo.UpgradeComponentInfo.EtcdImageRepository = bomConfiguration.ImageConfig.ImageRepository
	if bomConfiguration.ImageConfig.ImageRepository != bomConfiguration.KubeadmConfigSpec.ImageRepository {
		log.V(3).Infof("Using %s registry during the upgrade process...", bomConfiguration.ImageConfig.ImageRepository)
	}

	return upgradeInfo, nil
}

// Updating VM_TEMPLATE/AWS_AMI_ID in existing template will not help as it is passed as reference and controllers will not get reconciled unless the name
// of the InfrastructureMachineTemplate is changed under KCP.Spec.InfrastructureTemplate and MD.Spec.Template.Spec.infrastructureRef
// Because of the above reason we need to create new InfrastructureTemplates and update the reference in KCP and MD object of existing cluster
func (c *TkgClient) createInfrastructureTemplateForUpgrade(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error

	kcp, err := regionalClusterClient.GetKCPObjectForCluster(clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
	if err != nil {
		return errors.Wrapf(err, "unable to find control plane node object for cluster %s", clusterUpgradeConfig.ClusterName)
	}

	machineDeploymentObjects, err := regionalClusterClient.GetMDObjectForCluster(clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
	if err != nil {
		return errors.Wrapf(err, "unable to get MachineDeployment for cluster with name %s in namespace %s", clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
	}

	clusterUpgradeConfig.KCPObjectName = kcp.Name
	clusterUpgradeConfig.KCPObjectNamespace = kcp.Namespace

	clusterUpgradeConfig.ActualComponentInfo.KubernetesVersion = kcp.Spec.Version
	clusterUpgradeConfig.ActualComponentInfo.ImageRepository = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.ImageRepository
	clusterUpgradeConfig.ActualComponentInfo.CoreDNSImageRepository = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.DNS.ImageRepository
	clusterUpgradeConfig.ActualComponentInfo.CoreDNSImageTag = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.DNS.ImageTag
	clusterUpgradeConfig.ActualComponentInfo.EtcdDataDir = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.DataDir
	clusterUpgradeConfig.ActualComponentInfo.EtcdImageRepository = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ImageRepository
	clusterUpgradeConfig.ActualComponentInfo.EtcdImageTag = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ImageTag
	clusterUpgradeConfig.ActualComponentInfo.EtcdExtraArgs = kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs

	clusterUpgradeConfig.ActualComponentInfo.KCPInfrastructureTemplateName = kcp.Spec.MachineTemplate.InfrastructureRef.Name
	clusterUpgradeConfig.ActualComponentInfo.KCPInfrastructureTemplateNamespace = kcp.Spec.MachineTemplate.InfrastructureRef.Namespace

	clusterUpgradeConfig.MDObjects = machineDeploymentObjects
	clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates = make(map[string]mdInfastructureTemplateInfo)
	clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates = make(map[string]mdInfastructureTemplateInfo)

	for i := range clusterUpgradeConfig.MDObjects {
		// set actual MD object information in 'clusterUpgradeConfig.ActualComponentInfo'
		clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
			MDInfrastructureTemplateName:      clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name,
			MDInfrastructureTemplateNamespace: clusterUpgradeConfig.MDObjects[i].Namespace,
		}
	}

	switch kcp.Spec.MachineTemplate.InfrastructureRef.Kind {
	case constants.KindVSphereMachineTemplate:
		return c.createVsphereInfrastructureTemplateForUpgrade(regionalClusterClient, kcp, clusterUpgradeConfig)
	case constants.KindAWSMachineTemplate:
		return c.createAWSInfrastructureTemplateForUpgrade(regionalClusterClient, kcp, clusterUpgradeConfig)
	case constants.KindAzureMachineTemplate:
		return c.createAzureInfrastructureTemplateForUpgrade(regionalClusterClient, kcp, clusterUpgradeConfig)
	case constants.KindDockerMachineTemplate:
		return c.createCAPDInfrastructureTemplateForUpgrade(regionalClusterClient, kcp, clusterUpgradeConfig)
	default:
		return errors.New("infrastructure template associated with KubeadmControlPlane object is invalid")
	}
}

func isNewAWSTemplateRequired(machineTemplate *capav1beta2.AWSMachineTemplate, clusterUpgradeConfig *ClusterUpgradeInfo, actualK8sVersion *string) bool {
	if actualK8sVersion == nil || *actualK8sVersion != clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion {
		return true
	}

	// If vm template given is the same as we already have in AWSMachineTemplate
	if machineTemplate.Spec.Template.Spec.AMI.ID == nil ||
		*machineTemplate.Spec.Template.Spec.AMI.ID != clusterUpgradeConfig.UpgradeComponentInfo.AwsAMIID {
		return true
	}
	return false
}

func isNewDockerTemplateRequired(machineTemplate *capdv1beta1.DockerMachineTemplate, clusterUpgradeConfig *ClusterUpgradeInfo, actualK8sVersion *string) bool {
	if actualK8sVersion == nil || *actualK8sVersion != clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion {
		return true
	}
	// If machine template given is the same as we already have in DockerMachineTemplate
	if machineTemplate.Spec.Template.Spec.CustomImage == "" ||
		machineTemplate.Spec.Template.Spec.CustomImage != clusterUpgradeConfig.UpgradeComponentInfo.CAPDImageName {
		return true
	}
	return false
}

func (c *TkgClient) createAWSControlPlaneMachineTemplate(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error
	awsMachineTemplate := &capav1beta2.AWSMachineTemplate{}
	err = regionalClusterClient.GetResource(awsMachineTemplate, kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to find AWSMachineTemplate with name '%s' in namespace '%s'", kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace)
	}

	// Naming format of the template: Current naming format for AWSMachineTemplate for KCP is {CLUSTER_NAME}-control-plane-{KUBERNETES_VERSION}-{random-string}
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = clusterUpgradeConfig.ClusterName + "-control-plane-" +
		utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true)
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace = awsMachineTemplate.Namespace

	if !isNewAWSTemplateRequired(awsMachineTemplate, clusterUpgradeConfig, &kcp.Spec.Version) {
		clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = awsMachineTemplate.Name
		return nil
	}

	awsMachineTemplateForUpgrade := &capav1beta2.AWSMachineTemplate{}
	awsMachineTemplateForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName
	awsMachineTemplateForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace
	awsMachineTemplateForUpgrade.Spec = awsMachineTemplate.DeepCopy().Spec
	awsMachineTemplateForUpgrade.Spec.Template.Spec.AMI.ID = &clusterUpgradeConfig.UpgradeComponentInfo.AwsAMIID // TODO(anuj): Decide on AMI-ID vs ImageLookupOrg implementation approach

	err = regionalClusterClient.CreateResource(awsMachineTemplateForUpgrade, awsMachineTemplateForUpgrade.Name, awsMachineTemplateForUpgrade.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to create AWSMachineTemplate for upgrade with name '%s' in namespace '%s'", awsMachineTemplateForUpgrade.Name, awsMachineTemplateForUpgrade.Namespace)
	}

	return nil
}

func (c *TkgClient) createAWSMachineDeploymentMachineTemplateForWorkers(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error

	for i := range clusterUpgradeConfig.MDObjects {
		// get aws machine template for given machine deployment
		awsMachineTemplateForMD := &capav1beta2.AWSMachineTemplate{}
		err = regionalClusterClient.GetResource(awsMachineTemplateForMD, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace, nil, nil)
		if err != nil {
			return errors.Wrapf(err, "unable to find AWSMachineTemplate with name '%s' in namespace '%s'", clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace)
		}

		// if no template change required, update the clusterUpgradeConfig.UpgradeComponentInfo and return immediately
		if !isNewAWSTemplateRequired(awsMachineTemplateForMD, clusterUpgradeConfig, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.Version) {
			clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
				MDInfrastructureTemplateName:      awsMachineTemplateForMD.Name,
				MDInfrastructureTemplateNamespace: awsMachineTemplateForMD.Namespace,
			}
			return nil
		}

		// Naming format of the MD template: Current naming format for AWSMachineTemplate for MachineDeployment is {ACTUAL_TEMPLATE_NAME}-{KUBERNETES_VERSION}-{random-string}
		clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
			MDInfrastructureTemplateName: clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName + "-" +
				utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true),
			MDInfrastructureTemplateNamespace: awsMachineTemplateForMD.Namespace,
		}

		awsMachineTemplateMDForUpgrade := &capav1beta2.AWSMachineTemplate{}
		awsMachineTemplateMDForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName
		awsMachineTemplateMDForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateNamespace
		awsMachineTemplateMDForUpgrade.Spec = awsMachineTemplateForMD.DeepCopy().Spec
		awsMachineTemplateMDForUpgrade.Spec.Template.Spec.AMI.ID = &clusterUpgradeConfig.UpgradeComponentInfo.AwsAMIID // TODO(anuj): Decide on AMI-ID vs ImageLookupOrg implementation approach

		// create template for each machine deployment object
		err = regionalClusterClient.CreateResource(awsMachineTemplateMDForUpgrade, awsMachineTemplateMDForUpgrade.Name, awsMachineTemplateMDForUpgrade.Namespace)
		if err != nil {
			return errors.Wrapf(err, "unable to create AWSMachineTemplate for upgrade with name '%s' in namespace '%s'", awsMachineTemplateMDForUpgrade.Name, awsMachineTemplateMDForUpgrade.Namespace)
		}
	}

	return nil
}

func (c *TkgClient) createCAPDInfrastructureTemplateForUpgrade(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	err := c.getCAPDImageForK8sVersion(clusterUpgradeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to get docker image for CAPD template")
	}
	if err := c.createCAPDControlPlaneMachineTemplate(regionalClusterClient, kcp, clusterUpgradeConfig); err != nil {
		return err
	}
	if err := c.createCAPDMachineDeploymentMachineTemplateForWorkers(regionalClusterClient, clusterUpgradeConfig); err != nil {
		return err
	}

	return nil
}

func (c *TkgClient) createAWSInfrastructureTemplateForUpgrade(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	err := c.getAWSAMIIDForK8sVersion(regionalClusterClient, clusterUpgradeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to get AMIID for aws template")
	}
	if err := c.createAWSControlPlaneMachineTemplate(regionalClusterClient, kcp, clusterUpgradeConfig); err != nil {
		return err
	}
	if err := c.createAWSMachineDeploymentMachineTemplateForWorkers(regionalClusterClient, clusterUpgradeConfig); err != nil {
		return err
	}
	return nil
}

func isNewAzureTemplateRequired(machineTemplate *capzv1beta1.AzureMachineTemplate, clusterUpgradeConfig *ClusterUpgradeInfo, actualK8sVersion *string) bool { // nolint:gocyclo
	if actualK8sVersion == nil || *actualK8sVersion != clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion {
		return true
	}

	if machineTemplate.Spec.Template.Spec.Image == nil {
		return true
	}

	if isMarketplaceImage(&clusterUpgradeConfig.UpgradeComponentInfo.AzureImage) && // nolint:dupl
		(machineTemplate.Spec.Template.Spec.Image.Marketplace == nil ||
			machineTemplate.Spec.Template.Spec.Image.Marketplace.Publisher != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Publisher ||
			machineTemplate.Spec.Template.Spec.Image.Marketplace.Offer != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Offer ||
			machineTemplate.Spec.Template.Spec.Image.Marketplace.SKU != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Sku ||
			machineTemplate.Spec.Template.Spec.Image.Marketplace.Version != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Version ||
			machineTemplate.Spec.Template.Spec.Image.Marketplace.ThirdPartyImage != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.ThirdPartyImage) {
		return true
	}

	if isSharedGalleryImage(&clusterUpgradeConfig.UpgradeComponentInfo.AzureImage) && //nolint:dupl
		(machineTemplate.Spec.Template.Spec.Image.SharedGallery == nil ||
			machineTemplate.Spec.Template.Spec.Image.SharedGallery.ResourceGroup != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.ResourceGroup ||
			machineTemplate.Spec.Template.Spec.Image.SharedGallery.Name != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Name ||
			machineTemplate.Spec.Template.Spec.Image.SharedGallery.SubscriptionID != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.SubscriptionID ||
			machineTemplate.Spec.Template.Spec.Image.SharedGallery.Gallery != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Gallery ||
			machineTemplate.Spec.Template.Spec.Image.SharedGallery.Version != clusterUpgradeConfig.UpgradeComponentInfo.AzureImage.Version) {
		return true
	}

	return false
}

func isMarketplaceImage(azureImage *tkgconfigbom.AzureInfo) bool {
	return azureImage.Publisher != "" && azureImage.Offer != "" && azureImage.Sku != "" && azureImage.Version != ""
}

func isSharedGalleryImage(azureImage *tkgconfigbom.AzureInfo) bool {
	return azureImage.Name != "" && azureImage.ResourceGroup != "" && azureImage.SubscriptionID != "" && azureImage.Gallery != "" && azureImage.Version != ""
}

func getAzureImage(azureImage *tkgconfigbom.AzureInfo) *capzv1beta1.Image {
	if isMarketplaceImage(azureImage) {
		return &capzv1beta1.Image{
			Marketplace: &capzv1beta1.AzureMarketplaceImage{
				ImagePlan: capzv1beta1.ImagePlan{
					Publisher: azureImage.Publisher,
					Offer:     azureImage.Offer,
					SKU:       azureImage.Sku,
				},
				Version:         azureImage.Version,
				ThirdPartyImage: azureImage.ThirdPartyImage,
			},
		}
	}

	if isSharedGalleryImage(azureImage) {
		return &capzv1beta1.Image{
			SharedGallery: &capzv1beta1.AzureSharedGalleryImage{
				ResourceGroup:  azureImage.ResourceGroup,
				Name:           azureImage.Name,
				SubscriptionID: azureImage.SubscriptionID,
				Gallery:        azureImage.Gallery,
				Version:        azureImage.Version,
			},
		}
	}

	return nil
}

func (c *TkgClient) createAzureControlPlaneMachineTemplate(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error
	azureMachineTemplate := &capzv1beta1.AzureMachineTemplate{}
	err = regionalClusterClient.GetResource(azureMachineTemplate, kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to find AzureMachineTemplate with name %s in namespace %s", kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace)
	}

	// Naming format of the template: Current naming format for AzureMachineTemplate for KCP is {CLUSTER_NAME}-control-plane-{KUBERNETES_VERSION}-{random-string}
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = clusterUpgradeConfig.ClusterName + "-control-plane-" +
		utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true)
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace = azureMachineTemplate.Namespace

	if !isNewAzureTemplateRequired(azureMachineTemplate, clusterUpgradeConfig, &kcp.Spec.Version) {
		clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = azureMachineTemplate.Name
		return nil
	}

	azureMachineTemplateForUpgrade := &capzv1beta1.AzureMachineTemplate{}
	azureMachineTemplateForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName
	azureMachineTemplateForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace
	azureMachineTemplateForUpgrade.Spec = azureMachineTemplate.DeepCopy().Spec
	azureMachineTemplateForUpgrade.Spec.Template.Spec.Image = getAzureImage(&clusterUpgradeConfig.UpgradeComponentInfo.AzureImage)

	err = regionalClusterClient.CreateResource(azureMachineTemplateForUpgrade, azureMachineTemplateForUpgrade.Name, azureMachineTemplateForUpgrade.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to create AzureMachineTemplate for upgrade with name %s in namespace %s", azureMachineTemplateForUpgrade.Name, azureMachineTemplateForUpgrade.Namespace)
	}

	return nil
}

func (c *TkgClient) createCAPDControlPlaneMachineTemplate(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error
	dockerMachineTemplate := &capdv1beta1.DockerMachineTemplate{}
	err = regionalClusterClient.GetResource(dockerMachineTemplate, kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to find DockerMachineTemplate with name %s in namespace %s", kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace)
	}

	// Naming format of the template: Current naming format for DockerMachineTemplate for KCP is {CLUSTER_NAME}-control-plane-{KUBERNETES_VERSION}-{random-string}
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = clusterUpgradeConfig.ClusterName + "-control-plane-" +
		utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true)
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace = dockerMachineTemplate.Namespace

	if !isNewDockerTemplateRequired(dockerMachineTemplate, clusterUpgradeConfig, &kcp.Spec.Version) {
		clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = dockerMachineTemplate.Name
		return nil
	}

	dockerMachineTemplateForUpgrade := &capdv1beta1.DockerMachineTemplate{}
	dockerMachineTemplateForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName
	dockerMachineTemplateForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace
	dockerMachineTemplateForUpgrade.Spec = dockerMachineTemplate.DeepCopy().Spec
	dockerMachineTemplateForUpgrade.Spec.Template.Spec.CustomImage = clusterUpgradeConfig.UpgradeComponentInfo.CAPDImageName

	err = regionalClusterClient.CreateResource(dockerMachineTemplateForUpgrade, dockerMachineTemplateForUpgrade.Name, dockerMachineTemplateForUpgrade.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to create dockerMachineTemplate for upgrade with name %s in namespace %s", dockerMachineTemplateForUpgrade.Name, dockerMachineTemplateForUpgrade.Namespace)
	}

	return nil
}

func (c *TkgClient) createCAPDMachineDeploymentMachineTemplateForWorkers(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error
	for i := range clusterUpgradeConfig.MDObjects {
		dockerMachineTemplateForMD := &capdv1beta1.DockerMachineTemplate{}
		err = regionalClusterClient.GetResource(dockerMachineTemplateForMD, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Namespace, nil, nil)
		if err != nil {
			return errors.Wrapf(err, "unable to find DockerMachineTemplate with name %s in namespace %s", clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace)
		}

		if !isNewDockerTemplateRequired(dockerMachineTemplateForMD, clusterUpgradeConfig, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.Version) {
			clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
				MDInfrastructureTemplateName:      dockerMachineTemplateForMD.Name,
				MDInfrastructureTemplateNamespace: dockerMachineTemplateForMD.Namespace,
			}
			return nil
		}

		// Naming format of the MD template: Current naming format for AzureMachineTemplate for MachineDeployment is {ACTUAL_TEMPLATE_NAME}-{KUBERNETES_VERSION}-{random-string}
		clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
			MDInfrastructureTemplateName: clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName + "-" +
				utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true),
			MDInfrastructureTemplateNamespace: dockerMachineTemplateForMD.Namespace,
		}

		dockerMachineTemplateMDForUpgrade := &capdv1beta1.DockerMachineTemplate{}
		dockerMachineTemplateMDForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName
		dockerMachineTemplateMDForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateNamespace
		dockerMachineTemplateMDForUpgrade.Spec = dockerMachineTemplateForMD.DeepCopy().Spec
		dockerMachineTemplateMDForUpgrade.Spec.Template.Spec.CustomImage = clusterUpgradeConfig.UpgradeComponentInfo.CAPDImageName

		err = regionalClusterClient.CreateResource(dockerMachineTemplateMDForUpgrade, dockerMachineTemplateMDForUpgrade.Name, dockerMachineTemplateMDForUpgrade.Namespace)
		if err != nil {
			return errors.Wrapf(err, "unable to create DockerMachineTemplate for upgrade with name %s in namespace %s", dockerMachineTemplateMDForUpgrade.Name, dockerMachineTemplateMDForUpgrade.Namespace)
		}
	}

	return nil
}

func (c *TkgClient) createAzureMachineDeploymentMachineTemplateForWorkers(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error

	for i := range clusterUpgradeConfig.MDObjects {
		azureMachineTemplateForMD := &capzv1beta1.AzureMachineTemplate{}
		err = regionalClusterClient.GetResource(azureMachineTemplateForMD, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace, nil, nil)
		if err != nil {
			return errors.Wrapf(err, "unable to find AzureMachineTemplate with name '%s' in namespace '%s'", clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace)
		}

		if !isNewAzureTemplateRequired(azureMachineTemplateForMD, clusterUpgradeConfig, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.Version) {
			clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
				MDInfrastructureTemplateName:      azureMachineTemplateForMD.Name,
				MDInfrastructureTemplateNamespace: azureMachineTemplateForMD.Namespace,
			}
			return nil
		}

		// Naming format of the MD template: Current naming format for AzureMachineTemplate for MachineDeployment is {ACTUAL_TEMPLATE_NAME}-{KUBERNETES_VERSION}-{random-string}
		clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
			MDInfrastructureTemplateName: clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName + "-" +
				utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true),
			MDInfrastructureTemplateNamespace: azureMachineTemplateForMD.Namespace,
		}

		azureMachineTemplateMDForUpgrade := &capzv1beta1.AzureMachineTemplate{}
		azureMachineTemplateMDForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName
		azureMachineTemplateMDForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateNamespace
		azureMachineTemplateMDForUpgrade.Spec = azureMachineTemplateForMD.DeepCopy().Spec
		azureMachineTemplateMDForUpgrade.Spec.Template.Spec.Image = getAzureImage(&clusterUpgradeConfig.UpgradeComponentInfo.AzureImage)

		err = regionalClusterClient.CreateResource(azureMachineTemplateMDForUpgrade, azureMachineTemplateMDForUpgrade.Name, azureMachineTemplateMDForUpgrade.Namespace)
		if err != nil {
			return errors.Wrapf(err, "unable to create AzureMachineTemplate for upgrade with name %s in namespace %s", azureMachineTemplateMDForUpgrade.Name, azureMachineTemplateMDForUpgrade.Namespace)
		}
	}

	return nil
}

func (c *TkgClient) createAzureInfrastructureTemplateForUpgrade(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	if !isSharedGalleryImage(&clusterUpgradeConfig.UpgradeComponentInfo.AzureImage) && !isMarketplaceImage(&clusterUpgradeConfig.UpgradeComponentInfo.AzureImage) {
		return errors.New("unable to proceed with the upgrade due to invalid azure image information")
	}
	if err := c.createAzureControlPlaneMachineTemplate(regionalClusterClient, kcp, clusterUpgradeConfig); err != nil {
		return err
	}
	if err := c.createAzureMachineDeploymentMachineTemplateForWorkers(regionalClusterClient, clusterUpgradeConfig); err != nil {
		return err
	}
	return nil
}

func isNewVSphereTemplateRequired(machineTemplate *capvv1beta1.VSphereMachineTemplate, clusterUpgradeConfig *ClusterUpgradeInfo, actualK8sVersion *string) bool {
	if actualK8sVersion == nil || *actualK8sVersion != clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion {
		return true
	}
	// If vm moid given is not the same as we already have in VSphereMachineTemplate
	if machineTemplate.Annotations[vmTemplateMoidKey] != clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateMOID {
		return true
	}
	return false
}

func (c *TkgClient) createVSphereControlPlaneMachineTemplate(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	// Get the actual MachineTemplate object associated with actual KCP object
	actualVsphereMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	err := regionalClusterClient.GetResource(actualVsphereMachineTemplate, kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to find VSphereMachineTemplate with name '%s' in namespace '%s'", kcp.Spec.MachineTemplate.InfrastructureRef.Name, kcp.Spec.MachineTemplate.InfrastructureRef.Namespace)
	}

	// Naming format of the template: Current naming format for vsphereTemplate for KCP is {CLUSTER_NAME}-control-plane-{KUBERNETES_VERSION}-{random-string}
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = clusterUpgradeConfig.ClusterName + "-control-plane-" +
		utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true)
	clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace = actualVsphereMachineTemplate.Namespace

	if !isNewVSphereTemplateRequired(actualVsphereMachineTemplate, clusterUpgradeConfig, &kcp.Spec.Version) {
		clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName = actualVsphereMachineTemplate.Name
		return nil
	}

	vsphereMachineTemplateForUpgrade := &capvv1beta1.VSphereMachineTemplate{}
	vsphereMachineTemplateForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName
	vsphereMachineTemplateForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace
	vsphereMachineTemplateForUpgrade.Spec = actualVsphereMachineTemplate.DeepCopy().Spec
	vsphereMachineTemplateForUpgrade.Spec.Template.Spec.Template = clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateName
	vsphereMachineTemplateForUpgrade.Annotations = map[string]string{}
	vsphereMachineTemplateForUpgrade.Annotations[vmTemplateMoidKey] = clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateMOID

	err = regionalClusterClient.CreateResource(vsphereMachineTemplateForUpgrade, vsphereMachineTemplateForUpgrade.Name, vsphereMachineTemplateForUpgrade.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to create VSphereMachineTemplate for upgrade with name '%s' in namespace '%s'", vsphereMachineTemplateForUpgrade.Name, vsphereMachineTemplateForUpgrade.Namespace)
	}

	return nil
}

func (c *TkgClient) createVSphereMachineDeploymentMachineTemplateForWorkers(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error

	for i := range clusterUpgradeConfig.MDObjects {
		// Get the actual MachineTemplate object associated with actual MD object
		actualVsphereMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
		err = regionalClusterClient.GetResource(actualVsphereMachineTemplate, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace, nil, nil)
		if err != nil {
			return errors.Wrapf(err, "unable to find VSphereMachineTemplate with name '%s' in namespace '%s'", clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.InfrastructureRef.Name, clusterUpgradeConfig.MDObjects[i].Namespace)
		}

		if !isNewVSphereTemplateRequired(actualVsphereMachineTemplate, clusterUpgradeConfig, clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.Version) {
			clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
				MDInfrastructureTemplateName:      actualVsphereMachineTemplate.Name,
				MDInfrastructureTemplateNamespace: actualVsphereMachineTemplate.Namespace,
			}
			return nil
		}

		// Naming format of the MD template: Current naming format for VSphereTemplate for MachineDeployment is {ACTUAL_TEMPLATE_NAME}-{KUBERNETES_VERSION}-{random-string}
		clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name] = mdInfastructureTemplateInfo{
			MDInfrastructureTemplateName: clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName + "-" +
				utils.ReplaceSpecialChars(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion) + "-" + utils.GenerateRandomID(5, true),
			MDInfrastructureTemplateNamespace: actualVsphereMachineTemplate.Namespace,
		}

		vsphereMachineTemplateForUpgrade := &capvv1beta1.VSphereMachineTemplate{}
		vsphereMachineTemplateForUpgrade.Name = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName
		vsphereMachineTemplateForUpgrade.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateNamespace
		vsphereMachineTemplateForUpgrade.Spec = actualVsphereMachineTemplate.DeepCopy().Spec
		vsphereMachineTemplateForUpgrade.Spec.Template.Spec.Template = clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateName
		vsphereMachineTemplateForUpgrade.Annotations = map[string]string{}
		vsphereMachineTemplateForUpgrade.Annotations[vmTemplateMoidKey] = clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateMOID

		// create template for each machine deployment object
		err = regionalClusterClient.CreateResource(vsphereMachineTemplateForUpgrade, vsphereMachineTemplateForUpgrade.Name, vsphereMachineTemplateForUpgrade.Namespace)
		if err != nil {
			return errors.Wrapf(err, "unable to create VSphereMachineTemplate for upgrade with name '%s' in namespace '%s'", vsphereMachineTemplateForUpgrade.Name, vsphereMachineTemplateForUpgrade.Namespace)
		}
	}
	return nil
}

func (c *TkgClient) createVsphereInfrastructureTemplateForUpgrade(regionalClusterClient clusterclient.Client, kcp *capikubeadmv1beta1.KubeadmControlPlane, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error

	vcClient, dcName, err := regionalClusterClient.GetVCClientAndDataCenter(
		clusterUpgradeConfig.ClusterName,
		clusterUpgradeConfig.ClusterNamespace,
		kcp.Spec.MachineTemplate.InfrastructureRef.Name,
		c.vcClientFactory)
	if err != nil {
		return errors.Wrap(err, "unable to create vsphere client")
	}
	tkrBom, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(clusterUpgradeConfig.UpgradeComponentInfo.TkrVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to get the BOM configuration of the TanzuKubernetesRelease %s", clusterUpgradeConfig.UpgradeComponentInfo.TkrVersion)
	}
	vSphereVM, err := vcClient.GetAndValidateVirtualMachineTemplate(
		tkrBom.GetOVAVersions(),
		clusterUpgradeConfig.UpgradeComponentInfo.TkrVersion,
		clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateName,
		dcName,
		c.TKGConfigReaderWriter(),
	)
	if err != nil {
		return errors.Wrap(err, "unable to get/verify vsphere template")
	}

	clusterUpgradeConfig.UpgradeComponentInfo.OsInfo = tkgconfigbom.OSInfo{
		Name:    vSphereVM.DistroName,
		Version: vSphereVM.DistroVersion,
		Arch:    vSphereVM.DistroArch,
	}

	clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateName = vSphereVM.Name
	clusterUpgradeConfig.UpgradeComponentInfo.VSphereVMTemplateMOID = vSphereVM.Moid

	if err := c.createVSphereControlPlaneMachineTemplate(regionalClusterClient, kcp, clusterUpgradeConfig); err != nil {
		return err
	}
	if err := c.createVSphereMachineDeploymentMachineTemplateForWorkers(regionalClusterClient, clusterUpgradeConfig); err != nil {
		return err
	}
	return nil
}

func (c *TkgClient) PatchKubernetesVersionToKubeadmControlPlane(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	log.V(6).Infof("Cluster Name: %s, Cluster Namespace %s", clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
	currentKCP, err := regionalClusterClient.GetKCPObjectForCluster(clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
	if err != nil {
		return errors.Wrapf(err, "unable to get KCP object to increase the kube-vip timeouts. Continuing upgrade with old parameters. ")
	}

	var newKCP *capikubeadmv1beta1.KubeadmControlPlane
	// If iaas == vsphere, attempt increasing kube-vip parameters
	if currentKCP.Spec.MachineTemplate.InfrastructureRef.Kind == constants.KindVSphereMachineTemplate {
		log.V(6).Infof("Kind %s", currentKCP.Spec.MachineTemplate.InfrastructureRef.Kind)
		newKCP, _ = c.UpdateKCPObjectWithIncreasedKubeVip(currentKCP)
		if newKCP != nil {
			currentKCP = newKCP
		}
	}

	// If unable to make kubevip updates, and the k8s version and the template name are unchanged, skip patching
	if newKCP == nil && clusterUpgradeConfig.ActualComponentInfo.KubernetesVersion == clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion &&
		clusterUpgradeConfig.ActualComponentInfo.KCPInfrastructureTemplateName == clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName {
		log.Infof("Skipping KubeadmControlPlane patch as kubernetes versions are already same %s", clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion)
		return nil
	}

	if !(clusterUpgradeConfig.ActualComponentInfo.KubernetesVersion == clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion &&
		clusterUpgradeConfig.ActualComponentInfo.KCPInfrastructureTemplateName == clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName) {
		log.Infof("Updating the KCP object with k8s version %s", clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion)

		currentKCP.Spec.Version = clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion
		currentKCP.Spec.MachineTemplate.InfrastructureRef.Name = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateName
		currentKCP.Spec.MachineTemplate.InfrastructureRef.Namespace = clusterUpgradeConfig.UpgradeComponentInfo.KCPInfrastructureTemplateNamespace
		currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.ImageRepository = clusterUpgradeConfig.UpgradeComponentInfo.ImageRepository
		currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.DNS.ImageRepository = clusterUpgradeConfig.UpgradeComponentInfo.CoreDNSImageRepository
		currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.DNS.ImageTag = clusterUpgradeConfig.UpgradeComponentInfo.CoreDNSImageTag
		currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ImageRepository = clusterUpgradeConfig.UpgradeComponentInfo.EtcdImageRepository
		currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ImageTag = clusterUpgradeConfig.UpgradeComponentInfo.EtcdImageTag
		if currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs == nil {
			currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs = map[string]string{}
		}
		for k, v := range clusterUpgradeConfig.UpgradeComponentInfo.EtcdExtraArgs {
			currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs[k] = v
		}
	}

	if semver.Compare(clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion, "v1.24.0") >= 0 {
		newKCP := c.configurePodSecurityStandard(currentKCP)
		if newKCP != nil {
			log.Infof("Enabling Pod Security Standard for KCP")
			currentKCP = newKCP
		}
	}

	pollOptions := &clusterclient.PollOptions{Interval: upgradePatchInterval, Timeout: upgradePatchTimeout}
	err = regionalClusterClient.UpdateResourceWithPolling(currentKCP, clusterUpgradeConfig.KCPObjectName, clusterUpgradeConfig.KCPObjectNamespace, pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to update the kubernetes version for kubeadm control plane nodes")
	}

	operationTimeout := 15 * time.Minute
	err = regionalClusterClient.PatchClusterWithOperationStartedStatus(clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace, clusterclient.OperationTypeUpgrade, operationTimeout)
	if err != nil {
		log.V(6).Infof("unable to patch cluster object with operation status, %s", err.Error())
	}

	return nil
}

func (c *TkgClient) patchKubernetesVersionToMachineDeployment(regionalClusterClient clusterclient.Client, clusterUpgradeConfig *ClusterUpgradeInfo) error {
	var err error
	for i := range clusterUpgradeConfig.MDObjects {
		if clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.Version != nil &&
			clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion == *clusterUpgradeConfig.MDObjects[i].Spec.Template.Spec.Version &&
			clusterUpgradeConfig.ActualComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName == clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName {
			log.Infof("Skipping MachineDeployment patch as kubernetes versions are already same %s", clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion)
			return nil
		}

		patchString := `{
			"spec": {
			  "template": {
				"spec": {
				  "version": "%s",
				  "infrastructureRef": {
					"name": "%s",
					"namespace": "%s"
				  }
				}
			  }
			}
		  }`

		patchKubernetesVersion := fmt.Sprintf(patchString,
			clusterUpgradeConfig.UpgradeComponentInfo.KubernetesVersion,
			clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateName,
			clusterUpgradeConfig.UpgradeComponentInfo.MDInfastructureTemplates[clusterUpgradeConfig.MDObjects[i].Name].MDInfrastructureTemplateNamespace)

		log.V(3).Infof("Applying MachineDeployment Patch: %s", patchKubernetesVersion)

		// Using polling to retry on any failed patch attempt.
		pollOptions := &clusterclient.PollOptions{Interval: upgradePatchInterval, Timeout: upgradePatchTimeout}
		err = regionalClusterClient.PatchResource(&capi.MachineDeployment{}, clusterUpgradeConfig.MDObjects[i].Name, clusterUpgradeConfig.MDObjects[i].Namespace, patchKubernetesVersion, types.MergePatchType, pollOptions)
		if err != nil {
			return errors.Wrap(err, "unable to update the kubernetes version for worker nodes")
		}
	}
	return nil
}

// handleKappControllerUpgrade contains upgrade logic required for kapp-controller.
func (c *TkgClient) handleKappControllerUpgrade(regionalClusterClient, currentClusterClient clusterclient.Client, upgradeClusterConfig *ClusterUpgradeInfo) error {
	// In TKG version prior to v1.3, kapp-controller could have been deployed by user as part of tkg-extensions deployment.
	// We need to delete the existing kapp-controller since a new kapp-controller will be installed from TKG v1.3 for addons management.
	if err := currentClusterClient.DeleteExistingKappController(); err != nil {
		return errors.Wrapf(err, "unable to delete existing kapp-controller")
	}

	// Update AWSCluster cniIngressRules to include kapp-controller API port only if CAPA is running on the management cluster
	if err := regionalClusterClient.GetResource(&corev1.Namespace{}, clusterclient.CAPAControllerNamespace, clusterclient.CAPAControllerNamespace, nil, nil); err != nil {
		// if capa-system namespace doesn't exist, then assume that updates to AWSCluster are not required during upgrade.
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "unable to check if Cluster API Provider for AWS is enabled")
		}
	} else {
		if err := regionalClusterClient.UpdateAWSCNIIngressRules(upgradeClusterConfig.ClusterName, upgradeClusterConfig.ClusterNamespace); err != nil {
			return errors.Wrapf(err, "unable to update AWS CNI ingress rules")
		}
	}

	return nil
}

func (c *TkgClient) getWorkloadClusterClient(clusterName, namespace string) (clusterclient.Client, error) {
	workloadClusterKubeConfigPath, err := utils.CreateTempFile("", "workload-kubeconfig")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create temporary file to save workload cluster kubeconfig")
	}
	workloadClusterCredentialsOptions := GetWorkloadClusterCredentialsOptions{
		ClusterName: clusterName,
		Namespace:   namespace,
		ExportFile:  workloadClusterKubeConfigPath,
	}
	context, _, err := c.GetWorkloadClusterCredentials(workloadClusterCredentialsOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get workload cluster credentials")
	}

	workloadClusterClient, err := clusterclient.NewClient(workloadClusterKubeConfigPath, context, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create workload cluster client")
	}

	return workloadClusterClient, nil
}

// verifyK8sVersion
// - verify the version format is correct
// - make sure this is an upgrade and not downgrade or the same version
func (c *TkgClient) verifyK8sVersion(clusterClient clusterclient.Client, newVersion string) error {
	currentVersion, err := clusterClient.GetKubernetesVersion()
	if err != nil {
		return errors.New("unable to get current kubernetes version for the cluster")
	}

	// Make sure this is an upgrade and not a downgrade
	compareResult, err := utils.CompareVMwareVersionStrings(currentVersion, newVersion)
	if err != nil {
		return errors.Wrapf(err, "error while comparing kubernetes versions %s,%s", newVersion, currentVersion)
	}

	if compareResult > 0 {
		return errors.Errorf("attempted to upgrade kubernetes from %s to %s. Kubernetes version downgrade is not allowed.", currentVersion, newVersion)
	}

	if !utils.CheckKubernetesUpgradeCompatibility(currentVersion, newVersion) {
		return errors.Errorf("Upgrading Kubernetes from %s to %s is not supported", currentVersion, newVersion)
	}

	return nil
}

func listCAPIv1alpha3Clusters(clusterClient clusterclient.Client) ([]capiv1alpha3.Cluster, error) {
	var clusterList capiv1alpha3.ClusterList
	if err := clusterClient.ListResources(&clusterList); err != nil {
		return []capiv1alpha3.Cluster{}, err
	}

	return clusterList.Items, nil
}

func (c *TkgClient) getRegionalClusterNameAndNamespace(clusterClient clusterclient.Client) (string, string, error) {
	var clusterName string
	var clusterNamespace string

	regionalClusterInfo, err := c.GetCurrentRegionContext()
	if err != nil {
		return clusterName, clusterNamespace, err
	}

	clusterName = regionalClusterInfo.ClusterName

	clusters, err := listCAPIv1alpha3Clusters(clusterClient)
	if err != nil {
		return clusterName, clusterNamespace, err
	}

	for i := range clusters {
		if clusterName == clusters[i].Name {
			clusterNamespace = clusters[i].Namespace
		}
	}

	if clusterNamespace == "" {
		return clusterName, clusterNamespace, errors.Errorf("unable to find namespace of management cluster object %s", clusterName)
	}

	return clusterName, clusterNamespace, nil
}

func (c *TkgClient) getAWSAMIIDForK8sVersion(regionalClusterClient clusterclient.Client, upgradeInfo *ClusterUpgradeInfo) error {
	awsClusterObject := &capav1beta2.AWSCluster{}
	if err := regionalClusterClient.GetResource(awsClusterObject, upgradeInfo.ClusterName, upgradeInfo.ClusterNamespace, nil, nil); err != nil {
		return errors.Wrap(err, "unable to retrieve aws cluster object to retrieve AMI settings")
	}

	if ami, ok := upgradeInfo.UpgradeComponentInfo.AwsRegionToAMIMap[awsClusterObject.Spec.Region]; ok {
		selectedAMI := tkgconfighelper.SelectAWSImageBasedonOSOptions(ami, c.TKGConfigReaderWriter())
		if selectedAMI == nil {
			return errors.Errorf("unable to find the AMI ID for AWSTemplate for region %s and kubernetes version %s, with the provided os option", awsClusterObject.Spec.Region, upgradeInfo.UpgradeComponentInfo.KubernetesVersion)
		}
		upgradeInfo.UpgradeComponentInfo.AwsAMIID = selectedAMI.ID
		upgradeInfo.UpgradeComponentInfo.OsInfo = selectedAMI.OSInfo
	}

	if upgradeInfo.UpgradeComponentInfo.AwsAMIID == "" {
		return errors.Errorf("unable to find the AMI ID for AWSTemplate for region %s and kubernetes version %s", awsClusterObject.Spec.Region, upgradeInfo.UpgradeComponentInfo.KubernetesVersion)
	}

	return nil
}

func (c *TkgClient) getCAPDImageForK8sVersion(upgradeInfo *ClusterUpgradeInfo) error {
	if upgradeInfo.UpgradeComponentInfo.CAPDImageName == "" {
		bomConfiguration, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
		if err != nil {
			return errors.Wrapf(err, "cannot get kind node image for kubernetes version %s", upgradeInfo.UpgradeComponentInfo.KubernetesVersion)
		}

		// TODO: write util for taking in bomConfiguration and returning correct
		// path so the replace can be tested
		kindNodeImage := bomConfiguration.Components["kubernetes-sigs_kind"][0].Images["kindNodeImage"]
		defaultImageRepo := bomConfiguration.ImageConfig.ImageRepository
		if kindNodeImage.ImageRepository != "" {
			defaultImageRepo = kindNodeImage.ImageRepository
		}

		upgradeInfo.UpgradeComponentInfo.CAPDImageName = fmt.Sprintf("%s/%s:%s", defaultImageRepo, kindNodeImage.ImagePath, kindNodeImage.Tag)
	}

	// we can keep everything about the image the same except for the version,
	// which will be updated

	updatedVersion := strings.ReplaceAll(upgradeInfo.UpgradeComponentInfo.KubernetesVersion, "+", "_")
	newImage, err := utils.ReplaceVersionInDockerImage(upgradeInfo.UpgradeComponentInfo.CAPDImageName, updatedVersion)
	if err != nil {
		return errors.Wrap(err, "could not replace version in kind image")
	}

	upgradeInfo.UpgradeComponentInfo.CAPDImageName = newImage

	return nil
}

func (c *TkgClient) configureOSOptionsForUpgrade(regionalClusterClient clusterclient.Client, options *UpgradeClusterOptions) error {
	if options.OSName == "" && options.OSVersion == "" && options.OSArch == "" {
		clusterObject := &capi.Cluster{}
		if err := regionalClusterClient.GetResource(clusterObject, options.ClusterName, options.Namespace, nil, nil); err != nil {
			return errors.Wrap(err, "unable to get cluster object")
		}

		// Get `osInfo` annotation from cluster object to determine
		// the default OS info to use during the upgrade
		annotations := clusterObject.GetAnnotations()
		osInfo := annotations["osInfo"]
		if osInfo != "" {
			osMetadataTotalValues := 3 // contains 3 values, name,version,arch
			osMetadata := strings.Split(osInfo, ",")
			if len(osMetadata) == osMetadataTotalValues {
				options.OSName = osMetadata[0]
				options.OSVersion = osMetadata[1]
				options.OSArch = osMetadata[2]
				log.V(3).Infof("Detected OS for cluster: %v %v %v", options.OSName, options.OSVersion, options.OSArch)
			}
		}

		// if this values are still empty meaning that this is old cluster created with
		// TKG CLI v1.2 which does not store OS information to the cluster
		// So, use TKG v1.2 Default values for OS for upgrading clusters
		if options.OSName == "" && options.OSVersion == "" && options.OSArch == "" {
			kcp, err := regionalClusterClient.GetKCPObjectForCluster(options.ClusterName, options.Namespace)
			if err != nil {
				return errors.Wrapf(err, "unable to find control plane node object for cluster %s", options.ClusterName)
			}

			provider := ""
			switch kcp.Spec.MachineTemplate.InfrastructureRef.Name {
			case constants.KindVSphereMachineTemplate:
				provider = constants.InfrastructureProviderVSphere
			case constants.KindAWSMachineTemplate:
				provider = constants.InfrastructureProviderAWS
			case constants.KindAzureMachineTemplate:
				provider = constants.InfrastructureProviderAzure
			case constants.KindDockerMachineTemplate:
			}

			osInfo := tkgconfighelper.GetDefaultOsOptionsForTKG12(provider)
			options.OSName = osInfo.Name
			options.OSVersion = osInfo.Version
			options.OSArch = osInfo.Arch
			log.V(3).Infof("Unable to detect current OS for the cluster. Using name:%v version:%v arch:%v", options.OSName, options.OSVersion, options.OSArch)
		}
	}

	log.V(3).Infof("Using OS options, name:%v version:%v arch:%v", options.OSName, options.OSVersion, options.OSArch)

	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSName, options.OSName)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSVersion, options.OSVersion)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSArch, options.OSArch)

	return nil
}
