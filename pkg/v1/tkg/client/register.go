// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// RegisterManagementClusterToTmc register management cluster to TMC
func (c *TkgClient) RegisterManagementClusterToTmc(clusterName, tmcRegistrationURL string) error {
	contexts, err := c.GetRegionContexts(clusterName)
	if err != nil || len(contexts) == 0 {
		return errors.Errorf("management cluster %s not found", clusterName)
	}
	currentRegion := contexts[0]

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := c.clusterClientFactory.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, " while registering management cluster")
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err == nil && isPacific {
		return errors.Errorf("cannot register a management cluster which is on vSphere 7.0 or above to Tanzu Mission Control. Please contact your vSphere administrator")
	}

	err = clusterClient.ApplyFile(tmcRegistrationURL)
	if err != nil {
		return errors.Wrap(err, "failed to register management cluster to Tanzu Mission Control")
	}

	log.Infof("Successfully registered management cluster '%s' to Tanzu Mission Control...\n", clusterName)
	return nil
}
