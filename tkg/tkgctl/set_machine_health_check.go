// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"strings"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// SetMachineHealthCheckOptions options that can be passed while setting machine healthcheck of a cluster
type SetMachineHealthCheckOptions struct {
	ClusterName            string
	MachineHealthCheckName string
	Namespace              string
	MatchLabels            string
	UnhealthyConditions    string
	NodeStartupTimeout     string
}

// SetMachineHealthCheck apply machine health check to the cluster
//
//nolint:gocritic
func (t *tkgctl) SetMachineHealthCheck(options SetMachineHealthCheckOptions) error {
	optionsSMHC := client.SetMachineHealthCheckOptions{
		ClusterName:            options.ClusterName,
		MachineHealthCheckName: options.MachineHealthCheckName,
		Namespace:              options.Namespace,
		NodeStartupTimeout:     options.NodeStartupTimeout,
	}

	if options.MatchLabels != "" {
		optionsSMHC.MatchLables = strings.Split(options.MatchLabels, ",")
	}

	if options.UnhealthyConditions != "" {
		optionsSMHC.UnhealthyConditions = strings.Split(options.UnhealthyConditions, ",")
	}

	err := t.tkgClient.SetMachineHealthCheck(&optionsSMHC)
	if err != nil {
		return err
	}

	log.Info("The MachineHealthCheck was set successfully")

	return nil
}
