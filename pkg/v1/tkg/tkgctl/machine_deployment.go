// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
)

func (c *tkgctl) GetMachineDeployments(options client.GetMachineDeploymentOptions) ([]capi.MachineDeployment, error) {
	return c.tkgClient.GetMachineDeployments(options)
}

// GetPacificMachineDeployments returns the MachineDeployments of Pacific(TKGS) TKC cluster
// This is defined separately for Pacific (TKGS) provider because the TKGS and TKGm CAPI versions could be different
// and this should be deprecated after clusterclass is adopted by both TKGm and TKGS
func (c *tkgctl) GetPacificMachineDeployments(options client.GetMachineDeploymentOptions) ([]capi.MachineDeployment, error) {
	return c.tkgClient.GetPacificMachineDeployments(options)
}

func (c *tkgctl) SetMachineDeployment(options *client.SetMachineDeploymentOptions) error {
	if err := c.tkgClient.SetMachineDeployment(options); err != nil {
		return err
	}
	return nil
}

func (c *tkgctl) DeleteMachineDeployment(options client.DeleteMachineDeploymentOptions) error {
	if err := c.tkgClient.DeleteMachineDeployment(options); err != nil {
		return err
	}
	return nil
}
