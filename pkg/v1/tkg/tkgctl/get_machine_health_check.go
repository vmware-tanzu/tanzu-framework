// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
)

// GetMachineHealthCheckOptions options that can be passed while getting the machine health check of a cluster
type GetMachineHealthCheckOptions struct {
	ClusterName            string
	MachineHealthCheckName string
	Namespace              string
	MatchLabel             string
}

// GetMachineHealthCheck return machinehealthcheck configuration for the cluster
func (t *tkgctl) GetMachineHealthCheck(options GetMachineHealthCheckOptions) ([]client.MachineHealthCheck, error) {
	machineHealthCheckOptions := client.MachineHealthCheckOptions{
		ClusterName:            options.ClusterName,
		Namespace:              options.Namespace,
		MachineHealthCheckName: options.MachineHealthCheckName,
		MatchLabel:             options.MatchLabel,
	}

	return t.tkgClient.GetMachineHealthChecks(machineHealthCheckOptions)
}
