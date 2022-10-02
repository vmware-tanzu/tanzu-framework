// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// UpgradeClusterOptions options for upgrade cluster
type UpgradeClusterOptions struct {
	ClusterName string
	Namespace   string
	TkrVersion  string
	SkipPrompt  bool
	Timeout     time.Duration
	OSName      string
	OSVersion   string
	OSArch      string

	// VSphereTemplateName deprecated please use
	// OSName, OSVersion and OSArch config variable
	// to filter vSphereTemplate
	VSphereTemplateName string
	// Tanzu edition (either tce or tkg)
	Edition string
}

// UpgradeCluster upgrade tkg workload cluster
//
//nolint:gocritic
func (t *tkgctl) UpgradeCluster(options UpgradeClusterOptions) error {
	var err error
	var k8sVersion string

	if logPath, err := t.getAuditLogPath(options.ClusterName); err == nil {
		log.SetAuditLog(logPath)
	}

	// upgrade requires minimum 15 minutes timeout
	minTimeoutReq := 15 * time.Minute
	if options.Timeout < minTimeoutReq {
		log.V(6).Infof("timeout duration of at least 15 minutes is required, using default timeout %v", constants.DefaultLongRunningOperationTimeout)
		options.Timeout = constants.DefaultLongRunningOperationTimeout
	}
	defer t.restoreAfterSettingTimeout(options.Timeout)()

	isPacific, err := t.tkgClient.IsPacificManagementCluster()
	if err != nil {
		return err
	}

	if isPacific {
		// For TKGS kubernetesVersion will be same as TkrVersion
		k8sVersion = options.TkrVersion
	} else {
		options.TkrVersion, k8sVersion, err = t.getAndDownloadTkrIfNeeded(options.TkrVersion)
		if err != nil {
			return errors.Wrapf(err, "unable to determine the TKr version and kubernetes version based on '%v'", options.TkrVersion)
		}
	}

	// if --yes is set, kick off the upgrade process without waiting for confirmation
	if !options.SkipPrompt {
		if err := askForConfirmation(fmt.Sprintf("Upgrading workload cluster '%s' to kubernetes version '%s'. Are you sure?", options.ClusterName, k8sVersion)); err != nil {
			return err
		}
	}

	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}
	upgradeClusterOption := client.UpgradeClusterOptions{
		ClusterName:         options.ClusterName,
		Namespace:           options.Namespace,
		KubernetesVersion:   k8sVersion,
		TkrVersion:          options.TkrVersion,
		Kubeconfig:          t.kubeconfig,
		IsRegionalCluster:   false,
		VSphereTemplateName: options.VSphereTemplateName,
		OSName:              options.OSName,
		OSVersion:           options.OSVersion,
		OSArch:              options.OSArch,
		Edition:             options.Edition,
		IsTKGSCluster:       isPacific,
	}
	err = t.tkgClient.UpgradeCluster(&upgradeClusterOption)
	if err != nil {
		return err
	}

	log.Infof("Cluster '%s' successfully upgraded to kubernetes version '%s'\n", upgradeClusterOption.ClusterName, upgradeClusterOption.KubernetesVersion)
	return nil
}
