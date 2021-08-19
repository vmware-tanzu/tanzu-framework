// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/api/tmc/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// DeRegisterManagementClusterFromTmc performs steps to register management cluster to TMC
func (c *TkgClient) DeRegisterManagementClusterFromTmc(clusterName string) error {
	log.Infof("Deregistering management cluster %s from Tanzu Mission Control...\n", clusterName)
	contexts, err := c.GetRegionContexts(clusterName)
	if err != nil || len(contexts) == 0 {
		return errors.Errorf("management cluster %s not found", clusterName)
	}
	currentRegion := contexts[0]

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while deregistering management cluster")
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err == nil && isPacific {
		return errors.Errorf("cannot deregister a management cluster which is on vSphere 7.0 or above to Tanzu Mission Control")
	}

	// check if the management cluster is registered with TMC or not
	registered, err := clusterClient.IsClusterRegisteredToTMC()
	if err != nil {
		return errors.Wrapf(err, "failed to deregister the management cluster '%s' from Tanzu Mission Control", clusterName)
	}
	if !registered {
		return errors.Errorf("management cluster '%s' is not registered to Tanzu Mission Control", clusterName)
	}

	// TODO - Filter the extension CRs and delete the ones which are not required. More info is needed from the TMC agent team. Deleting all extension CRs for now
	var extensions v1alpha1.ExtensionList
	err = clusterClient.ListResources(&extensions, &crtclient.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to deregister the management cluster '%s' from Tanzu Mission Control", clusterName)
	}

	for i := range extensions.Items {
		extension := extensions.Items[i]
		err = clusterClient.DeleteResource(&extension)
		if err != nil {
			return err
		}
	}

	log.Infof("successfully deregistered management cluster '%s' from Tanzu Mission Control\n", clusterName)
	return nil
}
