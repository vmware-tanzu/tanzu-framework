// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// GetClusterPinnipedInfoOptions options that can be passed while getting pinniped info
type GetClusterPinnipedInfoOptions struct {
	ClusterName         string
	Namespace           string
	IsManagementCluster bool
}

// GetClusterPinnipedInfo returns the cluster and pinniped information
func (t *tkgctl) GetClusterPinnipedInfo(options GetClusterPinnipedInfoOptions) (*client.ClusterPinnipedInfo, error) {
	if !options.IsManagementCluster && options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}
	getClustersPinnipedInfoOptions := client.GetClusterPinnipedInfoOptions{
		ClusterName:         options.ClusterName,
		Namespace:           options.Namespace,
		IsManagementCluster: options.IsManagementCluster,
	}

	clusterPinnipedInfo, err := t.tkgClient.GetClusterPinnipedInfo(getClustersPinnipedInfoOptions)
	if err != nil {
		return nil, err
	}
	return clusterPinnipedInfo, nil
}
