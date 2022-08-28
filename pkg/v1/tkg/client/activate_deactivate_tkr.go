// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

// ActivateTanzuKubernetesReleases activates TKr
func (c *TkgClient) ActivateTanzuKubernetesReleases(tkrName string) error {
	regionalClusterClient, err := c.getRegionalClusterClient()
	if err != nil {
		return errors.Wrap(err, "failed to get management cluster client")
	}

	return regionalClusterClient.ActivateTanzuKubernetesReleases(tkrName)
}

// DeactivateTanzuKubernetesReleases deactivates TKR
func (c *TkgClient) DeactivateTanzuKubernetesReleases(tkrName string) error {
	regionalClusterClient, err := c.getRegionalClusterClient()
	if err != nil {
		return errors.Wrap(err, "failed to get management cluster client")
	}

	return regionalClusterClient.DeactivateTanzuKubernetesReleases(tkrName)
}

func (c *TkgClient) getRegionalClusterClient() (clusterclient.Client, error) {
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
	return regionalClusterClient, nil
}
