// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"time"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// UpdateCredentialsClusterOptions options that can be passed while updating cluster credentials
type UpdateCredentialsClusterOptions struct {
	ClusterName     string
	Namespace       string
	VSphereUsername string
	VSpherePassword string
	Timeout         time.Duration
}

// UpdateCredentialsCluster updates credentials used to access a cluster
func (t *tkgctl) UpdateCredentialsCluster(options UpdateCredentialsClusterOptions) error {
	defer t.restoreAfterSettingTimeout(options.Timeout)()

	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	updateCredentialsOptions := &client.UpdateCredentialsOptions{
		ClusterName:       options.ClusterName,
		Namespace:         options.Namespace,
		IsRegionalCluster: false,
		VSphereUpdateClusterOptions: &client.VSphereUpdateClusterOptions{
			Username: options.VSphereUsername,
			Password: options.VSpherePassword,
		},
	}

	return t.tkgClient.UpdateCredentialsCluster(updateCredentialsOptions)
}
