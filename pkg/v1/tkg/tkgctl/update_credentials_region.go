// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"time"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
)

// UpdateCredentialsRegionOptions options that can passed while updating credentials of a management-cluster
type UpdateCredentialsRegionOptions struct {
	ClusterName     string
	VSphereUsername string
	VSpherePassword string
	IsCascading     bool
	Timeout         time.Duration
}

// UpdateCredentialsRegion updates credentials used to login to a management-cluster
func (t *tkgctl) UpdateCredentialsRegion(options UpdateCredentialsRegionOptions) error {
	defer t.restoreAfterSettingTimeout(options.Timeout)()

	updateCredentialsOptions := &client.UpdateCredentialsOptions{
		ClusterName:       options.ClusterName,
		Namespace:         "tkg-system",
		IsRegionalCluster: true,
		VSphereUpdateClusterOptions: &client.VSphereUpdateClusterOptions{
			Username: options.VSphereUsername,
			Password: options.VSpherePassword,
		},
		IsCascading: options.IsCascading,
	}

	return t.tkgClient.UpdateCredentialsRegion(updateCredentialsOptions)
}
