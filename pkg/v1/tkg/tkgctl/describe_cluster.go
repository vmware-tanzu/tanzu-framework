// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctltree "sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
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

	isPacific, err := t.tkgClient.IsPacificManagementCluster()
	if err != nil {
		return results, errors.Wrap(err, "unable to determine if management cluster is on vSphere with Tanzu")
	}

	var isTKGSClusterClassFeatureActivated bool
	if isPacific {
		isTKGSClusterClassFeatureActivated, err = t.featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
		if err != nil {
			return results, errors.Wrap(err, fmt.Sprintf(constants.ErrorMsgFeatureGateStatus, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace))
		}
	}

	listTKGClustersOptions := client.ListTKGClustersOptions{
		Namespace:                          options.Namespace,
		IncludeMC:                          true,
		IsTKGSClusterClassFeatureActivated: isTKGSClusterClassFeatureActivated,
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

	objs, cluster, installedProviders, err := t.tkgClient.DescribeCluster(DescribeTKGClustersOptions)
	if err != nil {
		// If it is pacific(TKGS), it would be the best effort to return the objectTree and cluster, so if there is an error
		// fetching these objects, return empty objects without error.
		if isPacific {
			log.V(5).Infof("Failed to get cluster ObjectTree/cluster objects(so detailed(tree) view of cluster resources may be affected), reason: %v", err)
			return results, nil
		}
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
