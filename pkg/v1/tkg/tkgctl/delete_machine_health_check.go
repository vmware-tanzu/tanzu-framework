// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

// DeleteMachineHealthCheckOptions delete cluster options
type DeleteMachineHealthCheckOptions struct {
	ClusterName            string
	Namespace              string
	MachinehealthCheckName string
	SkipPrompt             bool
}

// DeleteMachineHealthCheck deletes MHC on cluster
func (t *tkgctl) DeleteMachineHealthCheck(options DeleteMachineHealthCheckOptions) error {
	var err error

	if !options.SkipPrompt {
		err = askForConfirmation(fmt.Sprintf("Deleting MachineHealthCheck for cluster %s. Are you sure?", options.ClusterName))
		if err != nil {
			return err
		}
	}

	optionsMHC := client.MachineHealthCheckOptions{
		ClusterName:            options.ClusterName,
		Namespace:              options.Namespace,
		MachineHealthCheckName: options.MachinehealthCheckName,
	}
	err = t.tkgClient.DeleteMachineHealthCheck(optionsMHC)
	if err != nil {
		return err
	}

	log.Info("The MachineHealthCheck was deleted successfully")

	return nil
}
