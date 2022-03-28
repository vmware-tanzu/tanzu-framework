// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package topology provides helper functions to work with Cluster and ClusterClass topology features, such as variables.
package topology

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

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
