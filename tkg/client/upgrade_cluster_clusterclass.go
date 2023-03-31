// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

func (c *TkgClient) DoClassyClusterUpgrade(regionalClusterClient clusterclient.Client,
	currentClusterClient clusterclient.Client, options *UpgradeClusterOptions) error {

	kubernetesVersion := options.KubernetesVersion
	tkrVersion := options.TkrVersion

	log.Infof("Upgrading kubernetes cluster to `%v` version, tkr version: `%s`", kubernetesVersion, tkrVersion)
	patchJSONString := fmt.Sprintf(`{"spec": {"topology": {"version": "%v"}}}`, tkrVersion)

	// Timeout set to 30 minutes because the continuousTKRDiscoverFreq for tkr-source-controller's fetcher is 10 minutes.
	// And kapp package reconcile frequency is 10 minutes.
	// Wait time should be longer than the fetcher's frequency of pulling tkrs plus the frequency of kapp package reconciliation.
	pollOptions := &clusterclient.PollOptions{Interval: upgradePatchInterval, Timeout: 30 * time.Minute}
	err := regionalClusterClient.PatchClusterObjectWithPollOptions(options.ClusterName, options.Namespace, patchJSONString, pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to patch kubernetes version to cluster")
	}

	log.Info("Waiting for kubernetes version to be updated for control plane nodes...")
	err = regionalClusterClient.WaitK8sVersionUpdateForCPNodes(options.ClusterName, options.Namespace, kubernetesVersion, currentClusterClient)
	if err != nil {
		return errors.Wrap(err, "error waiting for kubernetes version update for kubeadm control plane")
	}

	log.Info("Waiting for kubernetes version to be updated for worker nodes...")
	err = regionalClusterClient.WaitK8sVersionUpdateForWorkerNodes(options.ClusterName, options.Namespace, kubernetesVersion, currentClusterClient)
	if err != nil {
		return errors.Wrap(err, "error waiting for kubernetes version update for worker nodes")
	}

	if !options.IsRegionalCluster {
		// update autoscaler deployment if enabled
		err = regionalClusterClient.ApplyPatchForAutoScalerDeployment(c.tkgBomClient, options.ClusterName, options.KubernetesVersion, options.Namespace)
		if err != nil {
			return errors.Wrapf(err, "failed to upgrade autoscaler for cluster '%s'", options.ClusterName)
		}
	}

	return nil
}
