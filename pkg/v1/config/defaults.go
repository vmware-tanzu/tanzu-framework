// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"strings"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// Default Standalone Discovery configuration
// Value of this variables gets assigned during build time
var (
	DefaultStandaloneDiscoveryRepository = ""
	DefaultStandaloneDiscoveryImagePath  = ""
	DefaultStandaloneDiscoveryImageTag   = ""
	DefaultStandaloneDiscoveryName       = "default"
	DefaultStandaloneDiscoveryType       = common.DistributionTypeOCI
	DefaultStandaloneDiscoveryLocalPath  = ""
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

// DefaultStandaloneDiscoveryImage returns the default Standalone Discovery image
// from the configured build time variables
func DefaultStandaloneDiscoveryImage() string {
	defaultStandaloneDiscoveryRepository := DefaultStandaloneDiscoveryRepository
	defaultStandaloneDiscoveryImagePath := DefaultStandaloneDiscoveryImagePath
	defaultStandaloneDiscoveryImageTag := DefaultStandaloneDiscoveryImageTag

	// Run-time overrides of the configuration
	if customImageRepo := os.Getenv(constants.ConfigVariableCustomImageRepository); customImageRepo != "" {
		defaultStandaloneDiscoveryRepository = customImageRepo
	}
	if imagePath := os.Getenv(constants.ConfigVariableDefaultStandaloneDiscoveryImagePath); imagePath != "" {
		defaultStandaloneDiscoveryImagePath = imagePath
	}
	if imageTag := os.Getenv(constants.ConfigVariableDefaultStandaloneDiscoveryImageTag); imageTag != "" {
		defaultStandaloneDiscoveryImageTag = imageTag
	}

	return strings.Trim(defaultStandaloneDiscoveryRepository, "/") + "/" + strings.Trim(defaultStandaloneDiscoveryImagePath, "/") + ":" + defaultStandaloneDiscoveryImageTag
}
