// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// ListTKGClustersOptions ptions passed while getting a list of TKG Clusters
type ListTKGClustersOptions struct {
	ClusterName   string
	Namespace     string
	IncludeMC     bool
	AllNamespaces bool
}

// GetClusters returns list of cluster
func (t *tkgctl) GetClusters(options ListTKGClustersOptions) ([]client.ClusterInfo, error) {
	isPacific, err := t.tkgClient.IsPacificManagementCluster()
	if err != nil {
		return nil, err
	}
	var isTKGSClusterClassFeatureActivated bool
	if isPacific {
		isTKGSClusterClassFeatureActivated, err = t.featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(constants.ErrorMsgFeatureGateStatus, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace))
		}
	}
	namespace := options.Namespace
	allNamespaces := options.AllNamespaces
	// if allNamespaces is present, namespace in current context is ignored even if specified with --namespace.
	if allNamespaces {
		namespace = ""
	}
	listTKGClustersOptions := client.ListTKGClustersOptions{
		Namespace:                          namespace,
		IncludeMC:                          options.IncludeMC,
		IsTKGSClusterClassFeatureActivated: isTKGSClusterClassFeatureActivated,
	}
	clusters, err := t.tkgClient.ListTKGClusters(listTKGClustersOptions)
	if err != nil {
		return nil, err
	}

	sort.Slice(clusters, func(i, j int) bool {
		if clusters[i].Namespace < clusters[j].Namespace {
			return true
		}
		if clusters[i].Namespace > clusters[j].Namespace {
			return false
		}
		return clusters[i].Name < clusters[j].Name
	})

	return clusters, nil
}

func (t *tkgctl) IsClusterExists(clustername, namespace string) (bool, error) {
	clusters, err := t.GetClusters(ListTKGClustersOptions{
		Namespace: namespace,
		IncludeMC: true,
	})
	if err != nil {
		return false, err
	}
	for index := range clusters {
		if clusters[index].Name == clustername && clusters[index].Namespace == namespace {
			return true, nil
		}
	}
	return false, nil
}
