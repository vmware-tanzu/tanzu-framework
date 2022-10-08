// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cluster provides functions to manipulate the package
package cluster

import (
	"context"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	capdiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/constants"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = cliv1alpha1.AddToScheme(scheme)
}

// Client provides various aspects of interaction with a Kubernetes cluster provisioned by TKG
//
//go:generate counterfeiter -o ../fakes/clusterclient_fake.go --fake-name ClusterClient . Client
type Client interface {
	// ListCLIPluginResources lists CLIPlugin resources across all namespaces
	ListCLIPluginResources() ([]cliv1alpha1.CLIPlugin, error)
	// VerifyCLIPluginCRD returns true if CRD exists else return false
	VerifyCLIPluginCRD() (bool, error)
	// GetCLIPluginImageRepositoryOverride returns map of image repository override
	GetCLIPluginImageRepositoryOverride() (map[string]string, error)

	// BuildClusterQuery builds ClusterQuery with Dynamic client and Discovery client
	BuildClusterQuery() (*capdiscovery.ClusterQuery, error)
}

//go:generate counterfeiter -o ../fakes/CrtClient_fake.go --fake-name CrtClientFake . CrtClient
//go:generate counterfeiter -o ../fakes/discoveryclusterclient_fake.go --fake-name DiscoveryClient . DiscoveryClient

// CrtClient clientset interface
type CrtClient interface {
	NewClient(config *rest.Config, options crtclient.Options) (crtclient.Client, error)
	ListObjects(ctx context.Context, cliPlugins crtclient.ObjectList, listOptions *crtclient.ListOptions) error
}
type CrtClientImpl struct {
	crtClient crtclient.Client
}

func (c *CrtClientImpl) NewClient(config *rest.Config, options crtclient.Options) (crtclient.Client, error) {
	crtClient, err := crtclient.New(config, options)
	c.crtClient = crtClient
	return crtClient, err
}
func (c *CrtClientImpl) ListObjects(ctx context.Context, cliPlugins crtclient.ObjectList, listOptions *crtclient.ListOptions) error {
	return c.crtClient.List(ctx, cliPlugins, listOptions)
}

// DiscoveryClient discovery client interface
type DiscoveryClient interface {
	discovery.DiscoveryInterface
}

type DynamicClient interface {
	dynamic.Interface
}

type client struct {
	CrtClient       CrtClient
	DiscoveryClient DiscoveryClient
	DynamicClient   DynamicClient
	kubeConfigPath  string
	currentContext  string
}

// NewClient creates new clusterclient from kubeconfig file and poller
// if kubeconfig path is empty it gets default path
// if options.poller is nil it creates default poller. You should only pass custom poller for unit testing
// if options.crtClientFactory is nil it creates default CrtClientFactory
func NewClient(kubeConfigPath, contextStr string, options Options) (Client, error) {
	var err error
	var rules *clientcmd.ClientConfigLoadingRules
	if kubeConfigPath == "" {
		rules = clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfigPath = rules.GetDefaultFilename()
	}
	InitializeOptions(&options)
	client := &client{
		kubeConfigPath: kubeConfigPath,
		currentContext: contextStr,
	}

	err = client.getK8sClients(options.CrtClient, options.DiscoveryClientFactory, options.DynamicClientFactory)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// InitializeOptions initializes the options
func InitializeOptions(options *Options) {
	if options.CrtClient == nil {
		options.CrtClient = &CrtClientImpl{}
	}
	if options.DiscoveryClientFactory == nil {
		options.DiscoveryClientFactory = &discoveryClientFactory{}
	}

	if options.DynamicClientFactory == nil {
		options.DynamicClientFactory = &dynamicClientFactory{}
	}
}

// VerifyCLIPluginCRD returns true if CRD exists else return false
func (c *client) VerifyCLIPluginCRD() (bool, error) {
	// Since we're looking up API types via discovery, we don't need the dynamic client.
	cqc, err := c.BuildClusterQuery()
	if err != nil {
		return false, err
	}

	// Execute returns combined result of all queries.
	return cqc.Execute() // return (found, err) response
}

func (c *client) BuildClusterQuery() (*capdiscovery.ClusterQuery, error) {
	clusterQueryClient, err := capdiscovery.NewClusterQueryClient(c.DynamicClient, c.DiscoveryClient)
	if err != nil {
		return nil, err
	}

	var queryObject = capdiscovery.Group("cliPlugins", cliv1alpha1.GroupVersionKindCLIPlugin.Group).WithResource("cliplugins")

	// Build query client.
	return clusterQueryClient.Query(queryObject), nil
}

// ListCLIPluginResources lists CLIPlugin resources across all namespaces
func (c *client) ListCLIPluginResources() ([]cliv1alpha1.CLIPlugin, error) {
	var cliPlugins cliv1alpha1.CLIPluginList
	err := c.CrtClient.ListObjects(context.TODO(), &cliPlugins, &crtclient.ListOptions{Namespace: ""})
	if err != nil {
		return nil, err
	}
	return cliPlugins.Items, nil
}

// GetCLIPluginImageRepositoryOverride returns map of image repository override
func (c *client) GetCLIPluginImageRepositoryOverride() (map[string]string, error) {
	cmList := &corev1.ConfigMapList{}

	labelMatch, _ := labels.NewRequirement(constants.CLIPluginImageRepositoryOverrideLabel, selection.Exists, []string{})
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*labelMatch)

	err := c.CrtClient.ListObjects(context.TODO(), cmList, &crtclient.ListOptions{Namespace: constants.TanzuCLISystemNamespace, LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	return ConsolidateImageRepoMaps(cmList)
}

func ConsolidateImageRepoMaps(cmList *corev1.ConfigMapList) (map[string]string, error) {
	imageRepoMap := make(map[string]string)

	for i := range cmList.Items {
		mapString, ok := cmList.Items[i].Data["imageRepoMap"]
		if !ok {
			continue
		}
		irm := make(map[string]string)

		_ = yaml.Unmarshal([]byte(mapString), &irm)
		for k, v := range irm {
			if _, exists := imageRepoMap[k]; exists {
				return nil, errors.Errorf("multiple references of image repository %q found while doing image repository override", k)
			}
			imageRepoMap[k] = v
		}
	}
	return imageRepoMap, nil
}

func (c *client) getK8sClients(crtClient CrtClient, discoveryClientFactory DiscoveryClientFactory, dynamicClientFactory DynamicClientFactory) error {
	var discoveryClient discovery.DiscoveryInterface
	config, err := clientcmd.LoadFromFile(c.kubeConfigPath)
	if err != nil {
		return errors.Errorf("Failed to load Kubeconfig file from %q", c.kubeConfigPath)
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	if c.currentContext != "" {
		configOverrides.CurrentContext = c.currentContext
	}

	restConfig, err := clientcmd.NewDefaultClientConfig(*config, configOverrides).ClientConfig()
	if err != nil {
		return errors.Errorf("Unable to set up rest config due to : %v", err)
	}
	// As there are many registered resources in the cluster, set the values for the maximum number of
	// queries per second and the maximum burst for throttle to a high value to avoid throttling of messages
	restConfig.QPS = constants.DefaultQPS
	restConfig.Burst = constants.DefaultBurst
	mapper, err := apiutil.NewDynamicRESTMapper(restConfig, apiutil.WithLazyDiscovery)
	if err != nil {
		return errors.Errorf("Unable to set up rest mapper due to : %v", err)
	}

	_, err = crtClient.NewClient(restConfig, crtclient.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		// TODO catch real errors that doesn't warrant retrying and abort
		return errors.Errorf("Error getting controller client due to : %v", err)
	}

	discoveryClient, err = discoveryClientFactory.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return errors.Errorf("Error getting discovery client due to : %v", err)
	}

	if _, err := discoveryClientFactory.ServerVersion(discoveryClient); err != nil {
		return errors.Errorf("Failed to invoke API on cluster : %v", err)
	}

	dynamicClient, err := dynamicClientFactory.NewDynamicClientForConfig(restConfig)
	if err != nil {
		return errors.Errorf("Error getting dynamic client due to : %v", err)
	}
	c.CrtClient = crtClient
	c.DiscoveryClient = discoveryClient
	c.DynamicClient = dynamicClient

	return nil
}

//go:generate counterfeiter -o ../fakes/discoveryclientfactory_fake.go --fake-name DiscoveryClientFactory . DiscoveryClientFactory

// DiscoveryClientFactory is a interface to create discovery client
type DiscoveryClientFactory interface {
	NewDiscoveryClientForConfig(config *rest.Config) (discovery.DiscoveryInterface, error)
	ServerVersion(discoveryClient discovery.DiscoveryInterface) (*version.Info, error)
}

type discoveryClientFactory struct {
}

func NewDiscoveryClientFactory() DiscoveryClientFactory {
	return &discoveryClientFactory{}
}

// NewDiscoveryClientForConfig creates new discovery client factory
func (c *discoveryClientFactory) NewDiscoveryClientForConfig(config *rest.Config) (discovery.DiscoveryInterface, error) {
	return discovery.NewDiscoveryClientForConfig(config)
}

func (c *discoveryClientFactory) ServerVersion(discoveryClient discovery.DiscoveryInterface) (*version.Info, error) {
	return discoveryClient.ServerVersion()
}

//go:generate counterfeiter -o ../fakes/dynamicclientfactory_fake.go --fake-name DynamicClientFactory . DynamicClientFactory

// DynamicClientFactory is a interface to create adynamic client
type DynamicClientFactory interface {
	NewDynamicClientForConfig(config *rest.Config) (dynamic.Interface, error)
}

type dynamicClientFactory struct{}

// NewDynamicClientForConfig creates a new discovery client factory
func (c *dynamicClientFactory) NewDynamicClientForConfig(restConfig *rest.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(restConfig)
}

// Options provides way to customize creation of clusterClient
type Options struct {
	CrtClient              CrtClient
	DiscoveryClientFactory DiscoveryClientFactory
	DynamicClientFactory   DynamicClientFactory
}

// NewOptions returns new options
func NewOptions(crtClient CrtClient, discoveryClientFactory DiscoveryClientFactory, dynamicClientFactory DynamicClientFactory) Options {
	return Options{
		CrtClient:              crtClient,
		DiscoveryClientFactory: discoveryClientFactory,
		DynamicClientFactory:   dynamicClientFactory,
	}
}

//go:generate counterfeiter -o ../fakes/clusterclientfactory_fake.go --fake-name ClusterClientFactory . ClusterClientFactory

// ClusterClientFactory a factory for creating cluster clients
type ClusterClientFactory interface {
	NewClient(kubeConfigPath, context string, options Options) (Client, error)
}

type clusterClientFactory struct{}

// NewClient creates new clusterclient
func (c *clusterClientFactory) NewClient(kubeConfigPath, contextStr string, options Options) (Client, error) {
	return NewClient(kubeConfigPath, contextStr, options)
}

// NewClusterClientFactory creates new clusterclient factory
func NewClusterClientFactory() ClusterClientFactory {
	return &clusterClientFactory{}
}
