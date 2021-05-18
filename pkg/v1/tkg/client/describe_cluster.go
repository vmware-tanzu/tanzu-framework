// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"time"

	"github.com/fabriziopandini/capi-conditions/cmd/kubectl-capi-tree/status"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/clusterclient"
)

// Scheme runtime scheme
var Scheme = runtime.NewScheme()

// DescribeTKGClustersOptions contains options supported by DescribeCluster
type DescribeTKGClustersOptions struct {
	ClusterName         string
	Namespace           string
	ShowOtherConditions string
	DisableNoEcho       bool
	DisableGroupObjects bool
	IsManagementCluster bool
}

// DescribeCluster describes cluster details and status
func (c *TkgClient) DescribeCluster(options DescribeTKGClustersOptions) (*status.ObjectTree, *clusterv1.Cluster, *clusterctlv1.ProviderList, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "unable to get cluster client while listing tkg clusters")
	}

	ctx := context.Background()

	cluster := &clusterv1.Cluster{}
	clusterKey := client.ObjectKey{
		Namespace: options.Namespace,
		Name:      options.ClusterName,
	}

	f := regionalClusterClient.GetClientSet()

	if err := f.Get(ctx, clusterKey, cluster); err != nil {
		return nil, nil, nil, err
	}

	objs, err := status.Discovery(ctx, f, cluster, status.DiscoverOptions{
		ShowOtherConditions: options.ShowOtherConditions,
		DisableNoEcho:       options.DisableNoEcho,
		DisableGroupObjects: options.DisableGroupObjects,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	if options.IsManagementCluster {
		installedProviders := &clusterctlv1.ProviderList{}
		err = regionalClusterClient.ListResources(installedProviders, &client.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		return objs, cluster, installedProviders, nil
	}

	return objs, cluster, nil, nil
}

// DescribeProvider describes provider information
func (c *TkgClient) DescribeProvider() (*clusterctlv1.ProviderList, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster client while listing tkg clusters")
	}

	installedProviders := &clusterctlv1.ProviderList{}
	err = regionalClusterClient.ListResources(installedProviders, &client.ListOptions{})
	if err != nil {
		return nil, err
	}
	return installedProviders, nil
}
