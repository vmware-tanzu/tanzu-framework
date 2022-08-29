// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

// GetPacificClusterObject return Pacific cluster object
func (c *TkgClient) GetPacificClusterObject(clusterName, namespace string) (*tkgsv1alpha2.TanzuKubernetesCluster, error) {
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
		return nil, errors.Wrap(err, "unable to get cluster client")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return nil, errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if !isPacific {
		return nil, errors.New("the management cluster is not 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}

	return regionalClusterClient.GetPacificClusterObject(clusterName, namespace)
}

// IsPacificRegionalCluster checks if the cluster pointed to by kubeconfig  is Pacific management cluster(supervisor)
func (c *TkgClient) IsPacificRegionalCluster() (bool, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return false, errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return false, errors.Wrap(err, "unable to get cluster client")
	}

	return regionalClusterClient.IsPacificRegionalCluster()
}
