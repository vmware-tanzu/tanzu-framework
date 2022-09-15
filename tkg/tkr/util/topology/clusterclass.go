// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package topology provides helper functions to work with Cluster and ClusterClass topology features, such as variables.
package topology

import (
	"context"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetClusterClass gets ClusterClass for the cluster. Returns nil if the cluster is not classy.
// Pre-reqs: cluster != nil
func GetClusterClass(ctx context.Context, c client.Client, cluster *clusterv1.Cluster) (*clusterv1.ClusterClass, error) {
	if cluster.Spec.Topology == nil {
		return nil, nil
	}
	clusterClass := &clusterv1.ClusterClass{}
	if err := c.Get(ctx, client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cluster.Spec.Topology.Class,
	}, clusterClass); err != nil {
		return nil, err
	}
	return clusterClass, nil
}

// ClusterClassVariable finds a ClusterClass variable by name.
func ClusterClassVariable(clusterClass *clusterv1.ClusterClass, name string) *clusterv1.ClusterClassVariable {
	for i := range clusterClass.Spec.Variables {
		ccVar := &clusterClass.Spec.Variables[i]
		if ccVar.Name == name {
			return ccVar
		}
	}
	return nil
}
