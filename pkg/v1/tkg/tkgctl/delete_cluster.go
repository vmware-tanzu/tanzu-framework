// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// DeleteClustersOptions delete cluster options
type DeleteClustersOptions struct {
	ClusterName string
	Namespace   string
	SkipPrompt  bool
}

// DeleteCluster deletes workload cluster
func (t *tkgctl) DeleteCluster(options DeleteClustersOptions) error {
	// Make sure activity is captured in the audit log in case of deletion failure.
	if logPath, err := t.getAuditLogPath(options.ClusterName); err == nil {
		log.SetAuditLog(logPath)
	}

	// if --yes is set, kick off the delete process without waiting for confirmation
	if !options.SkipPrompt {
		if err := askForConfirmation(fmt.Sprintf("Deleting workload cluster '%s'. Are you sure?", options.ClusterName)); err != nil {
			return err
		}
	}

	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	deleteWcOptions := client.DeleteWorkloadClusterOptions{
		ClusterName: options.ClusterName,
		Namespace:   options.Namespace,
	}

	err := t.tkgClient.DeleteWorkloadCluster(deleteWcOptions)
	if err != nil {
		return err
	}

	log.Infof("Workload cluster '%s' is being deleted \n", options.ClusterName)

	// Clean up the audit log since we were successful
	t.removeAuditLog(options.ClusterName)
	log.SetAuditLog("")

	return nil
}
