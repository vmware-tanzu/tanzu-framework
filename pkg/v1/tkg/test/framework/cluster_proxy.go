// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterProxy struct {
	name           string
	kubeconfigPath string
	contextName    string
	scheme         *runtime.Scheme
}

func NewClusterProxy(name string, kubeconfigPath string, contextName string) *ClusterProxy {
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

func (p *ClusterProxy) GetClient() client.Client {
	config := p.GetRestConfig()

	c, err := client.New(config, client.Options{Scheme: p.scheme})
	Expect(err).ToNot(HaveOccurred(), "Failed to get controller-runtime client")

	return c
}

func (p *ClusterProxy) GetClientSet() *kubernetes.Clientset {
	restConfig := p.GetRestConfig()

	cs, err := kubernetes.NewForConfig(restConfig)
	Expect(err).ToNot(HaveOccurred(), "Failed to get client-go client")

	return cs
}

func (p *ClusterProxy) GetScheme() *runtime.Scheme {
	return p.scheme
}

func (p *ClusterProxy) GetClusterNodes() []corev1.Node {
	clientSet := p.GetClientSet()
	nodeList, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	return nodeList.Items
}

func (p *ClusterProxy) GetKubernetesVersion() string {
	clientSet := p.GetClientSet()
	version, err := clientSet.ServerVersion()
	Expect(err).ToNot(HaveOccurred())
	return version.String()
}

func (p *ClusterProxy) GetProviderVersions(ctx context.Context) map[string]string {
	client := p.GetClient()
	var providers clusterctlv1.ProviderList
	err := client.List(ctx, &providers)
	Expect(err).ToNot(HaveOccurred())

	providersMap := map[string]string{}
	for _, provider := range providers.Items {
		providersMap[provider.ProviderName] = provider.Version
	}

	return providersMap
}

func initScheme() *runtime.Scheme {
	sc := runtime.NewScheme()
	AddDefaultSchemes(sc)
	return sc
}
