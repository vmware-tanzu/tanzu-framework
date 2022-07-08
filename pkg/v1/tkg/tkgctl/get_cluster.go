// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// ListTKGClustersOptions ptions passed while getting a list of TKG Clusters
type ListTKGClustersOptions struct {
	ClusterName string
	Namespace   string
	IncludeMC   bool
}

// GetClusters returns list of cluster
func (t *tkgctl) GetClusters(options ListTKGClustersOptions) ([]client.ClusterInfo, error) {
	isPacific, err := t.tkgClient.IsPacificManagementCluster()
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine if management cluster is on vSphere with Tanzu")
	}
	var isTKGSClusterClassFeatureActivated bool
	if isPacific {
		isTKGSClusterClassFeatureActivated, err = t.featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(constants.ErrorMsgFeatureGateStatus, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace))
		}
	}
	listTKGClustersOptions := client.ListTKGClustersOptions{
		Namespace:                          options.Namespace,
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
