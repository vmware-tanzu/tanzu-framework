// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

// SetRegionOptions options for set active management cluster
type SetRegionOptions struct {
	ClusterName string
	ContextName string
}

// SetRegion sets active management cluster
func (t *tkgctl) SetRegion(options SetRegionOptions) error {
	err := t.tkgClient.SetRegionContext(options.ClusterName, options.ContextName)
	if err != nil {
		return err
	}
	log.Infof("The current management cluster context is switched to '%s'", options.ClusterName)
	return nil
}
