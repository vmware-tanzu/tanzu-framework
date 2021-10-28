// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"strings"

	"github.com/pkg/errors"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctltree "sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// DescribeTKGClustersOptions options that can be passed while requesting to describe a cluster
type DescribeTKGClustersOptions struct {
	ClusterName         string
	Namespace           string
	ShowOtherConditions string
	ShowDetails         bool
	ShowGroupMembers    bool
}

// DescribeClusterResult the result object for when the cluster's description is returned
type DescribeClusterResult struct {
	Objs               *clusterctltree.ObjectTree
	Cluster            *clusterv1.Cluster
	InstalledProviders *clusterctlv1.ProviderList
	ClusterInfo        client.ClusterInfo
}

// DescribeCluster returns list of cluster
func (t *tkgctl) DescribeCluster(options DescribeTKGClustersOptions) (DescribeClusterResult, error) {
	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	DescribeTKGClustersOptions := client.DescribeTKGClustersOptions{
		ClusterName:         options.ClusterName,
		Namespace:           options.Namespace,
		ShowOtherConditions: options.ShowOtherConditions,
		ShowDetails:         options.ShowDetails,
		ShowGroupMembers:    options.ShowGroupMembers,
		IsManagementCluster: false,
	}

	results := DescribeClusterResult{}

	listTKGClustersOptions := client.ListTKGClustersOptions{
		Namespace: options.Namespace,
		IncludeMC: true,
	}

	clusters, err := t.tkgClient.ListTKGClusters(listTKGClustersOptions)
	if err != nil {
		return results, err
	}

	for i := 0; i < len(clusters); i++ {
		if clusters[i].Name == options.ClusterName {
			results.ClusterInfo = clusters[i]

			clusterRoles := "<none>"
			if len(clusters[i].Roles) != 0 {
				clusterRoles = strings.Join(clusters[i].Roles, ",")
			}
			if clusterRoles == "management" {
				DescribeTKGClustersOptions.IsManagementCluster = true

				// management cluster will point to a kind cluster if infrastructure provisioning
				// or 'clusterctl move' fails.

				// setting status of the management cluster to "failed" if the mc is a kind cluster.
				isFailure, err := t.tkgClient.IsManagementClusterAKindCluster(options.ClusterName)
				if err != nil {
					return results, err
				}

				if isFailure {
					results.ClusterInfo.Status = "failed"
				}
			}
		}
	}

	isPacific, err := t.IsPacificRegionalCluster()
	if err != nil {
		return results, errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}

	// TODO: Can be removed when TKGS and TKGm converge to the same CAPI version.
	// https://github.com/vmware-tanzu/tanzu-framework/issues/1063
	if isPacific {
		return results, nil
	}

	objs, cluster, installedProviders, err := t.tkgClient.DescribeCluster(DescribeTKGClustersOptions)
	if err != nil {
		return results, err
	}
	results.Objs = objs
	results.Cluster = cluster
	results.InstalledProviders = installedProviders

	return results, nil
}

func (t *tkgctl) DescribeProviders() (*clusterctlv1.ProviderList, error) {
	installedProviders, err := t.tkgClient.DescribeProvider()
	if err != nil {
		return nil, err
	}

	return installedProviders, nil
}
