// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
)

// Scheme is a scheme that knows about TKG resources that are used to determine capabilities.
var Scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(runv1alpha1.AddToScheme(Scheme))
	utilruntime.Must(clusterctl.AddToScheme(Scheme))
	utilruntime.Must(corev1.AddToScheme(Scheme))
}

// DiscoveryClient allows clients to determine capabilities of a TKG cluster.
// Deprecated: This struct type will be removed in a future release.
type DiscoveryClient struct {
	k8sClient          client.Client
	clusterQueryClient *discovery.ClusterQueryClient
}

// NewDiscoveryClientForConfig returns a DiscoveryClient for a rest.Config.
// Deprecated: This function will be removed in a future release.
func NewDiscoveryClientForConfig(config *rest.Config) (*DiscoveryClient, error) {
	c, err := client.New(config, client.Options{Scheme: Scheme})
	if err != nil {
		return nil, err
	}
	queryClient, err := discovery.NewClusterQueryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &DiscoveryClient{k8sClient: c, clusterQueryClient: queryClient}, err
}

// NewDiscoveryClient returns a DiscoveryClient for a controller-runtime Client and ClusterQueryClient.
// Deprecated: This function will be removed in a future release.
func NewDiscoveryClient(c client.Client, clusterQueryClient *discovery.ClusterQueryClient) *DiscoveryClient {
	return &DiscoveryClient{k8sClient: c, clusterQueryClient: clusterQueryClient}
}
