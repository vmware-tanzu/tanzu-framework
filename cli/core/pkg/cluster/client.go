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
//go:generate counterfeiter -o ../fakes/clusterclient.go --fake-name ClusterClient . Client
type Client interface {
	// ListCLIPluginResources lists CLIPlugin resources across all namespaces
	ListCLIPluginResources() ([]cliv1alpha1.CLIPlugin, error)
	// VerifyCLIPluginCRD returns true if CRD exists else return false
	VerifyCLIPluginCRD() (bool, error)
	// GetCLIPluginImageRepositoryOverride returns map of image repository override
	GetCLIPluginImageRepositoryOverride() (map[string]string, error)
}

//go:generate counterfeiter -o ../fakes/crtclusterclient.go --fake-name CRTClusterClient . CrtClient
//go:generate counterfeiter -o ../fakes/discoveryclusterclient.go --fake-name DiscoveryClient . DiscoveryClient

// CrtClient clientset interface
type CrtClient interface {
	crtclient.Client
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
func NewClient(kubeConfigPath string, context string, options Options) (Client, error) { //nolint:gocritic
	var err error
	var rules *clientcmd.ClientConfigLoadingRules
	if kubeConfigPath == "" {
		rules = clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfigPath = rules.GetDefaultFilename()
	}

	if options.crtClientFactory == nil {
		options.crtClientFactory = &crtClientFactory{}
	}
	if options.discoveryClientFactory == nil {
		options.discoveryClientFactory = &discoveryClientFactory{}
	}

	if options.dynamicClientFactory == nil {
		options.dynamicClientFactory = &dynamicClientFactory{}
	}

	client := &client{
		kubeConfigPath: kubeConfigPath,
		currentContext: context,
	}

	err = client.getK8sClients(options.crtClientFactory, options.discoveryClientFactory, options.dynamicClientFactory)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// VerifyCLIPluginCRD returns true if CRD exists else return false
func (c *client) VerifyCLIPluginCRD() (bool, error) {
	// Since we're looking up API types via discovery, we don't need the dynamic client.
	clusterQueryClient, err := capdiscovery.NewClusterQueryClient(c.DynamicClient, c.DiscoveryClient)
	if err != nil {
		return false, err
	}

	var queryObject = capdiscovery.Group("cliPlugins", cliv1alpha1.GroupVersionKindCLIPlugin.Group).WithResource("cliplugins")

	// Build query client.
	cqc := clusterQueryClient.Query(queryObject)

	// Execute returns combined result of all queries.
	return cqc.Execute() // return (found, err) response
}

// ListCLIPluginResources lists CLIPlugin resources across all namespaces
func (c *client) ListCLIPluginResources() ([]cliv1alpha1.CLIPlugin, error) {
	var cliPlugins cliv1alpha1.CLIPluginList
	err := c.CrtClient.List(context.TODO(), &cliPlugins, &crtclient.ListOptions{Namespace: ""})
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

	err := c.CrtClient.List(context.TODO(), cmList, &crtclient.ListOptions{Namespace: constants.TanzuCLISystemNamespace, LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	imageRepoMap := make(map[string]string)

	//nolint:gocritic
	for _, cm := range cmList.Items {
		mapString, ok := cm.Data["imageRepoMap"]
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

func (c *client) getK8sClients(crtClientFactory CrtClientFactory, discoveryClientFactory DiscoveryClientFactory, dynamicClientFactory DynamicClientFactory) error {
	var crtClient crtclient.Client
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

	crtClient, err = crtClientFactory.NewClient(restConfig, crtclient.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		// TODO catch real errors that doesn't warrant retrying and abort
		return errors.Errorf("Error getting controller client due to : %v", err)
	}

	discoveryClient, err = discoveryClientFactory.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return errors.Errorf("Error getting discovery client due to : %v", err)
	}

	if _, err := discoveryClient.ServerVersion(); err != nil {
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

// The below factory interfaces are to provide fake methods to enable unit tests
//go:generate counterfeiter -o ../fakes/crtclientfactory.go --fake-name CrtClientFactory . CrtClientFactory

// CrtClientFactory is a interface to create controller runtime client
type CrtClientFactory interface {
	NewClient(config *rest.Config, options crtclient.Options) (crtclient.Client, error)
}

type crtClientFactory struct{}

// NewClient creates new clusterClient factory
func (c *crtClientFactory) NewClient(config *rest.Config, options crtclient.Options) (crtclient.Client, error) {
	return crtclient.New(config, options)
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

//go:generate counterfeiter -o ../fakes/dynamicclientfactory.go --fake-name DynamicClientFactory . DynamicClientFactory

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
	crtClientFactory       CrtClientFactory
	discoveryClientFactory DiscoveryClientFactory
	dynamicClientFactory   DynamicClientFactory
}

// NewOptions returns new options
func NewOptions(crtClientFactory CrtClientFactory, discoveryClientFactory DiscoveryClientFactory) Options {
	return Options{
		crtClientFactory:       crtClientFactory,
		discoveryClientFactory: discoveryClientFactory,
	}
}

//go:generate counterfeiter -o ../fakes/clusterclientfactory.go --fake-name ClusterClientFactory . ClusterClientFactory

// ClusterClientFactory a factory for creating cluster clients
type ClusterClientFactory interface {
	NewClient(kubeConfigPath, context string, options Options) (Client, error)
}

type clusterClientFactory struct{}

// NewClient creates new clusterclient
func (c *clusterClientFactory) NewClient(kubeConfigPath, context string, options Options) (Client, error) { //nolint:gocritic
	return NewClient(kubeConfigPath, context, options)
}

// NewClusterClientFactory creates new clusterclient factory
func NewClusterClientFactory() ClusterClientFactory {
	return &clusterClientFactory{}
}
