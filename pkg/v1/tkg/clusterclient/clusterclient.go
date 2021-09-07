// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package clusterclient implements generic functions for talking to cluster
package clusterclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	sysrt "runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/yalp/jsonpath"
	appsv1 "k8s.io/api/apps/v1"
	betav1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	capav1alpha3 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capvv1alpha3 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	bootstrapv1alpha3 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
	capdv1alpha3 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/conditions"
	containerutil "sigs.k8s.io/cluster-api/util/container"
	"sigs.k8s.io/cluster-api/util/patch"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	capdiscovery "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
	tmcv1alpha1 "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/api/tmc/v1alpha1"
	azureclient "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/docker"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	telemetrymanifests "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/manifest/telemetry"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
	tkrconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
)

const (
	kubectlApplyRetryTimeout  = 30 * time.Second
	kubectlApplyRetryInterval = 5 * time.Second
	// DefaultKappControllerHostPort is the default kapp-controller port for it's extension apiserver
	DefaultKappControllerHostPort = 10100
)

// Client provides various aspects of interaction with a Kubernetes cluster provisioned by TKG
//go:generate counterfeiter -o ../fakes/clusterclient.go --fake-name ClusterClient . Client
type Client interface {
	// MergeAndUseConfig takes a kubeconfig as a string, merges it into the client's kubeconfig
	// path, and return current and previous kube contexts. The current context is also updated in said
	// path to use the new context added. This allows other client-side tools like kubectl, and
	// octant to interact with the cluster associated with the context without additional configuration.
	MergeAndUseConfigForCluster(kubeConfig []byte, overrideContextName string) (string, string, error)

	// MergeConfigForCluster merges a kubeconfig into the client's kubeconfig path.
	MergeConfigForCluster(kubeConfig []byte, mergeFile string) error

	// Apply applies a yaml string to a cluster
	Apply(string) error
	// Apply configuration to a cluster by filename
	ApplyFile(string) error
	// WaitForClusterInitialized waits for a cluster to be initialized so the kubeconfig file can be fetched
	WaitForClusterInitialized(clusterName string, namespace string) error
	// WaitForClusterReady for a cluster to be fully provisioned and so ready to be moved
	// If checkReplicas is true, will also ensure that the number of ready
	// replicas matches the expected number in the cluster's spec
	WaitForClusterReady(clusterName string, namespace string, checkReplicas bool) error
	// WaitForClusterDeletion waits for cluster object to be deleted
	WaitForClusterDeletion(clusterName string, namespace string) error
	// WaitForDeployment for a deployment to be fully available
	WaitForDeployment(deploymentName string, namespace string) error
	// WaitForAutoscalerDeployment waits for the autoscaler deployment to be available
	WaitForAutoscalerDeployment(deploymentName string, namespace string) error
	// WaitForAVIResourceCleanUp waits for the avi resource clean up finished
	WaitForAVIResourceCleanUp(statefulSetName, namespace string) error
	// WaitForPackageInstall waits for the package to be installed successfully
	WaitForPackageInstall(packageName, namespace string, packageInstallTimeout time.Duration) error
	// WaitK8sVersionUpdateForCPNodes waits for k8s version to be updated
	WaitK8sVersionUpdateForCPNodes(clusterName, namespace, kubernetesVersion string, workloadClusterClient Client) error
	// WaitK8sVersionUpdateForWorkerNodes waits for k8s version to be updated in all worker nodes
	WaitK8sVersionUpdateForWorkerNodes(clusterName, namespace, kubernetesVersion string, workloadClusterClient Client) error
	// GetKubeConfigForCluster returns the admin kube config for accessing the cluster
	GetKubeConfigForCluster(clusterName string, namespace string, pollOptions *PollOptions) ([]byte, error)
	// GetSecretValue returns the value for a given key in a Secret
	GetSecretValue(secretName, key, namespace string, pollOptions *PollOptions) ([]byte, error)
	// GetCurrentNamespace returns the namespace from the current context in the kubeconfig file
	GetCurrentNamespace() (string, error)
	// CreateNamespace creates namespace if missing
	CreateNamespace(name string) error
	// UseContext updates current-context in the kubeconfig file
	// also updates the clientset of clusterclient to point to correct cluster
	UseContext(contextName string) error

	// GetResource gets the kubernetes resource passed as reference either directly or with polling mechanism
	// resourceReference is a reference to resource struct you want to retrieve
	// resourceName is name of the resource to get
	// namespace is namespace in which the resource to be searched, if empty current namespace will be used
	// postVerify verifies the resource with some state once it is retrieved from kubernetes, pass nil if nothing to verify
	// pollOptions use this if you want to continuously poll for object if error occurs, pass nil if don't want polling
	// Note: Make sure resource you are retrieving is added into Scheme with init function below
	GetResource(resourceReference interface{}, resourceName, namespace string, postVerify PostVerifyrFunc, pollOptions *PollOptions) error
	// GetResourceList gets the list kubernetes resources passed as reference either directly or with polling mechanism
	// resourceReference is a reference to resource struct you want to retrieve
	// resourceName is name of the resource to get
	// namespace is namespace in which the resource to be searched, if empty current namespace will be used
	// postVerify verifies the resource with some state once it is retrieved from kubernetes, pass nil if nothing to verify
	// pollOptions use this if you want to continuously poll for object if error occurs, pass nil if don't want polling
	// Note: Make sure resource you are retrieving is added into Scheme with init function below
	GetResourceList(resourceReference interface{}, clusterName, namespace string, postVerify PostVerifyrFunc, pollOptions *PollOptions) error

	// ListResources lists the kubernetes resources, pass reference of the object you want to get
	// Note: Make sure resource you are retrieving is added into Scheme in init function below
	ListResources(resourceReference interface{}, option ...crtclient.ListOption) error
	// DeleteResource deletes the kubernetes resource, pass reference of the object you want to delete
	DeleteResource(resourceReference interface{}) error
	// PatchResource patches the kubernetes resource with procide patch string
	// resourceReference is a reference to resource struct you want to retrieve
	// resourceName is name of the resource to patch
	// namespace is namespace in which the resource to be searched, if empty current namespace will be used
	// patchJSONString is string representation of json of resource configuration
	// pollOptions use this if you want to continuously poll and patch the object if error occurs, pass nil if don't want polling
	PatchResource(resourceReference interface{}, resourceName, namespace, patchJSONString string, patchType types.PatchType, pollOptions *PollOptions) error
	// CreateResource creates the kubernetes resource
	// resourceReference is a reference to resource struct you want to create
	// resourceName is name of the resource to create
	// namespace is namespace in which the resource to be created, if empty current namespace will be used
	// opts is options for resource creation
	CreateResource(resourceReference interface{}, resourceName, namespace string, opts ...crtclient.CreateOption) error
	// UpdateResource updates the kubernetes resource
	// resourceReference is a reference to resource struct you want to create
	// resourceName is name of the resource to create
	// namespace is namespace in which the resource to be created, if empty current namespace will be used
	// opts is options for resource creation
	UpdateResource(resourceReference interface{}, resourceName, namespace string, opts ...crtclient.UpdateOption) error
	// ExportCurrentKubeconfigToFile saves the current kubeconfig to temporary file and returns the file
	ExportCurrentKubeconfigToFile() (string, error)
	// GetCurrentKubeconfigFile returns currently used kubeconfig file path based on default loading rules
	GetCurrentKubeconfigFile() string
	// GetCurrentClusterName returns the current clusterName based on current context from kubeconfig file
	// If context parameter is not empty, then return clusterName corresponding to the context
	GetCurrentClusterName(context string) (string, error)
	// GetCurrentKubeContext returns the current kube xontext
	GetCurrentKubeContext() (string, error)
	// IsRegionalCluster() checks if the current kube context point to a management cluster
	IsRegionalCluster() error
	// GetRegionalClusterDefaultProviderName returns the default provider name of provider type
	GetRegionalClusterDefaultProviderName(providerType clusterctlv1.ProviderType) (string, error)
	// ListClusters lists workload cluster managed by a management cluster in a given namespace
	ListClusters(namespace string) ([]capi.Cluster, error)
	// DeleteCluster deletes cluster in the given namespace
	DeleteCluster(clusterName string, namespace string) error
	// GetKubernetesVersion gets kubernetes server version for a given cluster
	GetKubernetesVersion() (string, error)
	// GetMDObjectForCluster gets machine deployment object of worker nodes for cluster
	GetMDObjectForCluster(clusterName string, namespace string) ([]capi.MachineDeployment, error)
	// GetClusterControlPlaneNodeObject gets cluster control plane node for cluster
	GetKCPObjectForCluster(clusterName string, namespace string) (*controlplanev1.KubeadmControlPlane, error)
	// GetMachineObjectsForCluster gets control-plane machine and worker machine lists for cluster
	GetMachineObjectsForCluster(clusterName string, namespace string) (map[string]capi.Machine, map[string]capi.Machine, error)
	// UpdateReplicas updates the replica count for the given resource
	UpdateReplicas(resourceReference interface{}, resourceName, resourceNameSpace string, replicaCount int32) error
	// IsPacificRegionalCluster checks if the cluster pointed to by kubeconfig  is Pacific management cluster(supervisor)
	IsPacificRegionalCluster() (bool, error)
	// WaitForPacificCluster waits for the Vsphere-pacific provider workload cluster to be fully provisioned
	WaitForPacificCluster(clusterName string, namespace string, version string) error
	// ListPacificClusterObjects returns TanzuKubernetesClusterList object
	ListPacificClusterObjects(apiVersion string, listOptions *crtclient.ListOptions) ([]interface{}, error)
	// ScalePacificClusterControlPlane scales Pacific workload cluster control plane
	ScalePacificClusterControlPlane(clusterName, namespace, apiVersion string, controlPlaneCount int32) error
	// ScalePacificClusterWorkerNodes scales Pacific workload cluster worker nodes
	ScalePacificClusterWorkerNodes(clusterName, namespace, apiVersion string, workersCount int32) error
	// LoadCurrentKubeconfigBytes returns the current kubeconfig with current regional context in bytes
	LoadCurrentKubeconfigBytes() ([]byte, error)

	// CloneWithTimeout returns a new client with the same attributes of the current one except for get client timeout settings
	CloneWithTimeout(getClientTimeout time.Duration) Client
	// GetVCClientAndDataCenter returns vsphere client and datacenter name by reading on cluster resources
	GetVCClientAndDataCenter(clusterName, clusterNamespace, vsphereMachineTemplateObjectName string) (vc.Client, string, error)
	// PatchK8SVersionToPacificCluster patches the Pacific TKC object to update the k8s version on the cluster
	PatchK8SVersionToPacificCluster(clusterName, namespace, apiVersion string, kubernetesVersion string) error
	// WaitForPacificClusterK8sVersionUpdate waits for the Pacific TKC cluster to update k8s version
	WaitForPacificClusterK8sVersionUpdate(clusterName, namespace, apiVersion, kubernetesVersion string) error
	// PatchClusterWithOperationStartedStatus applies patch to cluster objects annotations
	// with operation status information which includes type of operation, start time and timeout
	// This information along with operation last observed timestamp will be used to determine
	// stalled state of the cluster
	PatchClusterWithOperationStartedStatus(clusterName, namespace, operationType string, timeout time.Duration) error
	// PatchClusterObjectWithTKGVersion applies patch to cluster objects based on given tkgVersion string
	PatchClusterObjectWithTKGVersion(clusterName, clusterNamespace, tkgVersion string) error
	// GetManagementClusterTKGVersion returns the TKG version of a management cluster based on the
	// annotation value present in cluster object
	GetManagementClusterTKGVersion(mgmtClusterName, clusterNamespace string) (string, error)
	// PatchCalicoNodeDaemonSetWithNewNodeSelector patches calico daemonset with new nodeSelector
	PatchCalicoNodeDaemonSetWithNewNodeSelector(selectorKey, selectorValue string) error
	// PatchCalicoKubeControllerDeploymentWithNewNodeSelector patches calico-kube-controller deployment with new nodeSelector
	PatchCalicoKubeControllerDeploymentWithNewNodeSelector(selectorKey, selectorValue string) error
	// PatchImageRepositoryInKubeProxyDaemonSet updates kubeproxy daemonset with new/custom image repository
	PatchImageRepositoryInKubeProxyDaemonSet(newImageRepository string) error
	// PatchClusterAPIAWSControllersToUseEC2Credentials ensures that the Cluster API Provider AWS
	// controller is pinned to control plane nodes and is running without static credentials such
	// that Cluster API AWS runs using the EC2 instance profile attached to the control plane node.
	// This is done by zeroing out the credentials secret for CAPA, causing the AWS SDK to fall back
	// to the default credential provider chain. We additionally patch the deployment to ensure
	// the controller has node affinity to only run on the control plane nodes.
	// This should NOT be used when running Cluster API Provider AWS on managed control planes, e.g. EKS
	PatchClusterAPIAWSControllersToUseEC2Credentials() error
	// PatchCoreDNSImageRepositoryInKubeadmConfigMap updates kubeadm-config configMap with new/custom image repository
	PatchCoreDNSImageRepositoryInKubeadmConfigMap(newImageRepository string) error
	// PatchClusterObjectWithOptionalMetadata applies patch to cluster objects based on given optional metadata
	// under the key provided as metadataKey (e.g. annotations, labels) where the value is in the form of a
	// map[string]string (e.g. [Description]some-description) where the key is the name of the metadata property.
	PatchClusterObjectWithOptionalMetadata(clusterName, clusterNamespace, metadataKey string, metadata map[string]string) (string, error)
	// PatchClusterObject patches cluster object with specified json patch
	PatchClusterObject(clusterName, clusterNamespace string, patchJSONString string) error
	// DeleteExistingKappController deletes the kapp-controller that already exists in the cluster.
	DeleteExistingKappController() error
	// UpdateAWSCNIIngressRules updates the cniIngressRules field for the AWSCluster resource.
	UpdateAWSCNIIngressRules(clusterName, clusterNamespace string) error
	// AddCEIPTelemetryJob creates telemetry cronjob component on cluster
	AddCEIPTelemetryJob(clusterName, providerName string, bomConfig *tkgconfigbom.BOMConfiguration, isProd, labels, httpProxy, httpsProxy, noProxy string) error
	// RemoveCEIPTelemetryJob deletes telemetry cronjob component on cluster
	RemoveCEIPTelemetryJob(clusterName string) error
	// HasCEIPTelemetryJob checks if telemetry cronjob component is on cluster
	HasCEIPTelemetryJob(clusterName string) (bool, error)
	// GetPacificTKCAPIVersion gets the Pacific TKC API version
	GetPacificTKCAPIVersion() (string, error)
	// GetPacificTanzuKubernetesReleases returns the list of TanzuKubernetesRelease versions if TKr object is available in TKGS
	GetPacificTanzuKubernetesReleases() ([]string, error)
	// GetVCCredentialsFromSecret gets the vSphere username and password used to deploy the cluster
	GetVCCredentialsFromSecret(string) (string, string, error)
	// GetVCServer gets the vSphere server that used to deploy the cluster
	GetVCServer() (string, error)
	// GetAWSEncodedCredentialsFromSecret gets the AWS base64 credentials used to deploy the cluster
	GetAWSCredentialsFromSecret() (string, error)
	// GetAzureCredentialsFromSecret gets the Azure base64 credentials used to deploy the cluster
	GetAzureCredentialsFromSecret() (azureclient.Credentials, error)
	// UpdateCapvManagerBootstrapCredentialsSecret updates the vsphere creds used by the capv provider
	UpdateCapvManagerBootstrapCredentialsSecret(username string, password string) error
	// UpdateVsphereIdentityRefSecret updates vsphere cluster identityRef secret
	UpdateVsphereIdentityRefSecret(clusterName, namespace, username, password string) error
	// UpdateVsphereCloudProviderCredentialsSecret updates the vsphere creds used by the vsphere cloud provider
	UpdateVsphereCloudProviderCredentialsSecret(clusterName string, namespace string, username string, password string) error
	// UpdateVsphereCsiConfigSecret updates the vsphere csi config secret
	UpdateVsphereCsiConfigSecret(clusterName string, namespace string, username string, password string) error
	// GetClientSet gets one clientset used to generate objects list
	GetClientSet() CrtClient
	// GetPinnipedIssuerURLAndCA fetches Pinniped supervisor IssuerURL and IssuerCA data from management cluster
	GetPinnipedIssuerURLAndCA() (string, string, error)
	// GetTanzuKubernetesReleases returns the TKr's with 'tkrName' prefix match. If tkrName is not provided it returns all the available TKr's
	GetTanzuKubernetesReleases(tkrName string) ([]runv1alpha1.TanzuKubernetesRelease, error)
	// GetBomConfigMap returns configmap associated w3ith the tkrNameLabel
	GetBomConfigMap(tkrNameLabel string) (corev1.ConfigMap, error)
	// GetClusterInfrastructure gets cluster infrastructure name like VSphereCluster, AWSCluster, AzureCluster
	GetClusterInfrastructure() (string, error)
	// ActivateTanzuKubernetesReleases activates TanzuKubernetesRelease
	ActivateTanzuKubernetesReleases(tkrName string) error
	// DeactivateTanzuKubernetesReleases deactivates TanzuKubernetesRelease
	DeactivateTanzuKubernetesReleases(tkrName string) error
	// IsClusterRegisteredToTMC returns true if cluster is registered to Tanzu Mission Control
	IsClusterRegisteredToTMC() (bool, error)
}

// PollOptions is options for polling
type PollOptions struct {
	Interval time.Duration
	Timeout  time.Duration
}

// NewPollOptions returns new poll options
func NewPollOptions(interval, timeout time.Duration) *PollOptions {
	return &PollOptions{Interval: interval, Timeout: timeout}
}

//go:generate counterfeiter -o ../fakes/crtclusterclient.go --fake-name CRTClusterClient . CrtClient
//go:generate counterfeiter -o ../fakes/clusterclient.go --fake-name ClusterClient . Client
//go:generate counterfeiter -o ../fakes/discoveryclusterclient.go --fake-name DiscoveryClient . DiscoveryClient

// CrtClient clientset interface
type CrtClient interface {
	crtclient.Client
}

// DiscoveryClient discovery client interface
type DiscoveryClient interface {
	discovery.DiscoveryInterface
}

type client struct {
	clientSet                 CrtClient
	discoveryClient           DiscoveryClient
	kubeConfigPath            string
	currentContext            string
	poller                    Poller
	crtClientFactory          CrtClientFactory
	discoveryClientFactory    DiscoveryClientFactory
	configLoadingRules        *clientcmd.ClientConfigLoadingRules
	getClientInterval         time.Duration
	getClientTimeout          time.Duration
	operationTimeout          time.Duration
	verificationClientFactory *VerificationClientFactory
}

// constants regarding timeout and configs
const (
	operationDefaultTimeout           = 30 * time.Minute
	CheckResourceInterval             = 5 * time.Second
	CheckClusterInterval              = 10 * time.Second
	getClientDefaultInterval          = 10 * time.Second
	getClientDefaultTimeout           = 5 * time.Minute
	CheckAutoscalerDeploymentTimeout  = 2 * time.Minute
	AVIResourceCleanupTimeout         = 2 * time.Minute
	PackageInstallPollInterval        = 10 * time.Second
	PackageInstallTimeout             = 10 * time.Minute
	kubeConfigSecretSuffix            = "kubeconfig"
	kubeConfigDataField               = "value"
	embeddedTelemetryConfigYamlPrefix = "pkg/manifest/telemetry/config-"
	telemetryBomImagesMapKey          = "tkgTelemetryImage"
	prodTelemetryPath                 = "https://scapi.vmware.com/sc/api/collectors/tkg-telemetry.v1.4.0/batch"
	stageTelemetryPath                = "https://scapi-stg.vmware.com/sc/api/collectors/tkg-telemetry.v1.4.0/batch"
	statusRunning                     = "running"
)

const annotationPatchFormat = `
{
	"metadata": {
		"annotations": {
			"%s" : "%s"
		}
	}
}`

var providerTypes = []clusterctlv1.ProviderType{
	clusterctlv1.CoreProviderType,
	clusterctlv1.BootstrapProviderType,
	clusterctlv1.InfrastructureProviderType,
	clusterctlv1.ControlPlaneProviderType,
}

var (
	ctx    = context.Background()
	scheme = runtime.NewScheme()
)

func init() {
	_ = capi.AddToScheme(scheme)
	_ = capiv1alpha2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = clusterctlv1.AddToScheme(scheme)
	_ = controlplanev1.AddToScheme(scheme)
	_ = capvv1alpha3.AddToScheme(scheme)
	_ = capav1alpha3.AddToScheme(scheme)
	_ = capzv1alpha3.AddToScheme(scheme)
	_ = capdv1alpha3.AddToScheme(scheme)
	_ = bootstrapv1alpha3.AddToScheme(scheme)
	_ = runv1alpha1.AddToScheme(scheme)
	_ = betav1.AddToScheme(scheme)
	_ = tmcv1alpha1.AddToScheme(scheme)
	_ = extensionsV1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = addonsv1.AddToScheme(scheme)
	_ = runv1alpha1.AddToScheme(scheme)
	_ = kappipkg.AddToScheme(scheme)
}

// ClusterStatusInfo defines the cluster status involving all main components
type ClusterStatusInfo struct {
	KubernetesVersion    string
	ClusterObject        *capi.Cluster
	KCPObject            *controlplanev1.KubeadmControlPlane
	MDObjects            []capi.MachineDeployment
	CPMachineObjects     map[string]capi.Machine
	WorkerMachineObjects map[string]capi.Machine
	RetrievalError       error
}

// MergeAndUseConfigForCluster merges a provided kubeConfig byte slice and
// merge it into the client's kubeconfig path and returns previous and current contexts
func (c *client) MergeAndUseConfigForCluster(kubeConfig []byte, overrideContext string) (string, string, error) {
	// TODO: Support custom kube context name when merging kube configuration of cluster
	prevContext, err := c.GetCurrentKubeContext()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get current context before merging kubeconfig")
	}
	err = c.MergeConfigForCluster(kubeConfig, "")
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to merge kubeconfig of cluster")
	}
	currentContext, err := getCurrentContextFromKubeConfig(kubeConfig)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to get current context from new kubeconfig")
	}
	err = c.UseContext(currentContext)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to set context after updating kubeconfig")
	}
	return currentContext, prevContext, nil
}

func (c *client) GetClientSet() CrtClient {
	return c.clientSet
}

func (c *client) GetKubeConfigPath() string {
	return c.kubeConfigPath
}

// Apply runs kubectl apply every `interval` until it succeeds or a timeout is reached.
func (c *client) Apply(yaml string) error {
	_, err := c.poller.PollImmediateWithGetter(kubectlApplyRetryInterval, kubectlApplyRetryTimeout, func() (interface{}, error) {
		return nil, c.kubectlApply(yaml)
	})
	return err
}

// Apply runs kubectl apply on a file/url every `interval` until it succeeds or a timeout is reached.
func (c *client) ApplyFile(filePath string) error {
	_, err := c.poller.PollImmediateWithGetter(kubectlApplyRetryInterval, kubectlApplyRetryTimeout, func() (interface{}, error) {
		return nil, c.kubectlApplyFile(filePath)
	})
	return err
}

func (c *client) patchClusterWithOperationObservedTimestamp(clusterName, namespace string) {
	patchAnnotations := fmt.Sprintf(annotationPatchFormat, TKGOperationLastObservedTimestampKey, time.Now().UTC().String())
	err := c.PatchClusterObject(clusterName, namespace, patchAnnotations)
	if err != nil {
		log.V(6).Infof("unable to patch cluster object with operation status, %s", err.Error())
	}
}

func (c *client) WaitForClusterInitialized(clusterName, namespace string) error {
	var err error
	var currentClusterInfo ClusterStatusInfo
	var lastClusterInfo ClusterStatusInfo
	unchangedCounter := 0
	interval := 15 * time.Second

	getterFunc := func() (interface{}, error) {
		currentClusterInfo = c.GetClusterStatusInfo(clusterName, namespace, nil)
		err = currentClusterInfo.RetrievalError

		if err == nil {
			// If cluster's ReadyCondition is False and severity is Error, it implies non-retriable error, so return error
			if conditions.IsFalse(currentClusterInfo.ClusterObject, capi.ReadyCondition) &&
				(*conditions.GetSeverity(currentClusterInfo.ClusterObject, capi.ReadyCondition) == capi.ConditionSeverityError) {
				return true, errors.Errorf("cluster creation failed, reason:'%s', message:'%s'",
					conditions.GetReason(currentClusterInfo.ClusterObject, capi.ReadyCondition),
					conditions.GetMessage(currentClusterInfo.ClusterObject, capi.ReadyCondition))
			}
			// Could have checked cluster's ReadyCondition is True which is currently aggregation of ControlPlaneReadyCondition
			// and InfrastructureReadyCondition, however in future if capi adds WorkersReadyCondition into aggregation, it would
			// hold this method to wait till the workers are also ready which is not necessary for getting kubeconfig secret
			err = VerifyClusterInitialized(currentClusterInfo.ClusterObject)
			if err == nil {
				return false, nil
			}
		}

		if isClusterStateChanged(&lastClusterInfo, &currentClusterInfo) ||
			isClusterStateChangedForMD(&lastClusterInfo, &currentClusterInfo) {
			unchangedCounter = 0
			// Patch cluster object with updated timestamp information
			// This timestamp will be used with operation timeout to determine
			// stalled state of the cluster
			// TODO(anuj): consider reducing the number of patches done to the cluster
			// by batching a couple of those that happen close to one another.
			// Even a timestamp patch spaced at least 2-3 minutes apart will not affect
			// the correctness of this algo, but can potentially reduce patches by quite
			// a bit if there are quite a number of changes detected during that span.
			c.patchClusterWithOperationObservedTimestamp(clusterName, namespace)
		} else {
			unchangedCounter++
			log.V(7).Infof("cluster state is unchanged %v", unchangedCounter)
		}

		lastClusterInfo = currentClusterInfo

		// if unchanged for operationTimeout(30 min default), return error
		if interval*time.Duration(unchangedCounter) > c.operationTimeout {
			return true, errors.Wrap(err, "timed out waiting for cluster creation to complete")
		}

		return false, err
	}

	return c.poller.PollImmediateInfiniteWithGetter(interval, getterFunc)
}

func (c *client) WaitForClusterReady(clusterName, namespace string, checkAllReplicas bool) error {
	if err := c.GetResource(&capi.Cluster{}, clusterName, namespace, VerifyClusterReady, &PollOptions{Interval: CheckClusterInterval, Timeout: c.operationTimeout}); err != nil {
		return err
	}
	if checkAllReplicas {
		// Check and wait only for MD replicas as KCP replicas would be checked in above VerifyClusterReady() since the KCP Ready condition is mirrored into cluster ControPlaneReady condition
		if err := c.GetResourceList(&capi.MachineDeploymentList{}, clusterName, namespace, VerifyMachineDeploymentsReplicas, &PollOptions{Interval: CheckClusterInterval, Timeout: c.operationTimeout}); err != nil {
			return err
		}
	}
	if err := c.GetResourceList(&capi.MachineList{}, clusterName, namespace, VerifyMachinesReady, &PollOptions{Interval: CheckClusterInterval, Timeout: c.operationTimeout}); err != nil {
		return err
	}
	return nil
}

func (c *client) WaitForClusterDeletion(clusterName, namespace string) error {
	return c.WaitForResourceDeletion(&capi.Cluster{}, clusterName, namespace, nil, &PollOptions{Interval: CheckClusterInterval, Timeout: c.operationTimeout})
}

func (c *client) WaitForResourceDeletion(resourceReference interface{}, resourceName, namespace string, postVerify PostVerifyrFunc, pollOptions *PollOptions) error {
	var err error

	if pollOptions == nil {
		return errors.New("missing pollOptions")
	}
	if namespace == "" {
		if namespace, err = c.GetCurrentNamespace(); err != nil {
			return err
		}
	}

	// get the runtime object from interface
	obj, err := c.getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}

	log.V(4).Infof("Waiting for %s resource of type %s to be deleted", resourceName, reflect.TypeOf(resourceReference))
	_, err = c.poller.PollImmediateWithGetter(pollOptions.Interval, pollOptions.Timeout, func() (interface{}, error) {
		getErr := c.get(resourceName, namespace, obj, postVerify)
		if getErr != nil {
			// if error is not found, means the resource is deleted and stop polling
			if apierrors.IsNotFound(getErr) {
				return nil, nil
			}
			return nil, getErr
		}
		return nil, errors.New("resource is still present")
	})
	return err
}

func (c *client) WaitForDeployment(deploymentName, namespace string) error {
	return c.GetResource(&appsv1.Deployment{}, deploymentName, namespace, VerifyDeploymentAvailable, &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout})
}

func (c *client) WaitForAutoscalerDeployment(deploymentName, namespace string) error {
	return c.GetResource(&appsv1.Deployment{}, deploymentName, namespace, VerifyAutoscalerDeploymentAvailable, &PollOptions{Interval: CheckResourceInterval, Timeout: CheckAutoscalerDeploymentTimeout})
}

func (c *client) WaitForAVIResourceCleanUp(statefulSetName, namespace string) error {
	return c.GetResource(&appsv1.StatefulSet{}, statefulSetName, namespace, VerifyAVIResourceCleanupFinished, &PollOptions{Interval: CheckResourceInterval, Timeout: AVIResourceCleanupTimeout})
}

func (c *client) WaitForPackageInstall(packageName, namespace string, packageInstallTimeout time.Duration) error {
	if packageInstallTimeout == 0 {
		packageInstallTimeout = PackageInstallTimeout
	}
	return c.GetResource(&kappipkg.PackageInstall{}, packageName, namespace, VerifyPackageInstallReconciledSuccessfully, &PollOptions{Interval: PackageInstallPollInterval, Timeout: packageInstallTimeout})
}

func verifyKubernetesUpgradeForCPNodes(clusterStatusInfo *ClusterStatusInfo, newK8sVersion string) error {
	if clusterStatusInfo.RetrievalError != nil {
		return clusterStatusInfo.RetrievalError
	}

	clusterObj := clusterStatusInfo.ClusterObject
	if !conditions.IsTrue(clusterObj, capi.ControlPlaneReadyCondition) {
		return errors.Errorf("control-plane is still being upgraded, reason:'%s', message:'%s' ",
			conditions.GetReason(clusterObj, capi.ControlPlaneReadyCondition), conditions.GetMessage(clusterObj, capi.ControlPlaneReadyCondition))
	}

	if clusterStatusInfo.KubernetesVersion != newK8sVersion {
		return errors.Errorf("waiting for kubernetes version update, current kubernetes version %s but expecting %s", clusterStatusInfo.KubernetesVersion, newK8sVersion)
	}

	return nil
}

func verifyKubernetesUpgradeForWorkerNodes(clusterStatusInfo *ClusterStatusInfo, newK8sVersion string) error {
	if clusterStatusInfo.RetrievalError != nil {
		return clusterStatusInfo.RetrievalError
	}

	var desiredReplica int32 = 1
	errList := []error{}

	for i := range clusterStatusInfo.MDObjects {
		if clusterStatusInfo.MDObjects[i].Spec.Replicas != nil {
			desiredReplica = *clusterStatusInfo.MDObjects[i].Spec.Replicas
		}
		isReplicasUpgraded := clusterStatusInfo.MDObjects[i].Status.ReadyReplicas == desiredReplica && clusterStatusInfo.MDObjects[i].Status.UpdatedReplicas == desiredReplica && clusterStatusInfo.MDObjects[i].Status.Replicas == desiredReplica
		if !isReplicasUpgraded {
			err := errors.Errorf("worker nodes are still being upgraded for MachineDeployment '%s', DesiredReplicas=%v Replicas=%v ReadyReplicas=%v UpdatedReplicas=%v",
				clusterStatusInfo.MDObjects[i].Name, desiredReplica, clusterStatusInfo.MDObjects[i].Status.Replicas, clusterStatusInfo.MDObjects[i].Status.ReadyReplicas, clusterStatusInfo.MDObjects[i].Status.UpdatedReplicas)
			errList = append(errList, err)
		}
	}

	if len(errList) != 0 {
		return kerrors.NewAggregate(errList)
	}

	unupgradedMachineList := []string{}
	for i := range clusterStatusInfo.WorkerMachineObjects {
		if clusterStatusInfo.WorkerMachineObjects[i].Spec.Version == nil || *clusterStatusInfo.WorkerMachineObjects[i].Spec.Version != newK8sVersion {
			unupgradedMachineList = append(unupgradedMachineList, clusterStatusInfo.WorkerMachineObjects[i].Name)
		}
	}

	if len(unupgradedMachineList) > 0 {
		sort.Strings(unupgradedMachineList)
		return errors.Errorf("worker machines %v are still not upgraded", unupgradedMachineList)
	}

	return nil
}

// isClusterStateChanged functions verifies if the cluster object ready conditions(includes InfrastructureReady  and ControlplaneReady) have
// any transitions since last observation
func isClusterStateChanged(lastClusterInfo, curClusterInfo *ClusterStatusInfo) bool {
	if curClusterInfo.RetrievalError != nil ||
		lastClusterInfo.ClusterObject == nil ||
		curClusterInfo.ClusterObject == nil {
		return false
	}

	// If the ReadyCondition's lastTransitionTime is updated it implies there is some state change
	if !conditions.GetLastTransitionTime(curClusterInfo.ClusterObject, capi.ReadyCondition).Equal(
		conditions.GetLastTransitionTime(lastClusterInfo.ClusterObject, capi.ReadyCondition)) {
		return true
	}

	return false
}

// isClusterStateChangedForKCP functions verifies if the cluster object ControlplaneReady condition has
// any transitions since last observation
func isClusterStateChangedForKCP(lastClusterInfo, curClusterInfo *ClusterStatusInfo) bool {
	if curClusterInfo.RetrievalError != nil ||
		lastClusterInfo.ClusterObject == nil ||
		curClusterInfo.ClusterObject == nil {
		return false
	}
	// If the ControlPlaneReadyCondition's lastTransitionTime is updated it implies there is some state change
	if !conditions.GetLastTransitionTime(curClusterInfo.ClusterObject, capi.ControlPlaneReadyCondition).Equal(
		conditions.GetLastTransitionTime(lastClusterInfo.ClusterObject, capi.ControlPlaneReadyCondition)) {
		return true
	}
	return false
}

func isClusterStateChangedForMD(lastClusterInfo, curClusterInfo *ClusterStatusInfo) bool { // nolint:gocyclo
	if curClusterInfo.RetrievalError != nil ||
		lastClusterInfo.MDObjects == nil ||
		curClusterInfo.MDObjects == nil ||
		len(lastClusterInfo.WorkerMachineObjects) == 0 ||
		len(curClusterInfo.WorkerMachineObjects) == 0 {
		return false
	}

	compareMDStatus := func(lastMDObject capi.MachineDeployment, curMDObject capi.MachineDeployment) bool {
		if lastMDObject.Status.Replicas != curMDObject.Status.Replicas ||
			lastMDObject.Status.ReadyReplicas != curMDObject.Status.ReadyReplicas ||
			lastMDObject.Status.UpdatedReplicas != curMDObject.Status.UpdatedReplicas ||
			len(lastClusterInfo.WorkerMachineObjects) != len(curClusterInfo.WorkerMachineObjects) {
			return true
		}
		return false
	}

	for i := range lastClusterInfo.MDObjects {
		for j := range curClusterInfo.MDObjects {
			if lastClusterInfo.MDObjects[i].Name == curClusterInfo.MDObjects[j].Name && compareMDStatus(lastClusterInfo.MDObjects[i], curClusterInfo.MDObjects[j]) {
				return true
			}
		}
	}

	// As machines are stored in a map as "Name-Phase" key->Machine pair
	// if the key does not exist in lastClusterInfo.WorkerMachineObjects that mean
	// state of the machine object is changed
	for curMachineNamePhase := range curClusterInfo.WorkerMachineObjects {
		if _, found := lastClusterInfo.WorkerMachineObjects[curMachineNamePhase]; !found {
			return true
		}
	}

	return false
}

func (c *client) PatchClusterWithOperationStartedStatus(clusterName, namespace, operationType string, timeout time.Duration) error {
	currentTimestamp := time.Now().UTC().String()
	operationStatus := OperationStatus{
		Operation:               operationType,
		OperationTimeout:        int(timeout.Seconds()),
		OperationStartTimestamp: currentTimestamp,
	}
	operationStatusBytes, err := json.Marshal(operationStatus)
	if err != nil {
		return err
	}
	operationStatusString := strings.ReplaceAll((string(operationStatusBytes)), "\"", "\\\"")

	patchFormat := `
	{
		"metadata": {
			"annotations": {
				"%s" : "%s",
				"%s" : "%s"
			}
		}
	}`
	patchAnnotations := fmt.Sprintf(patchFormat, TKGOperationInfoKey, operationStatusString, TKGOperationLastObservedTimestampKey, currentTimestamp)
	log.V(6).Infof("patch cluster object with operation status: %s", patchAnnotations)
	err = c.PatchClusterObject(clusterName, namespace, patchAnnotations)
	if err != nil {
		return errors.Wrap(err, "unable to patch cluster object with operation status")
	}
	return nil
}

func (c *client) WaitK8sVersionUpdateForCPNodes(clusterName, namespace, newK8sVersion string, workloadClusterClient Client) error {
	return c.waitK8sVersionUpdateGeneric(clusterName, namespace, newK8sVersion, workloadClusterClient, true)
}

func (c *client) WaitK8sVersionUpdateForWorkerNodes(clusterName, namespace, newK8sVersion string, workloadClusterClient Client) error {
	return c.waitK8sVersionUpdateGeneric(clusterName, namespace, newK8sVersion, workloadClusterClient, false)
}

func (c *client) waitK8sVersionUpdateGeneric(clusterName, namespace, newK8sVersion string, workloadClusterClient Client, isCP bool) error {
	verifyKubernetesUpgradeFunc := verifyKubernetesUpgradeForCPNodes
	isClusterStateChangedFunc := isClusterStateChangedForKCP
	if !isCP {
		verifyKubernetesUpgradeFunc = verifyKubernetesUpgradeForWorkerNodes
		isClusterStateChangedFunc = isClusterStateChangedForMD
	}

	// client can have additional verificationClientFactory mainly for unit testing purpose
	if c.verificationClientFactory != nil && c.verificationClientFactory.VerifyKubernetesUpgradeFunc != nil {
		verifyKubernetesUpgradeFunc = c.verificationClientFactory.VerifyKubernetesUpgradeFunc
	}

	var err error
	var curClusterInfo ClusterStatusInfo
	var lastClusterInfo ClusterStatusInfo
	unchangedCounter := 0
	interval := 15 * time.Second
	timeout := c.operationTimeout

	getterFunc := func() (interface{}, error) {
		curClusterInfo = c.GetClusterStatusInfo(clusterName, namespace, workloadClusterClient)

		// If cluster's ReadyCondition is False and severity is Error, it implies non-retriable error, so return error
		if conditions.IsFalse(curClusterInfo.ClusterObject, capi.ReadyCondition) &&
			(*conditions.GetSeverity(curClusterInfo.ClusterObject, capi.ReadyCondition) == capi.ConditionSeverityError) {
			return true, errors.Errorf("kubernetes version update failed, reason:'%s', message:'%s' ",
				conditions.GetReason(curClusterInfo.ClusterObject, capi.ReadyCondition),
				conditions.GetMessage(curClusterInfo.ClusterObject, capi.ReadyCondition))
		}
		err = verifyKubernetesUpgradeFunc(&curClusterInfo, newK8sVersion)
		if err == nil {
			return false, nil
		}

		if isClusterStateChangedFunc(&lastClusterInfo, &curClusterInfo) {
			unchangedCounter = 0
			// Patch cluster object with updated timestamp information
			// This timestamp will be used with operation timeout to determine
			// stalled state of the cluster
			// TODO(anuj): consider reducing the number of patches done to the cluster
			// by batching a couple of those that happen close to one another.
			// Even a timestamp patch spaced at least 2-3 minutes apart will not affect
			// the correctness of this algo, but can potentially reduce patches by quite
			// a bit if there are quite a number of changes detected during that span.
			c.patchClusterWithOperationObservedTimestamp(clusterName, namespace)
		} else {
			unchangedCounter++
			log.V(7).Infof("cluster state is unchanged %v", unchangedCounter)
		}

		lastClusterInfo = curClusterInfo

		// if waiting for more than timeout time, return error
		if interval*time.Duration(unchangedCounter) > timeout {
			return true, errors.New("timed out waiting for upgrade to complete")
		}

		return false, err
	}

	return c.poller.PollImmediateInfiniteWithGetter(interval, getterFunc)
}

func (c *client) PatchClusterObject(clusterName, clusterNamespace, patchJSONString string) error {
	err := c.PatchResource(&capi.Cluster{}, clusterName, clusterNamespace, patchJSONString, types.MergePatchType, nil)
	if err != nil {
		return errors.Wrap(err, "unable to patch the cluster object")
	}
	return nil
}

func (c *client) GetClusterStatusInfo(clusterName, namespace string, workloadClusterClient Client) ClusterStatusInfo {
	var err error
	errList := []error{}
	clusterStatusInfo := ClusterStatusInfo{}

	// not all operation will require current k8s version so workloadClusterClient value can be nil.
	if workloadClusterClient != nil {
		clusterStatusInfo.KubernetesVersion, err = workloadClusterClient.GetKubernetesVersion()
		if err != nil {
			errList = append(errList, err)
		}
	}

	clusterStatusInfo.ClusterObject = &capi.Cluster{}
	if err := c.GetResource(clusterStatusInfo.ClusterObject, clusterName, namespace, nil, nil); err != nil {
		errList = append(errList, err)
	}

	if clusterStatusInfo.KCPObject, err = c.GetKCPObjectForCluster(clusterName, namespace); err != nil {
		errList = append(errList, err)
	}

	if clusterStatusInfo.MDObjects, err = c.GetMDObjectForCluster(clusterName, namespace); err != nil {
		errList = append(errList, err)
	}

	if clusterStatusInfo.CPMachineObjects, clusterStatusInfo.WorkerMachineObjects, err = c.GetMachineObjectsForCluster(clusterName, namespace); err != nil {
		errList = append(errList, err)
	}

	clusterStatusInfo.RetrievalError = kerrors.NewAggregate(errList)

	return clusterStatusInfo
}

func (c *client) GetKCPObjectForCluster(clusterName, namespace string) (*controlplanev1.KubeadmControlPlane, error) {
	kcpList := &controlplanev1.KubeadmControlPlaneList{}
	if err := c.GetResourceList(kcpList, clusterName, namespace, nil, nil); err != nil {
		return nil, err
	}
	if len(kcpList.Items) != 1 {
		return nil, errors.Errorf("zero or multiple KCP objects found for the given cluster, %v %v %v", len(kcpList.Items), clusterName, namespace)
	}
	return &kcpList.Items[0], nil
}

func (c *client) GetMDObjectForCluster(clusterName, namespace string) ([]capi.MachineDeployment, error) {
	mdList := &capi.MachineDeploymentList{}
	if err := c.GetResourceList(mdList, clusterName, namespace, nil, nil); err != nil {
		return nil, err
	}
	if len(mdList.Items) == 0 {
		return nil, errors.New("no MachineDeployment objects found for the given cluster")
	}
	return mdList.Items, nil
}

func (c *client) GetMachineObjectsForCluster(clusterName, namespace string) (map[string]capi.Machine, map[string]capi.Machine, error) {
	mdList := &capi.MachineList{}
	if err := c.GetResourceList(mdList, clusterName, namespace, nil, nil); err != nil {
		return nil, nil, err
	}

	cpMachines := make(map[string]capi.Machine)
	workerMachines := make(map[string]capi.Machine)

	for i := range mdList.Items {
		key := mdList.Items[i].Name + "-" + mdList.Items[i].Status.Phase
		if _, labelFound := mdList.Items[i].Labels[capi.MachineControlPlaneLabelName]; labelFound {
			cpMachines[key] = mdList.Items[i]
		} else {
			workerMachines[key] = mdList.Items[i]
		}
	}

	return cpMachines, workerMachines, nil
}

func (c *client) GetKubernetesVersion() (string, error) {
	versionInfo, err := c.discoveryClient.ServerVersion()
	if err != nil {
		return "", err
	}
	return versionInfo.GitVersion, nil
}

func (c *client) PatchClusterObjectWithOptionalMetadata(clusterName, namespace, metadataKey string, metadata map[string]string) (string, error) {
	if len(metadata) == 0 {
		return "", nil
	}
	metadataFormat := "%q : %q,\n"
	var metadataBuilder strings.Builder
	for key, value := range metadata {
		metadataBuilder.WriteString(fmt.Sprintf(metadataFormat, key, value))
	}
	metadataStr := metadataBuilder.String()
	// removes trailing comma, newline
	metadataStr = metadataStr[:len(metadataStr)-2]
	patchFormat := `
	{
		"metadata": {
			%q: {
				%s
			}
		}
	}`
	patchAnnotations := fmt.Sprintf(patchFormat, metadataKey, metadataStr)
	err := c.PatchClusterObject(clusterName, namespace, patchAnnotations)
	if err != nil {
		return "", errors.Wrap(err, "unable to patch the management cluster object with optional metadata")
	}
	return patchAnnotations, nil
}

func (c *client) PatchClusterObjectWithTKGVersion(clusterName, namespace, tkgVersion string) error {
	patchAnnotations := fmt.Sprintf(annotationPatchFormat, TKGVersionKey, tkgVersion)
	err := c.PatchClusterObject(clusterName, namespace, patchAnnotations)
	if err != nil {
		return errors.Wrap(err, "unable to patch the management cluster object with TKG version")
	}
	return nil
}

func (c *client) GetManagementClusterTKGVersion(mgmtClusterName, clusterNamespace string) (string, error) {
	mcObject := &capi.Cluster{}
	err := c.GetResource(mcObject, mgmtClusterName, clusterNamespace, nil, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to get the cluster object")
	}
	version, exists := mcObject.Annotations[TKGVersionKey]
	// if TKGVersionKey does not exists in annotation of the management cluster object
	// assume this is v1.0.0 management cluster as TKGVERSION was not patched for
	// v1.0.x release
	if !exists {
		version = "v1.0.0"
	}

	return version, nil
}

// GetTanzuKubernetesReleases returns the list of available TanzuKubernetesReleases
func (c *client) GetTanzuKubernetesReleases(tkrName string) ([]runv1alpha1.TanzuKubernetesRelease, error) {
	var tkrList runv1alpha1.TanzuKubernetesReleaseList
	getTKRTimeout := 2 * CheckResourceInterval
	_, err := c.poller.PollImmediateWithGetter(CheckResourceInterval, getTKRTimeout, func() (interface{}, error) {
		return nil, c.ListResources(&tkrList)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list current TKr's")
	}
	if tkrName == "" {
		return tkrList.Items, nil
	}

	result := []runv1alpha1.TanzuKubernetesRelease{}
	for i := range tkrList.Items {
		if strings.HasPrefix(tkrList.Items[i].Name, tkrName) {
			result = append(result, tkrList.Items[i])
		}
	}
	return result, nil
}

// GetBomConfigMap gets the BOM ConfigMap
func (c *client) GetBomConfigMap(tkrNameLabel string) (corev1.ConfigMap, error) {
	selectors := []crtclient.ListOption{
		crtclient.InNamespace(tkrconstants.TKRNamespace),
		crtclient.MatchingLabels(map[string]string{tkrconstants.BomConfigMapTKRLabel: tkrNameLabel}),
	}

	cmList := &corev1.ConfigMapList{}
	err := c.clientSet.List(context.Background(), cmList, selectors...)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "failed to list current TKr's")
	}
	if len(cmList.Items) != 1 {
		return corev1.ConfigMap{}, errors.Wrapf(err, "failed to find the BOM ConfigMap matching the label %s: %v", tkrNameLabel, err)
	}

	return cmList.Items[0], nil
}

// GetClusterInfrastructure gets the underlying infrastructure being used
func (c *client) GetClusterInfrastructure() (string, error) {
	clusters := &capi.ClusterList{}

	selectors := []crtclient.ListOption{
		crtclient.MatchingLabels(map[string]string{tkrconstants.ManagememtClusterRoleLabel: ""}),
	}
	err := c.clientSet.List(context.Background(), clusters, selectors...)
	if err != nil || len(clusters.Items) != 1 {
		return "", errors.Wrap(err, "unable to get current management cluster")
	}

	return clusters.Items[0].Spec.InfrastructureRef.Kind, nil
}

// DeactivateTanzuKubernetesReleases deactivates the given TanzuKubernetesReleases
// TKr is deactivated by adding label (inactive: "" ) to the TKr resource
func (c *client) DeactivateTanzuKubernetesReleases(tkrName string) error {
	var tkr runv1alpha1.TanzuKubernetesRelease
	deactivateTKRTimeout := 2 * CheckResourceInterval
	patchFormat := `
	{
		"metadata": {
		    "labels": {
			    %q: ""
		    }
	    }
	}`
	patchStr := fmt.Sprintf(patchFormat, tkrconstants.TanzuKubernetesReleaseInactiveLabel)
	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: deactivateTKRTimeout}
	err := c.PatchResource(&tkr, tkrName, "", patchStr, types.MergePatchType, pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to patch the TKr object with inactive label")
	}

	return nil
}

// ActivateTanzuKubernetesReleases activates the given TanzuKubernetesReleases
// TKr is activated by removing the inactive label (by patching labels with inactive: null) on TKr resource
func (c *client) ActivateTanzuKubernetesReleases(tkrName string) error {
	var tkr runv1alpha1.TanzuKubernetesRelease
	activateTKRTimeout := 2 * CheckResourceInterval
	patchFormat := `
	{
		"metadata": {
		    "labels": {
			    %q: null
		    }
	    }
	}`
	patchStr := fmt.Sprintf(patchFormat, tkrconstants.TanzuKubernetesReleaseInactiveLabel)
	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: activateTKRTimeout}
	err := c.PatchResource(&tkr, tkrName, "", patchStr, types.MergePatchType, pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to patch the TKr object with inactive label")
	}

	return nil
}

// GetResource gets the kubernetes resource passed as reference either directly or with polling mechanism
// resourceReference is a reference to resource struct you want to retrieve
// resourceName is name of the resource to get
// namespace is namespace in which the resource to be searched, if empty current namespace will be used
// postVerify verifies the resource with some state once it is retrieved from kubernetes, pass nil if nothing to verify
// pollOptions use this if you want to continuously poll for object if error occurs, pass nil if don't want polling
// Note: Make sure resource you are retrieving is added into Scheme with init function below
func (c *client) GetResource(resourceReference interface{}, resourceName, namespace string, postVerify PostVerifyrFunc, pollOptions *PollOptions) error {
	var err error
	if namespace == "" {
		if namespace, err = c.GetCurrentNamespace(); err != nil {
			return err
		}
	}

	// get the runtime object from interface
	obj, err := c.getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}

	// if pollOptions are provided use the polling and wait for the result/error/timeout
	// else use normal get
	if pollOptions != nil {
		log.V(4).Infof("Waiting for resource %s of type %s to be up and running", resourceName, reflect.TypeOf(resourceReference))
		_, err = c.poller.PollImmediateWithGetter(pollOptions.Interval, pollOptions.Timeout, func() (interface{}, error) {
			return nil, c.get(resourceName, namespace, obj, postVerify)
		})
		return err
	}

	return c.get(resourceName, namespace, obj, postVerify)
}

// GetResourceList gets the list kubernetes resources passed as reference either directly or with polling mechanism
// resourceReference is a reference to resource struct you want to retrieve
// resourceName is name of the resource to get
// namespace is namespace in which the resource to be searched, if empty current namespace will be used
// postVerify verifies the resource with some state once it is retrieved from kubernetes, pass nil if nothing to verify
// pollOptions use this if you want to continuously poll for object if error occurs, pass nil if don't want polling
// Note: Make sure resource you are retrieving is added into Scheme with init function below
func (c *client) GetResourceList(resourceReference interface{}, clusterName, namespace string, postVerify PostVerifyrFunc, pollOptions *PollOptions) error {
	var err error
	if namespace == "" {
		if namespace, err = c.GetCurrentNamespace(); err != nil {
			return err
		}
	}

	// get the runtime object from interface
	obj, err := c.getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}

	// if pollOptions are provided use the polling and wait for the result/error/timeout
	// else use normal list
	if pollOptions != nil {
		log.V(4).Infof("Waiting for resources type %s to be up and running", reflect.TypeOf(resourceReference))
		_, err = c.poller.PollImmediateWithGetter(pollOptions.Interval, pollOptions.Timeout, func() (interface{}, error) {
			return nil, c.list(clusterName, namespace, obj, postVerify)
		})
		return err
	}

	return c.list(clusterName, namespace, obj, postVerify)
}

func removeAppliedFile(f *os.File) {
	filePath := f.Name()
	if os.Getenv("TKG_KEEP_APPLIED_YAML") != "" {
		log.Infof("Applied yaml saved at %s", filePath)
	} else {
		// In windows environment it is required to close the file first before remove call otherwise
		// it throws error "The process cannot access the file because it is being used by another process"
		_ = f.Close()
		if err := os.Remove(filePath); err != nil {
			log.Infof("Unable to remove file %s, Error: %s", filePath, err.Error())
		}
	}
}

func (c *client) kubectlApplyFile(url string) error {
	args := []string{"apply"}
	if c.kubeConfigPath != "" {
		args = append(args, "--kubeconfig", c.kubeConfigPath)
	}

	if c.currentContext != "" {
		args = append(args, "--context", c.currentContext)
	}

	args = append(args, "-f", url)
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "kubectl apply failed, output: %s", string(out))
	}

	return nil
}

func (c *client) kubectlApply(yaml string) error {
	f, err := os.CreateTemp("", "kubeapply-")
	if err != nil {
		return errors.Wrap(err, "unable to create temp file")
	}
	defer removeAppliedFile(f)
	err = os.WriteFile(f.Name(), []byte(yaml), constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrap(err, "unable to write temp file")
	}
	return c.kubectlApplyFile(f.Name())
}

func (c *client) kubectlExplainResource(resource string) ([]byte, error) {
	args := []string{"explain"}
	if c.kubeConfigPath != "" {
		args = append(args, "--kubeconfig", c.kubeConfigPath)
	}

	if c.currentContext != "" {
		args = append(args, "--context", c.currentContext)
	}

	args = append(args, resource)
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "kubectl explain failed, output: %s", string(out))
	}
	return out, nil
}

func (c *client) kubectlGetResource(resource string, args ...string) ([]byte, error) {
	getargs := []string{"get"}
	if c.kubeConfigPath != "" {
		getargs = append(getargs, "--kubeconfig", c.kubeConfigPath)
	}

	if c.currentContext != "" {
		getargs = append(getargs, "--context", c.currentContext)
	}

	getargs = append(getargs, resource)
	getargs = append(getargs, args...)
	cmd := exec.Command("kubectl", getargs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "kubectl get failed, output: %s", string(out))
	}
	return out, nil
}

func (c *client) GetSecretValue(secretName, key, namespace string, pollOptions *PollOptions) ([]byte, error) {
	var err error

	if pollOptions == nil {
		pollOptions = &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	}

	secret := &corev1.Secret{}
	err = c.GetResource(secret, secretName, namespace, nil, pollOptions)
	if err != nil {
		return nil, err
	}

	data, ok := secret.Data[key]
	if !ok {
		return nil, errors.Errorf("Unable to obtain %s field from secret's data", kubeConfigDataField)
	}

	return data, nil
}

func (c *client) GetKubeConfigForCluster(clusterName, namespace string, pollOptions *PollOptions) ([]byte, error) {
	log.V(4).Info("Getting secret for cluster")
	clusterSecretName := fmt.Sprintf("%s-%s", clusterName, kubeConfigSecretSuffix)
	kubeConfigBytes, err := c.GetSecretValue(clusterSecretName, kubeConfigDataField, namespace, pollOptions)
	if err != nil {
		return nil, err
	}

	if sysrt.GOOS == "darwin" {
		// grab kcp object machine Kind to determine infrastructure provider
		infraProvider, err := c.GetInfrastructureMachineKindForCluster(clusterName, namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting infrastructure provider for cluster with name %s", clusterName)
		}

		// If this does not get applied on macOS and with the docker provider then stalling occurs
		if infraProvider == constants.DockerMachineTemplate {
			// get the docker client with environment options
			ctx := context.Background()
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
			if err != nil {
				return nil, errors.Wrapf(err, "Error getting docker client when fixing kubeconfig")
			}
			return utils.FixKubeConfigForMacEnvironment(ctx, cli, kubeConfigBytes)
		}
	}

	return kubeConfigBytes, nil
}

// GetInfrastructureMachineKindForCluster gets the kcp object infrastructure
// template's Kind using the cluster name and namespace
// TODO(dcline) write tests for this
func (c *client) GetInfrastructureMachineKindForCluster(clusterName, namespace string) (string, error) {
	kcpObject, err := c.GetKCPObjectForCluster(clusterName, namespace)
	if err != nil {
		return "", err
	}

	return kcpObject.Spec.InfrastructureTemplate.Kind, nil
}

// GetCurrentNamespace returns the namespace from the current context in the kubeconfig file
func (c *client) GetCurrentNamespace() (string, error) {
	kubeconfig := cluster.Kubeconfig{Path: c.kubeConfigPath, Context: c.currentContext}
	clusterclient := cluster.New(kubeconfig, nil)
	return clusterclient.Proxy().CurrentNamespace()
}

func (c *client) CreateNamespace(name string) error {
	namespace := &corev1.Namespace{}
	namespace.Name = name
	err := c.clientSet.Create(ctx, namespace)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *client) GetCurrentClusterName(ctx string) (string, error) {
	config, err := clientcmd.LoadFromFile(c.kubeConfigPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load kubeconfig file from %q", c.kubeConfigPath)
	}

	if ctx == "" {
		ctx = config.CurrentContext
	}

	for contextName, ctxObj := range config.Contexts {
		if contextName == ctx {
			return ctxObj.Cluster, nil
		}
	}
	return "", errors.Errorf("unable to find cluster name from kubeconfig file: %q", c.kubeConfigPath)
}

func (c *client) GetCurrentKubeContext() (string, error) {
	config, err := clientcmd.LoadFromFile(c.kubeConfigPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load kubeconfig file from %q", c.kubeConfigPath)
	}

	return config.CurrentContext, nil
}

// UseContext updates current-context in the kubeconfig file
// also updates the clientset of clusterclient to point to new cluster
func (c *client) UseContext(contextName string) error {
	config, err := clientcmd.LoadFromFile(c.kubeConfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to load kubeconfig file from %q", c.kubeConfigPath)
	}
	if _, ok := config.Contexts[contextName]; !ok {
		return errors.Errorf("context is not defined for %q in file %q", contextName, c.kubeConfigPath)
	}
	config.CurrentContext = contextName
	err = clientcmd.WriteToFile(*config, c.kubeConfigPath)
	if err != nil {
		return err
	}
	// update the current clientset to point to new cluster
	err = c.updateK8sClients(contextName)
	if err != nil {
		return err
	}
	return nil
}

// ExportCurrentKubeconfigToFile saves the current kubeconfig to temporary file and returns the file
func (c *client) ExportCurrentKubeconfigToFile() (string, error) {
	kubeConfigBytes, err := c.LoadCurrentKubeconfigBytes()
	if err != nil {
		return "", errors.Wrap(err, "unable to load current kubeconfig bytes")
	}
	filename, err := utils.CreateTempFile("", "tmp_kubeconfig")
	if err != nil {
		return "", errors.Wrap(err, "unable to save kubeconfig to temporary file")
	}
	err = utils.WriteToFile(filename, kubeConfigBytes)
	if err != nil {
		return "", errors.Wrap(err, "unable to write kubeconfig to temporary file")
	}
	return filename, nil
}

// GetCurrentKubeconfigFile returns currently used kubeconfig file path based on default loading rules
func (c *client) GetCurrentKubeconfigFile() string {
	return c.kubeConfigPath
}

// IsRegionalCluster() checks if the current kube context point to a management cluster
func (c *client) IsRegionalCluster() error {
	var providers clusterctlv1.ProviderList

	err := c.ListResources(&providers, &crtclient.ListOptions{})
	if err != nil {
		return err
	}

	for _, t := range providerTypes {
		found := false
		for i := range providers.Items {
			if clusterctlv1.ProviderType(providers.Items[i].Type) == t {
				found = true
				break
			}
		}

		if !found {
			return errors.Errorf("not a valid management cluster, missing provider: %s", string(t))
		}
	}

	return nil
}

func (c *client) GetRegionalClusterDefaultProviderName(providerType clusterctlv1.ProviderType) (string, error) {
	var providers clusterctlv1.ProviderList
	err := c.ListResources(&providers, &crtclient.ListOptions{})
	if err != nil {
		return "", err
	}
	names := sets.NewString()
	for i := range providers.Items {
		if clusterctlv1.ProviderType(providers.Items[i].Type) == providerType {
			names.Insert(providers.Items[i].ProviderName)
		}
	}
	// If there is only one provider, this is the default
	if names.Len() == 1 {
		return names.List()[0], nil
	}
	// There is no provider or more than one provider of this type; in both cases, a default provider name cannot be decided.
	return "", errors.New("unable to find the default provider,since there are more than 1 providers")
}

func (c *client) ListClusters(namespace string) ([]capi.Cluster, error) {
	var clusters capi.ClusterList

	err := c.ListResources(&clusters, &crtclient.ListOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}
	return clusters.Items, nil
}

func (c *client) DeleteCluster(clusterName, namespace string) error {
	isPacific, err := c.IsPacificRegionalCluster()
	if err == nil && isPacific {
		tkcObj, err := c.getPacificClusterObject(clusterName, namespace, constants.DefaultPacificClusterAPIVersion)
		if err != nil {
			errString := fmt.Sprintf("failed to get cluster object for delete: %s", err.Error())
			return errors.New(errString)
		}
		return c.DeleteResource(tkcObj)
	}

	clusterObject := &capi.Cluster{}
	clusterObject.Name = clusterName
	clusterObject.Namespace = namespace

	return c.DeleteResource(clusterObject)
}

func (c *client) UpdateReplicas(resourceReference interface{}, resourceName, resourceNameSpace string, replicaCount int32) error {
	patchReplicaCount := fmt.Sprintf("{\"spec\":{\"replicas\": %v}}", replicaCount)
	err := c.PatchResource(resourceReference, resourceName, resourceNameSpace, patchReplicaCount, types.MergePatchType, nil)
	if err != nil {
		return errors.Wrap(err, "unable to patch the replica count")
	}
	return nil
}

func (c *client) GetPacificTKCAPIVersion() (string, error) {
	yaml, err := c.poller.PollImmediateWithGetter(kubectlApplyRetryInterval, kubectlApplyRetryTimeout, func() (interface{}, error) {
		return c.kubectlExplainResource("tkc")
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to get kubectl explain response for tkc resource")
	}

	yamlBytes := yaml.([]byte)
	re := regexp.MustCompile(`VERSION:(.*)\n`)
	match := re.FindStringSubmatch(string(yamlBytes))
	if len(match) == 0 || strings.TrimSpace(match[1]) == "" {
		return "", errors.Wrap(err, "failed to get TKC API version")
	}
	return strings.TrimSpace(match[1]), nil
}

func (c *client) IsPacificRegionalCluster() (bool, error) {
	return c.isTKCCrdAvailableInTanzuRunAPIGroup()
}

func (c *client) isTKCCrdAvailableInTanzuRunAPIGroup() (bool, error) {
	// for pacific we should be able to fetch the api group "run.tanzu.vmware.com"
	data, err := c.discoveryClient.RESTClient().Get().AbsPath(constants.TanzuRunAPIGroupPath).Do().Raw()
	if err != nil {
		//  If the url is not available return false
		if apierrors.IsNotFound(err) {
			return false, nil
		}
	}
	// fetch the groupVersion from the preferred version in api group
	var groupdata interface{}
	err = json.Unmarshal(data, &groupdata)
	if err != nil {
		return false, err
	}
	groupversion, err := jsonpath.Read(groupdata, "$.preferredVersion.groupVersion")
	if err != nil {
		return false, errors.Wrap(err, "failed to read group version")
	}

	groupVersionURL := fmt.Sprintf("/apis/%s", groupversion.(string))
	data, err = c.discoveryClient.RESTClient().Get().AbsPath(groupVersionURL).Do().Raw()
	if err != nil {
		//  If the url is not available return false
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to get resources in %s", groupVersionURL)
	}

	var resourceslist interface{}
	err = json.Unmarshal(data, &resourceslist)
	if err != nil {
		return false, errors.Wrap(err, "unable to unmarshall the resources list")
	}
	// get all resource kinds
	kinds, err := jsonpath.Read(resourceslist, "$.resources[*].kind")
	if err != nil {
		return false, nil
	}
	// check for resource of kind "TanzuKubernetesCluster"
	kindlist := kinds.([]interface{})
	for _, kind := range kindlist {
		if kind.(string) == "TanzuKubernetesCluster" {
			return true, nil
		}
	}
	return false, nil
}

func (c *client) PatchK8SVersionToPacificCluster(clusterName, namespace, apiVersion, kubernetesVersion string) error {
	tkcObj, err := c.getPacificClusterObject(clusterName, namespace, apiVersion)
	if err != nil {
		return errors.Wrap(err, "failed to patch kubernetes version")
	}
	patchString := `{
		"spec" : {
		  "distribution" : {
			  "fullVersion" : "",
			  "version" : "%s"
		  }
		}
	  }`

	patchKubernetesVersion := fmt.Sprintf(patchString, kubernetesVersion)
	log.V(3).Infof("Applying TanzuKubernetesCluster kubernetes version update patch: %s", patchKubernetesVersion)
	err = c.PatchResource(tkcObj, clusterName, namespace, patchKubernetesVersion, types.MergePatchType, nil)
	if err != nil {
		return errors.Wrap(err, "unable to patch the k8s version for tkc object")
	}
	return nil
}

func (c *client) ScalePacificClusterControlPlane(clusterName, namespace, apiVersion string, controlPlaneCount int32) error {
	tkcObj, err := c.getPacificClusterObject(clusterName, namespace, apiVersion)
	if err != nil {
		return err
	}
	patchPacificControlPlaneCount := fmt.Sprintf("{\"spec\":{ \"topology\":{\"controlPlane\":{\"count\": %v}}}}", controlPlaneCount)
	err = c.PatchResource(tkcObj, clusterName, namespace, patchPacificControlPlaneCount, types.MergePatchType, nil)
	if err != nil {
		return errors.Wrap(err, "unable to patch the cluster controlPlane count")
	}
	return nil
}

func (c *client) ScalePacificClusterWorkerNodes(clusterName, namespace, apiVersion string, workersCount int32) error {
	tkcObj, err := c.getPacificClusterObject(clusterName, namespace, apiVersion)
	if err != nil {
		return err
	}
	patchPacificWorkersCount := fmt.Sprintf("{\"spec\":{ \"topology\":{\"workers\":{\"count\": %v}}}}", workersCount)
	err = c.PatchResource(tkcObj, clusterName, namespace, patchPacificWorkersCount, types.MergePatchType, nil)
	if err != nil {
		return errors.Wrap(err, "unable to patch the cluster workers count")
	}
	return nil
}

func (c *client) WaitForPacificCluster(clusterName, namespace, apiVersion string) error {
	_, err := c.poller.PollImmediateWithGetter(CheckClusterInterval, c.operationTimeout, func() (interface{}, error) {
		tkcObj, err := c.getPacificClusterObject(clusterName, namespace, apiVersion)
		if err != nil {
			return false, err
		}
		jsonoutput, _ := tkcObj.MarshalJSON()
		var cluster interface{}
		_ = json.Unmarshal(jsonoutput, &cluster)
		clusterstatus, _ := jsonpath.Read(cluster, "$.status.phase")
		if clusterstatus != statusRunning {
			return false, errors.New("cluster is still not provisioned, retrying")
		}
		return true, nil
	})
	return err
}

func (c *client) getPacificClusterObject(clusterName, namespace, apiVersion string) (*unstructured.Unstructured, error) {
	// if version is not supplied, get the pacific TKC api version
	var err error
	if apiVersion == "" {
		apiVersion, err = c.GetPacificTKCAPIVersion()
		if err != nil {
			return nil, errors.Wrap(err, "unable to get TKC API Version")
		}
	}
	objKey := crtclient.ObjectKey{Name: clusterName, Namespace: namespace}
	tkcObj := &unstructured.Unstructured{}
	tkcObj.SetKind(constants.PacificClusterKind)
	tkcObj.SetAPIVersion(apiVersion)
	ns := namespace
	if namespace == "" {
		ns = constants.DefaultNamespace
	}
	if err := c.clientSet.Get(ctx, objKey, tkcObj); err != nil {
		return nil, errors.Wrapf(err, "failed to get cluster object in namespace: '%s'", ns)
	}

	return tkcObj, nil
}

// ListPacificClusterObjects returns list of TanzuKubernetesCluster as interface
func (c *client) ListPacificClusterObjects(apiVersion string, listOptions *crtclient.ListOptions) ([]interface{}, error) {
	// if version is not supplied, get the pacific TKC api version
	var err error
	if apiVersion == "" {
		apiVersion, err = c.GetPacificTKCAPIVersion()
		if err != nil {
			return nil, errors.Wrap(err, "unable to get TKC API Version")
		}
	}

	// Create TanzuKubernetesClusterList object
	tkcObjList := &unstructured.UnstructuredList{}
	tkcObjList.SetKind(constants.PacificClusterListKind)
	tkcObjList.SetAPIVersion(apiVersion)
	if err := c.clientSet.List(ctx, tkcObjList, listOptions); err != nil {
		return nil, errors.Wrap(err, "failed to list TanzuKubernetesCluster objects")
	}

	jsonoutput, _ := tkcObjList.MarshalJSON()
	var clusterList interface{}
	err = json.Unmarshal(jsonoutput, &clusterList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json unmarshal TanzuKubernetesClusterList object")
	}
	items, err := jsonpath.Read(clusterList, "$.items")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read TanzuKubernetesClusterList with json path")
	}
	clusters := items.([]interface{})
	return clusters, nil
}

func (c *client) verifyPacificK8sVersionUpdate(clusterName, namespace, newK8sVersion string) error {
	workerMachines, err := c.getWorkerMachineObjectsForPacificCluster(clusterName, namespace)
	if err != nil {
		return errors.Wrap(err, "failed to get worker machine objects to verify k8s version update")
	}

	unupgradedMachineList := []string{}
	for i := range workerMachines {
		if workerMachines[i].Spec.Version == nil || !strings.HasPrefix(strings.TrimPrefix(newK8sVersion, "v"), strings.TrimPrefix(*workerMachines[i].Spec.Version, "v")) {
			log.V(9).Infof("worker machine '%s' is still not upgraded, current kubernetes version:'%s' but expecting :'%s'",
				workerMachines[i].Name, strings.TrimPrefix(*workerMachines[i].Spec.Version, "v"), strings.TrimPrefix(newK8sVersion, "v"))
			unupgradedMachineList = append(unupgradedMachineList, workerMachines[i].Name)
		}
	}

	if len(unupgradedMachineList) > 0 {
		return errors.Errorf("worker machines %v are still not upgraded", unupgradedMachineList)
	}
	return nil
}

func (c *client) getWorkerMachineObjectsForPacificCluster(clusterName, namespace string) ([]capiv1alpha2.Machine, error) {
	mdList := &capiv1alpha2.MachineList{}
	if err := c.GetResourceList(mdList, clusterName, namespace, nil, nil); err != nil {
		return nil, err
	}

	workerMachines := []capiv1alpha2.Machine{}
	for i := range mdList.Items {
		if _, labelFound := mdList.Items[i].Labels[capi.MachineControlPlaneLabelName]; !labelFound {
			workerMachines = append(workerMachines, mdList.Items[i])
		}
	}
	return workerMachines, nil
}

// WaitForPacificClusterK8sVersionUpdate waits for Pacific cluster K8s version upgrade to complete.
// Unlike other CAPA/CAPV providers, Pacific has an update job to update the k8s version and it would update status as follows
//  -sets TKCObject.status.phase to 'updateFailed' if the update failed
//  -sets TKCObject.status.phase to 'updating' if the update is in progress
//  -sets TKCObject.status.phase to 'running' if the update is complete , however it can potentially return to
//        running between controlplane and workernode update.So worker nodes k8s version is verified to determine the update is indeed complete
func (c *client) WaitForPacificClusterK8sVersionUpdate(clusterName, namespace, apiversion, newK8sVersion string) error {
	var err error
	counter := 0
	interval := 15 * time.Second
	errcount := 0

	getterFunc := func() (interface{}, error) {
		var tkcObj *unstructured.Unstructured
		tkcObj, err = c.getPacificClusterObject(clusterName, namespace, apiversion)
		if err != nil {
			// if control-plane API server couldn't respond to the get TKC object requests for more than 2 minutes continuously,
			// break from poll with error instead of waiting for long time period
			if interval*time.Duration(errcount) > 2*time.Minute {
				return true, err
			}
			errcount++
			return false, err
		}

		errcount = 0
		jsonoutput, _ := tkcObj.MarshalJSON()
		var cluster interface{}
		_ = json.Unmarshal(jsonoutput, &cluster)
		clusterstatus, _ := jsonpath.Read(cluster, "$.status.phase")
		if clusterstatus == "updateFailed" || clusterstatus == "deleting" {
			return true, errors.New("cluster kubernetes version update failed")
		}

		if clusterstatus == "updating" {
			err = errors.New("cluster kubernetes version is still being upgraded")
		} else if clusterstatus == statusRunning {
			// check if the version is updated on worker nodes, if yes return
			err = c.verifyPacificK8sVersionUpdate(clusterName, namespace, newK8sVersion)
			if err == nil {
				return false, nil
			}
		} else {
			// any other status is still consider as upgrade still in progress
			err = errors.New("cluster kubernetes version is still being upgraded")
		}
		// still updating, keep waiting
		counter++
		// if wait time is more than operationTimeout, return error
		if interval*time.Duration(counter) > c.operationTimeout {
			return true, errors.New("timed out waiting for upgrade to complete. Upgrade is still in progress, you can check the status using `tanzu cluster get` command")
		}

		return false, err
	}
	return c.poller.PollImmediateInfiniteWithGetter(interval, getterFunc)
}

func (c *client) GetPacificTanzuKubernetesReleases() ([]string, error) {
	result, err := c.poller.PollImmediateWithGetter(kubectlApplyRetryInterval, kubectlApplyRetryTimeout, func() (interface{}, error) {
		return c.kubectlGetResource("tanzukubernetesrelease", "-o=custom-columns=VERSION:.spec.version", "--no-headers")
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get 'tanzukubernetesrelease' from vSphere with Kubernetes")
	}
	outbytes := result.([]byte)
	versions := strings.Fields(string(outbytes))
	return versions, nil
}

// PatchCalicoNodeDaemonSetWithNewNodeSelector patches calico daemonset with new nodeSelector
func (c *client) PatchCalicoNodeDaemonSetWithNewNodeSelector(selectorKey, selectorValue string) error { // nolint:dupl
	ds := &appsv1.DaemonSet{}
	if err := c.GetResource(ds, calicoNodeKey, metav1.NamespaceSystem, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// if ds is missing, return without errors
			return nil
		}
		return errors.Wrapf(err, "failed to look up '%s' daemonset", calicoNodeKey)
	}

	helper, err := patch.NewHelper(ds, c.clientSet)
	if err != nil {
		return err
	}

	// If selector already present, skip patch operation
	for key, value := range ds.Spec.Template.Spec.NodeSelector {
		if key == selectorKey && value == selectorValue {
			log.V(3).Infof("nodeSelector '%v: %v' already present for '%v' daemonset, skipping patch operation", selectorKey, selectorValue, calicoNodeKey)
			return nil
		}
	}

	log.V(3).Infof("Patching '%s' daemonset...", calicoNodeKey)
	nodeSelectorMap := map[string]string{}
	nodeSelectorMap[selectorKey] = selectorValue
	ds.Spec.Template.Spec.NodeSelector = nodeSelectorMap
	err = helper.Patch(ctx, ds)
	if err != nil {
		return errors.Wrapf(err, "unable to update the '%s' daemonset", calicoNodeKey)
	}
	return nil
}

// PatchCalicoKubeControllerDeploymentWithNewNodeSelector patches calico-kube-controller deployment with new nodeSelector
func (c *client) PatchCalicoKubeControllerDeploymentWithNewNodeSelector(selectorKey, selectorValue string) error { // nolint:dupl
	deployment := &appsv1.Deployment{}
	if err := c.GetResource(deployment, calicoKubeControllerKey, metav1.NamespaceSystem, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// if deployment is missing, return without errors
			return nil
		}
		return errors.Wrapf(err, "failed to look up '%s' deployment", calicoKubeControllerKey)
	}

	helper, err := patch.NewHelper(deployment, c.clientSet)
	if err != nil {
		return err
	}

	// If selector already present, skip patch operation
	for key, value := range deployment.Spec.Template.Spec.NodeSelector {
		if key == selectorKey && value == selectorValue {
			log.V(3).Infof("nodeSelector '%v: %v' already present for '%v' deployment, skipping patch operation", selectorKey, selectorValue, calicoKubeControllerKey)
			return nil
		}
	}

	log.V(3).Infof("Patching '%s' deployment...", calicoKubeControllerKey)
	nodeSelectorMap := map[string]string{}
	nodeSelectorMap[selectorKey] = selectorValue
	deployment.Spec.Template.Spec.NodeSelector = nodeSelectorMap
	err = helper.Patch(ctx, deployment)
	if err != nil {
		return errors.Wrapf(err, "unable to update the '%s' deployment", calicoKubeControllerKey)
	}
	return nil
}

// PatchImageRepositoryInKubeProxyDaemonSet updates kubeproxy daemonset with new/custom image repository
func (c *client) PatchImageRepositoryInKubeProxyDaemonSet(newImageRepository string) error {
	ds := &appsv1.DaemonSet{}
	if err := c.GetResource(ds, kubeProxyKey, metav1.NamespaceSystem, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// if kube-proxy is missing, return without errors
			return nil
		}
		return errors.Wrapf(err, "failed to look up '%s' daemonset", kubeProxyKey)
	}

	findKubeProxyContainer := func(ds *appsv1.DaemonSet) *corev1.Container {
		containers := ds.Spec.Template.Spec.Containers
		for idx := range containers {
			if containers[idx].Name == kubeProxyKey {
				return &containers[idx]
			}
		}
		return nil
	}
	patchKubeProxyImage := func(ds *appsv1.DaemonSet, image string) {
		containers := ds.Spec.Template.Spec.Containers
		for idx := range containers {
			if containers[idx].Name == kubeProxyKey {
				containers[idx].Image = image
			}
		}
	}

	container := findKubeProxyContainer(ds)
	if container == nil {
		return nil
	}

	newImageName, err := containerutil.ModifyImageRepository(container.Image, newImageRepository)
	if err != nil {
		return err
	}

	// TODO: current image accessibility verification is based on machine where TKG CLI is used.
	// It would be better to perform the image accessibility check in the target infra instead.
	err = docker.VerifyImageIsAccessible(newImageName)
	if err != nil {
		log.Warningf("Warning: Image accessibility verification failed. Image %s is not reachable from current machine. Please make sure the image is pullable from the Kubernetes node for upgrade to complete successfully", newImageName)
	}

	if container.Image != newImageName {
		helper, err := patch.NewHelper(ds, c.clientSet)
		if err != nil {
			return err
		}
		patchKubeProxyImage(ds, newImageName)
		err = helper.Patch(ctx, ds)
		if err != nil {
			return errors.Wrap(err, "unable to update the kube-proxy daemonset")
		}
	}
	return nil
}

// PatchClusterAPIAWSControllersToUseEC2Credentials ensures that the Cluster API Provider AWS
// controller is pinned to control plane nodes and is running without static credentials such
// that Cluster API AWS runs using the EC2 instance profile attached to the control plane node.
// This is done by zeroing out the credentials secret for CAPA, causing the AWS SDK to fall back
// to the default credential provider chain. We additionally patch the deployment to ensure
// the controller has node affinity to only run on the control plane nodes.
// This should NOT be used when running Cluster API Provider AWS on managed control planes, e.g. EKS
func (c *client) PatchClusterAPIAWSControllersToUseEC2Credentials() error {
	ns := &corev1.Namespace{}
	if err := c.GetResource(ns, CAPAControllerNamespace, CAPAControllerNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// no capa-system namespace, return without errors
			return nil
		}
		return err
	}

	log.V(6).Info("Kubernetes Cluster API Provider AWS detected, attempting to zero out credentials and pivot to EC2 instance profile")

	creds := &corev1.Secret{}
	if err := c.GetResource(creds, CAPACredentialsSecretName, CAPAControllerNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// Warn if secret isn't found
			log.V(4).Warningf("Could not find Kubernetes Cluster API Provider AWS credentials secret: %s/%s ", CAPAControllerNamespace, CAPACredentialsSecretName)
			return nil
		}
		return err
	}

	if creds.StringData == nil {
		creds.StringData = map[string]string{}
	}
	creds.StringData["credentials"] = "\n"

	if err := c.clientSet.Update(ctx, creds); err != nil {
		return errors.Wrap(err, "unable to update the Cluster API Provider AWS credentials secret")
	}

	deployment := &appsv1.Deployment{}
	if err := c.GetResource(deployment, CAPAControllerDeploymentName, CAPAControllerNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// Warn, but do not block if controller deployment is not found
			log.V(4).Warningf("Could not find Kubernetes Cluster API Provider AWS controller deployment: %s/%s ", CAPAControllerNamespace, CAPAControllerDeploymentName)
			return nil
		}
		return err
	}

	helper, err := patch.NewHelper(deployment, c.clientSet)
	if err != nil {
		return err
	}
	capaPodSpec := &deployment.Spec.Template.Spec
	ensurePodSpecControlPlaneAffinity(capaPodSpec)
	if err := helper.Patch(ctx, deployment); err != nil {
		return errors.Wrap(err, "unable to update the Cluster API Provider AWS deployment")
	}

	return nil
}

// PatchCoreDNSImageRepositoryInKubeadmConfigMap updates kubeadm-config configMap with new/custom image repository
func (c *client) PatchCoreDNSImageRepositoryInKubeadmConfigMap(newImageRepository string) error {
	if newImageRepository == "" {
		return nil
	}

	kubedmconfigmap := &corev1.ConfigMap{}
	if err := c.GetResource(kubedmconfigmap, kubeadmConfigKey, metav1.NamespaceSystem, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// if kubeadm-config ConfigMap is missing, return without errors
			return nil
		}
		return errors.Wrapf(err, "failed to determine if %s ConfigMap already exists", kubeadmConfigKey)
	}

	if err := UpdateCoreDNSImageRepositoryInKubeadmConfigMap(kubedmconfigmap, newImageRepository); err != nil {
		return err
	}

	if err := c.clientSet.Update(ctx, kubedmconfigmap); err != nil {
		return errors.Wrap(err, "error updating kubeadm ConfigMap")
	}
	return nil
}

// GetPinnipedIssuerURLAndCA fetches Pinniped supervisor IssuerURL and IssuerCA data from management cluster
func (c *client) GetPinnipedIssuerURLAndCA() (string, string, error) {
	configMap, err := c.getPinnipedInfoConfigMap()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get pinniped-info ConfigMap")
	}
	issuerURL, ok := configMap.Data["issuer"]
	if !ok {
		return "", "", errors.New("failed to read issuer value from the pinniped-info ConfigMap")
	}
	log.V(9).Infof("Pinniped issuer URL fetched from ConfigMap is : %s ", issuerURL)

	issuerCA, ok := configMap.Data["issuer_ca_bundle_data"]
	if !ok || issuerCA == "" {
		return "", "", errors.New("failed to get pinniped issuer CA data")
	}

	return issuerURL, issuerCA, nil
}

// getPinnipedInfoConfigMap return the pinniped-info ConfigMap
func (c *client) getPinnipedInfoConfigMap() (corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: 5 * CheckResourceInterval}
	err := c.GetResource(&configMap, "pinniped-info", "kube-public", nil, pollOptions)
	if err != nil {
		return corev1.ConfigMap{}, err
	}

	return configMap, nil
}

// DeleteExistingKappController deletes kapp-controller that already exists in the cluster.
func (c *client) DeleteExistingKappController() error {
	if err := c.GetResource(&appsv1.Deployment{}, kappControllerKey, kappControllerNamespace, nil, nil); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to determine if %s deployment already exists in %s namespace", kappControllerKey, kappControllerNamespace)
		}
	} else {
		// If kapp-controller deployment is found in tkg-system namespace, then dont delete anything and return.
		return nil
	}

	deployment := &appsv1.Deployment{}
	if err := c.GetResource(deployment, kappControllerKey, kappControllerOldNamespace, nil, nil); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to determine if %s deployment already exists in %s namespace", kappControllerKey, kappControllerOldNamespace)
		}
	} else {
		if err := c.DeleteResource(deployment); err != nil {
			return err
		}
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := c.GetResource(clusterRoleBinding, kappControllerClusterRoleBinding, "", nil, nil); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to determine if %s cluster role binding already exists", kappControllerClusterRoleBinding)
		}
	} else {
		if err := c.DeleteResource(clusterRoleBinding); err != nil {
			return err
		}
	}

	clusterRole := &rbacv1.ClusterRole{}
	if err := c.GetResource(clusterRole, kappControllerClusterRole, "", nil, nil); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to determine if %s cluster role already exists", kappControllerClusterRole)
		}
	} else {
		if err := c.DeleteResource(clusterRole); err != nil {
			return err
		}
	}

	serviceAccount := &corev1.ServiceAccount{}
	if err := c.GetResource(serviceAccount, kappControllServiceAccount, kappControllerOldNamespace, nil, nil); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to determine if %s service account already exists in %s namesapce", kappControllServiceAccount, kappControllerOldNamespace)
		}
	} else {
		if err := c.DeleteResource(serviceAccount); err != nil {
			return err
		}
	}

	return nil
}

// UpdateAWSCNIIngressRules updates the cniIngressRules field for AWSCluster to allow for
// kapp-controller host port that was added in newer versions.
func (c *client) UpdateAWSCNIIngressRules(clusterName, clusterNamespace string) error {
	awsCluster := &capav1alpha3.AWSCluster{}
	if err := c.GetResource(awsCluster, clusterName, clusterNamespace, nil, nil); err != nil {
		return err
	}

	if awsCluster.Spec.NetworkSpec.CNI == nil {
		awsCluster.Spec.NetworkSpec.CNI = &capav1alpha3.CNISpec{}
	}

	if awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules == nil {
		awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules = capav1alpha3.CNIIngressRules{}
	}

	cniIngressRules := awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules
	// first check if existing ingress rules contains the kapp-controller port
	for _, ingressRule := range cniIngressRules {
		if ingressRule.Description != "kapp-controller" {
			continue
		}

		if ingressRule.Protocol != capav1alpha3.SecurityGroupProtocolTCP {
			continue
		}

		if ingressRule.FromPort != DefaultKappControllerHostPort || ingressRule.ToPort != DefaultKappControllerHostPort {
			continue
		}

		return nil
	}

	cniIngressRules = append(cniIngressRules, &capav1alpha3.CNIIngressRule{
		Description: "kapp-controller",
		Protocol:    capav1alpha3.SecurityGroupProtocolTCP,
		FromPort:    DefaultKappControllerHostPort,
		ToPort:      DefaultKappControllerHostPort,
	})

	awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules = cniIngressRules
	if err := c.UpdateResource(awsCluster, clusterName, clusterNamespace); err != nil {
		return err
	}

	return nil
}

// RemoveCEIPTelemetryJob removes installed telemetry job
func (c *client) RemoveCEIPTelemetryJob(clusterName string) error {
	hasCeip, err := c.HasCEIPTelemetryJob(clusterName)
	if err != nil {
		return errors.Wrap(err, "failed to find telemetry cronjob")
	}
	if !hasCeip {
		// Don't attempt to delete cronjob if it doesn't exist
		return nil
	}
	jobResource := &betav1.CronJob{}
	jobResource.Namespace = constants.CeipNamespace
	jobResource.Name = constants.CeipJobName
	err = c.DeleteResource(jobResource)
	if err != nil {
		return errors.Wrap(err, "failed to delete telemetry cronjob")
	}
	return nil
}

func (c *client) AddCEIPTelemetryJob(clusterName, providerName string, bomConfig *tkgconfigbom.BOMConfiguration, isProd, labels, httpProxy, httpsProxy, noProxy string) error {
	var telemetryPath string
	log.V(5).Infof("IsProd: %s", isProd)
	if buildinfo.IsOfficialBuild == "True" {
		telemetryPath = prodTelemetryPath
	} else {
		telemetryPath = stageTelemetryPath
	}

	if isProd != "" {
		isProdVal, _ := strconv.ParseBool(isProd)
		if isProdVal {
			telemetryPath = prodTelemetryPath
		} else {
			telemetryPath = stageTelemetryPath
		}
	}
	log.V(5).Infof("IsOfficialBuild: %s", buildinfo.IsOfficialBuild)

	fullImagePath := tkgconfigbom.GetFullImagePath(bomConfig.Components["tkg_telemetry"][0].Images[telemetryBomImagesMapKey], bomConfig.ImageConfig.ImageRepository)
	imageRepository := fullImagePath + ":" + bomConfig.Components["tkg_telemetry"][0].Images[telemetryBomImagesMapKey].Tag

	telemetryConfigFilePath := embeddedTelemetryConfigYamlPrefix + providerName + ".yaml"
	telemetryConfigYaml, err := telemetrymanifests.Asset(telemetryConfigFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to parse telemetry spec yaml")
	}
	log.V(5).Infof(string(telemetryConfigYaml), clusterName, imageRepository, telemetryPath, labels, httpProxy, httpsProxy, noProxy)
	err = c.kubectlApply(fmt.Sprintf(string(telemetryConfigYaml), clusterName, imageRepository, telemetryPath, labels, httpProxy, httpsProxy, noProxy))
	if err != nil {
		return errors.Wrap(err, "failed to apply telemetry spec")
	}
	return nil
}

// HasCEIPTelemetryJob check whether CEIP telemetry job is present or not
func (c *client) HasCEIPTelemetryJob(clusterName string) (bool, error) {
	cronJobs := &betav1.CronJobList{}
	err := c.GetResourceList(cronJobs, clusterName, constants.CeipNamespace, nil, nil)
	if err != nil {
		return false, errors.Wrap(err, "failed to find telemetry cronjob")
	}
	if cronJobs == nil {
		return false, nil
	}
	return len(cronJobs.Items) > 0, nil
}

// IsClusterRegisteredToTMC() returns true if cluster is registered to Tanzu Mission Control
func (c *client) IsClusterRegisteredToTMC() (bool, error) {
	restconfigClient, err := c.GetRestConfigClient()
	if err != nil {
		return false, err
	}
	clusterQueryClient, err := capdiscovery.NewClusterQueryClientForConfig(restconfigClient)
	if err != nil {
		return false, err
	}

	// Check if 'cluster-agent' resource of type 'agents.clusters.tmc.cloud.vmware.com/v1alpha1' present
	// in 'vmware-system-tmc' namespace. If present, we can say the cluster is registered to TMC
	agent := &corev1.ObjectReference{
		Kind:       "Agent",
		Name:       "cluster-agent",
		Namespace:  constants.TmcNamespace,
		APIVersion: "clusters.tmc.cloud.vmware.com/v1alpha1",
	}
	var testObject = capdiscovery.Object("tmcClusterAgentObj", agent)

	// Build query client.
	cqc := clusterQueryClient.Query(testObject)

	// Execute returns combined result of all queries.
	return cqc.Execute() // return (found, err) response
}

// Options provides way to customize creation of clusterClient
type Options struct {
	poller                    Poller
	crtClientFactory          CrtClientFactory
	discoveryClientFactory    DiscoveryClientFactory
	GetClientInterval         time.Duration
	GetClientTimeout          time.Duration
	OperationTimeout          time.Duration
	verificationClientFactory *VerificationClientFactory
}

// NewOptions returns new options
func NewOptions(poller Poller, crtClientFactory CrtClientFactory, discoveryClientFactory DiscoveryClientFactory, verificationClientFactory *VerificationClientFactory) Options {
	return Options{
		poller:                    poller,
		crtClientFactory:          crtClientFactory,
		discoveryClientFactory:    discoveryClientFactory,
		verificationClientFactory: verificationClientFactory,
	}
}

// NewClient creates new clusterclient from kubeconfig file and poller
// if kubeconfig path is empty it gets default path
// if options.poller is nil it creates default poller. You should only pass custom poller for unit testing
// if options.crtClientFactory is nil it creates default CrtClientFactory
func NewClient(kubeConfigPath string, context string, options Options) (Client, error) { // nolint:gocritic
	var err error
	var rules *clientcmd.ClientConfigLoadingRules
	if kubeConfigPath == "" {
		rules = clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfigPath = rules.GetDefaultFilename()
	}
	if options.poller == nil {
		options.poller = NewPoller()
	}
	if options.crtClientFactory == nil {
		options.crtClientFactory = &crtClientFactory{}
	}
	if options.discoveryClientFactory == nil {
		options.discoveryClientFactory = &discoveryClientFactory{}
	}

	if options.GetClientInterval.Seconds() == 0 {
		options.GetClientInterval = getClientDefaultInterval
	}
	if options.GetClientTimeout.Seconds() == 0 {
		options.GetClientTimeout = getClientDefaultTimeout
	}
	if options.OperationTimeout.Seconds() == 0 {
		options.OperationTimeout = operationDefaultTimeout
	}

	client := &client{
		kubeConfigPath:            kubeConfigPath,
		currentContext:            context,
		poller:                    options.poller,
		crtClientFactory:          options.crtClientFactory,
		discoveryClientFactory:    options.discoveryClientFactory,
		configLoadingRules:        rules,
		getClientInterval:         options.GetClientInterval,
		getClientTimeout:          options.GetClientTimeout,
		operationTimeout:          options.OperationTimeout,
		verificationClientFactory: options.verificationClientFactory,
	}
	err = client.updateK8sClients(context)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// CloneWithTimeout clones clusterctl client with updating timeout value
func (c *client) CloneWithTimeout(getClientTimeout time.Duration) Client {
	return &client{
		clientSet:              c.clientSet,
		discoveryClient:        c.discoveryClient,
		kubeConfigPath:         c.kubeConfigPath,
		currentContext:         c.currentContext,
		poller:                 c.poller,
		crtClientFactory:       c.crtClientFactory,
		discoveryClientFactory: c.discoveryClientFactory,
		configLoadingRules:     c.configLoadingRules,
		getClientInterval:      c.getClientInterval,
		getClientTimeout:       getClientTimeout,
		// copy the getClientTimeout to operationTimeout as well
		operationTimeout: getClientTimeout,
	}
}

func (c *client) loadKubeconfigAndEnsureContext(ctx string) ([]byte, error) {
	config, err := clientcmd.LoadFromFile(c.kubeConfigPath)
	if err != nil {
		return []byte{}, err
	}
	if ctx != "" {
		config.CurrentContext = ctx
	}

	return clientcmd.Write(*config)
}

func (c *client) updateK8sClients(ctx string) error {
	kubeConfigBytes, err := c.loadKubeconfigAndEnsureContext(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to read kube config")
	}

	clientSet, err := c.poller.PollImmediateWithGetter(c.getClientInterval, c.getClientTimeout, func() (interface{}, error) {
		return getK8sClients(kubeConfigBytes, c.crtClientFactory, c.discoveryClientFactory)
	})
	if err != nil {
		return errors.Wrap(err, "unable to get client")
	}

	k8sClients := clientSet.(k8ClientSet)
	c.clientSet = k8sClients.crtClient
	c.discoveryClient = k8sClients.discoveryClient
	c.currentContext = ctx

	return nil
}

func getCurrentContextFromKubeConfig(kubeConfig []byte) (string, error) {
	config, err := clientcmd.Load(kubeConfig)
	if err != nil {
		return "", errors.Wrap(err, "unable to load kubeconfig")
	}
	return config.CurrentContext, nil
}

// MergeConfigForCluster merge kubeconfig for cluster
func (c *client) MergeConfigForCluster(kubeConfig []byte, mergeFile string) error {
	newConfig, err := clientcmd.Load(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to load kubeconfig")
	}

	if mergeFile == "" {
		mergeFile = c.kubeConfigPath
	}

	if _, err := os.Stat(mergeFile); os.IsNotExist(err) {
		return clientcmd.WriteToFile(*newConfig, mergeFile)
	}

	dest, err := clientcmd.LoadFromFile(mergeFile)
	if err != nil {
		return errors.Wrap(err, "unable to load kube config")
	}

	ctx := dest.CurrentContext
	err = mergo.MergeWithOverwrite(dest, newConfig)
	if err != nil {
		return errors.Wrap(err, "failed to merge config")
	}
	dest.CurrentContext = ctx

	err = clientcmd.WriteToFile(*dest, mergeFile)
	if err != nil {
		return errors.Wrapf(err, "failed to write config to %s: %s", mergeFile, err)
	}
	return nil
}

func getK8sClients(kubeConfigBytes []byte, crtClientFactory CrtClientFactory, discoveryClientFactory DiscoveryClientFactory) (interface{}, error) {
	var crtClient crtclient.Client
	var discoveryClient discovery.DiscoveryInterface
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
	if err != nil {
		return nil, errors.Errorf("Unable to set up rest config due to : %v", err)
	}
	mapper, err := apiutil.NewDynamicRESTMapper(restConfig, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, errors.Errorf("Unable to set up rest mapper due to : %v", err)
	}

	crtClient, err = crtClientFactory.NewClient(restConfig, crtclient.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		// TODO catch real errors that doesn't warrant retrying and abort
		return nil, errors.Errorf("Error getting controller client due to : %v", err)
	}

	discoveryClient, err = discoveryClientFactory.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, errors.Errorf("Error getting discovery client due to : %v", err)
	}

	if _, err := discoveryClient.ServerVersion(); err != nil {
		return nil, errors.Errorf("Failed to invoke API on cluster : %v", err)
	}

	clientSet := k8ClientSet{
		crtClient:       crtClient,
		discoveryClient: discoveryClient,
	}

	return clientSet, nil
}

type k8ClientSet struct {
	crtClient       crtclient.Client
	discoveryClient discovery.DiscoveryInterface
}

// LoadCurrentKubeconfigBytes loads current kubeconfig bytes
func (c *client) LoadCurrentKubeconfigBytes() ([]byte, error) {
	return c.loadKubeconfigAndEnsureContext(c.currentContext)
}

func (c *client) GetRestConfigClient() (*rest.Config, error) {
	kubeConfigBytes, err := c.LoadCurrentKubeconfigBytes()
	if err != nil {
		return nil, err
	}
	return clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
}

//go:generate counterfeiter -o ../fakes/crtclientfactory.go --fake-name CrtClientFactory . CrtClientFactory

// CrtClientFactory is a interface to create controller runtime client
type CrtClientFactory interface {
	NewClient(config *rest.Config, options crtclient.Options) (crtclient.Client, error)
}

//go:generate counterfeiter -o ../fakes/discoveryclientfactory.go --fake-name DiscoveryClientFactory . DiscoveryClientFactory

// DiscoveryClientFactory is a interface to create discovery client
type DiscoveryClientFactory interface {
	NewDiscoveryClientForConfig(config *rest.Config) (discovery.DiscoveryInterface, error)
}

type discoveryClientFactory struct{}

// NewDiscoveryClientForConfig creates new discovery client factory
func (c *discoveryClientFactory) NewDiscoveryClientForConfig(restConfig *rest.Config) (discovery.DiscoveryInterface, error) {
	return discovery.NewDiscoveryClientForConfig(restConfig)
}

type crtClientFactory struct{}

// NewClient creates new clusterClient factory
func (c *crtClientFactory) NewClient(config *rest.Config, options crtclient.Options) (crtclient.Client, error) {
	return crtclient.New(config, options)
}

//go:generate counterfeiter -o ../fakes/clusterclientfactory.go --fake-name ClusterClientFactory . ClusterClientFactory

// ClusterClientFactory a factory for creating cluster clients
type ClusterClientFactory interface { // nolint
	NewClient(kubeConfigPath string, context string, options Options) (Client, error)
}

type clusterClientFactory struct{}

// NewClient creates new clusterclient
func (c *clusterClientFactory) NewClient(kubeConfigPath string, context string, options Options) (Client, error) { // nolint:gocritic
	return NewClient(kubeConfigPath, context, options)
}

// NewClusterClientFactory creates new clusterclient factory
func NewClusterClientFactory() ClusterClientFactory {
	return &clusterClientFactory{}
}
