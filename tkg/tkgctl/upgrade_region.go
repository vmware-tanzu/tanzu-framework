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
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

// UpgradeRegionOptions upgrade management cluster options
type UpgradeRegionOptions struct {
	ClusterName string
	SkipPrompt  bool
	Timeout     time.Duration
	OSName      string
	OSVersion   string
	OSArch      string

	// VSphereTemplateName (deprecated: please use OSName, OSVersion and OSArch
	// config variables to filter vSphereTemplate)
	VSphereTemplateName string
	// Tanzu edition (either tce or tkg)
	Edition string
}

// UpgradeRegion upgrades management cluster
//
//nolint:gocritic
func (t *tkgctl) UpgradeRegion(options UpgradeRegionOptions) error {
	var err error

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

	tkrBomConfig, err := t.tkgBomClient.GetDefaultTkrBOMConfiguration()
	if err != nil {
		return errors.Wrap(err, "unable to get default TKr BoM")
	}
	kubernetesVersion, err := tkgconfigbom.GetK8sVersionFromTkrBoM(tkrBomConfig)
	if err != nil {
		return errors.Wrap(err, "unable to get default kubernetes version")
	}

	defaultBomConfig, err := t.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return errors.Wrap(err, "unable to get default bom configuration")
	}

	// if --yes is set, kick off the upgrade process without waiting for confirmation
	if !options.SkipPrompt {
		if err := askForConfirmation(fmt.Sprintf("Upgrading management cluster '%s' to TKG version '%s' with Kubernetes version '%s'. Are you sure?",
			options.ClusterName, defaultBomConfig.Release.Version, kubernetesVersion)); err != nil {
			return err
		}
	}

	upgradeClusterOption := client.UpgradeClusterOptions{
		KubernetesVersion:   kubernetesVersion,
		TkrVersion:          tkrBomConfig.Release.Version,
		Kubeconfig:          t.kubeconfig,
		ClusterName:         options.ClusterName,
		IsRegionalCluster:   true,
		VSphereTemplateName: options.VSphereTemplateName,
		OSName:              options.OSName,
		OSVersion:           options.OSVersion,
		OSArch:              options.OSArch,
		SkipPrompt:          options.SkipPrompt,
		Edition:             options.Edition,
	}
	err = t.tkgClient.UpgradeManagementCluster(&upgradeClusterOption)
	if err != nil {
		return err
	}

	log.Infof("Management cluster '%s' successfully upgraded to TKG version '%s' with kubernetes version '%s'\n", options.ClusterName, defaultBomConfig.Release.Version, upgradeClusterOption.KubernetesVersion)
	return nil
}
