// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"sort"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
)

// ListTKGClustersOptions ptions passed while getting a list of TKG Clusters
type ListTKGClustersOptions struct {
	ClusterName string
	Namespace   string
	IncludeMC   bool
}

// GetClusters returns list of cluster
func (t *tkgctl) GetClusters(options ListTKGClustersOptions) ([]client.ClusterInfo, error) {
	listTKGClustersOptions := client.ListTKGClustersOptions{
		Namespace: options.Namespace,
		IncludeMC: options.IncludeMC,
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
