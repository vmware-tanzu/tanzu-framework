// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package framework implements the test framework.
package framework

import (
	"context"

	. "github.com/onsi/gomega" // nolint:golint,stylecheck

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterProxy hold information to connect to a cluster
type ClusterProxy struct {
	name           string
	kubeconfigPath string
	contextName    string
	scheme         *runtime.Scheme
}

// NewClusterProxy returns clusterProxy
func NewClusterProxy(name, kubeconfigPath, contextName string) *ClusterProxy {
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	}

	proxy := &ClusterProxy{
		name:           name,
		kubeconfigPath: kubeconfigPath,
		contextName:    contextName,
		scheme:         initScheme(),
	}

	return proxy
}

// GetRestConfig returns the RestConfig of a cluster
func (p *ClusterProxy) GetRestConfig() *rest.Config {
	config, err := clientcmd.LoadFromFile(p.kubeconfigPath)
	Expect(err).ToNot(HaveOccurred(), "Failed to load Kubeconfig file from %q", p.kubeconfigPath)

	configOverrides := &clientcmd.ConfigOverrides{}
	if p.contextName != "" {
		configOverrides.CurrentContext = p.contextName
	}

	restConfig, err := clientcmd.NewDefaultClientConfig(*config, configOverrides).ClientConfig()
	Expect(err).ToNot(HaveOccurred(), "Failed to get ClientConfig for context %q from %q", p.contextName, p.kubeconfigPath)

	restConfig.UserAgent = "tkg-cli-e2e"
	return restConfig
}

// GetClient gets the Client of a cluster
func (p *ClusterProxy) GetClient() client.Client {
	config := p.GetRestConfig()

	c, err := client.New(config, client.Options{Scheme: p.scheme})
	Expect(err).ToNot(HaveOccurred(), "Failed to get controller-runtime client")

	return c
}

// GetClientSet gets the ClientSet of a cluster
func (p *ClusterProxy) GetClientSet() *kubernetes.Clientset {
	restConfig := p.GetRestConfig()

	cs, err := kubernetes.NewForConfig(restConfig)
	Expect(err).ToNot(HaveOccurred(), "Failed to get client-go client")

	return cs
}

// GetScheme returns scheme
func (p *ClusterProxy) GetScheme() *runtime.Scheme {
	return p.scheme
}

// GetClusterNodes gets the cluster Nodes
func (p *ClusterProxy) GetClusterNodes() []corev1.Node {
	clientSet := p.GetClientSet()
	nodeList, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	return nodeList.Items
}

// GetKubernetesVersion gets the k8s version
func (p *ClusterProxy) GetKubernetesVersion() string {
	clientSet := p.GetClientSet()
	version, err := clientSet.ServerVersion()
	Expect(err).ToNot(HaveOccurred())
	return version.String()
}

// GetProviderVersions gets the TKG provider versions
func (p *ClusterProxy) GetProviderVersions(ctx context.Context) map[string]string {
	c := p.GetClient()
	var providers clusterctlv1.ProviderList
	err := c.List(ctx, &providers)
	Expect(err).ToNot(HaveOccurred())

	providersMap := map[string]string{}
	for i := range providers.Items {
		providersMap[providers.Items[i].ProviderName] = providers.Items[i].Version
	}

	return providersMap
}

func initScheme() *runtime.Scheme {
	sc := runtime.NewScheme()
	AddDefaultSchemes(sc)
	return sc
}
