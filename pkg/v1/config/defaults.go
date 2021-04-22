// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configv1alpha1 "github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
)

// CoreRepositoryName is the core repository name.
const CoreRepositoryName = "core"

// CoreGCPBucketRepository is the default GCP bucket repository.
var CoreGCPBucketRepository = configv1alpha1.GCPPluginRepository{
	BucketName: "tanzu-cli",
	Name:       CoreRepositoryName,
}

// AdvancedRepositoryName is the advanced repository name.
const AdvancedRepositoryName = "advanced"

// AdvancedGCPBucketRepository is the GCP bucket repository for advanced plugins.
var AdvancedGCPBucketRepository = configv1alpha1.GCPPluginRepository{
	BucketName: "tanzu-cli-advanced-plugins",
	Name:       AdvancedRepositoryName,
}

// TKGRepositoryName is the TKG repository name.
const TKGRepositoryName = "tkg"

// TKGGCPBucketRepository is the GCP bucket repository for TKG plugins.
var TKGGCPBucketRepository = configv1alpha1.GCPPluginRepository{
	BucketName: "tanzu-cli-tkg-plugins",
	Name:       TKGRepositoryName,
}

// DefaultRepositories are the default repositories for the CLI.
var DefaultRepositories []configv1alpha1.PluginRepository = []configv1alpha1.PluginRepository{
	{
		GCPPluginRepository: &CoreGCPBucketRepository,
	},
	{
		GCPPluginRepository: &AdvancedGCPBucketRepository,
	},
	{
		GCPPluginRepository: &TKGGCPBucketRepository,
	},
}
