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

// DeleteRegionOptions delete region options
type DeleteRegionOptions struct {
	ClusterName        string
	Force              bool
	UseExistingCluster bool
	SkipPrompt         bool
	Timeout            time.Duration
}

// DeleteRegion deletes management cluster
func (t *tkgctl) DeleteRegion(options DeleteRegionOptions) error {
	// Make sure activity is captured in the audit log in case of deletion failure.
	if logPath, err := t.getAuditLogPath(options.ClusterName); err == nil {
		log.SetAuditLog(logPath)
	}

	var err error

	// delete region requires minimum 15 minutes timeout
	minTimeoutReq := 15 * time.Minute
	if options.Timeout < minTimeoutReq {
		log.V(6).Infof("timeout duration of at least 15 minutes is required, using default timeout %v", constants.DefaultLongRunningOperationTimeout)
		options.Timeout = constants.DefaultLongRunningOperationTimeout
	}

	defer t.restoreAfterSettingTimeout(options.Timeout)()

	// if --yes is set, kick off the delete process without waiting for confirmation
	if !options.SkipPrompt {
		err = askForConfirmation(fmt.Sprintf("Deleting management cluster '%s'. Are you sure?", options.ClusterName))
		if err != nil {
			return err
		}
	}

	log.V(1).Infof("\nDeleting management cluster...\n")

	optionsDR := client.DeleteRegionOptions{
		Force:              options.Force,
		Kubeconfig:         t.kubeconfig,
		UseExistingCluster: options.UseExistingCluster,
		ClusterName:        options.ClusterName,
	}

	err = t.tkgClient.DeleteRegion(optionsDR)
	if err != nil {
		return errors.Wrap(err, "unable to delete management cluster")
	}

	log.Infof("\nManagement cluster deleted!\n")

	// Clean up the audit log since we were successful
	t.removeAuditLog(options.ClusterName)
	log.SetAuditLog("")

	return nil
}
