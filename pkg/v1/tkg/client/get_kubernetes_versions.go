// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

// KubernetesVersionsInfo kubernetes version info struct
type KubernetesVersionsInfo struct {
	Versions []string `json:"versions" yaml:"versions"`
}

// GetKubernetesVersions get kubernetes version info
func (c *TkgClient) GetKubernetesVersions() (*KubernetesVersionsInfo, error) {
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

	return c.DoGetTanzuKubernetesReleases(regionalClusterClient)
}

// DoGetTanzuKubernetesReleases gets TKr
func (c *TkgClient) DoGetTanzuKubernetesReleases(regionalClusterClient clusterclient.Client) (*KubernetesVersionsInfo, error) {
	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err == nil && isPacific {
		availablek8sVersions, err := regionalClusterClient.GetPacificTanzuKubernetesReleases()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get supported kubernetes release versions for vSphere with Kubernetes clusters")
		}
		return &KubernetesVersionsInfo{
			Versions: availablek8sVersions,
		}, nil
	}
	availablek8sVersions, err := c.tkgBomClient.GetAvailableK8sVersionsFromBOMFiles()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get supported kubernetes version from BOM files")
	}
	return &KubernetesVersionsInfo{
		Versions: availablek8sVersions,
	}, nil
}
