// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"time"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// UpdateCredentialsClusterOptions options that can be passed while updating cluster credentials
type UpdateCredentialsClusterOptions struct {
	ClusterName         string
	Namespace           string
	VSphereUsername     string
	VSpherePassword     string
	AzureTenantID       string
	AzureSubscriptionID string
	AzureClientID       string
	AzureClientSecret   string
	Timeout             time.Duration
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
		AzureUpdateClusterOptions: &client.AzureUpdateClusterOptions{
			AzureTenantID:       options.AzureTenantID,
			AzureSubscriptionID: options.AzureSubscriptionID,
			AzureClientID:       options.AzureClientID,
			AzureClientSecret:   options.AzureClientSecret,
		},
	}

	return t.tkgClient.UpdateCredentialsCluster(updateCredentialsOptions)
}
