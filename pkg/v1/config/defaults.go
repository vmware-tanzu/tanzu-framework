// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// CoreRepositoryName is the core repository name.
const CoreRepositoryName = "core"

// CoreBucketName is the name of the core plugin repository bucket to use.
var CoreBucketName = "tanzu-cli-framework"

// DefaultVersionSelector is to only use stable versions of plugins
const DefaultVersionSelector = configv1alpha1.NoUnstableVersions

// CoreGCPBucketRepository is the default GCP bucket repository.
var CoreGCPBucketRepository = configv1alpha1.GCPPluginRepository{
	BucketName: CoreBucketName,
	Name:       CoreRepositoryName,
}

// AdvancedRepositoryName is the advanced repository name.
const AdvancedRepositoryName = "advanced"

// AdvancedGCPBucketRepository is the GCP bucket repository for advanced plugins.
var AdvancedGCPBucketRepository = configv1alpha1.GCPPluginRepository{
	BucketName: "tanzu-cli-advanced-plugins",
	Name:       AdvancedRepositoryName,
}

// DefaultRepositories are the default repositories for the CLI.
var DefaultRepositories []configv1alpha1.PluginRepository = []configv1alpha1.PluginRepository{
	{
		GCPPluginRepository: &CoreGCPBucketRepository,
	},
}
