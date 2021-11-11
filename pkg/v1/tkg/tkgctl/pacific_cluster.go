// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
)

// GetPacificClusterObject return Pacific cluster object
func (t *tkgctl) GetPacificClusterObject(clusterName, namespace string) (*tkgsv1alpha2.TanzuKubernetesCluster, error) {
	return t.tkgClient.GetPacificClusterObject(clusterName, namespace)
}

// GetPacificClusterObject checks if the cluster pointed to by kubeconfig  is Pacific management cluster(supervisor)
func (t *tkgctl) IsPacificRegionalCluster() (bool, error) {
	return t.tkgClient.IsPacificRegionalCluster()
}
