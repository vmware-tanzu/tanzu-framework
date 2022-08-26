// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
)

var (
	// IsContextAwareDiscoveryEnabled defines default to use when the user has not configured a value
	// This variable is configured at the build time of the CLI
	IsContextAwareDiscoveryEnabled = ""
)

// This block is for global feature constants, to allow them to be used more broadly
const (
	// FeatureContextAwareCLIForPlugins determines whether to use legacy way of discovering plugins or
	// to use the new context-aware Plugin API based plugin discovery mechanism
	// Users can set this featureflag so that we can have context-aware plugin discovery be opt-in for now.
	FeatureContextAwareCLIForPlugins = "features.global.context-aware-cli-for-plugins"
	// FeatureContextCommand determines whether to surface the context command. This is disabled by default.
	FeatureContextCommand = "features.global.context-target"
	// DualStack feature flags determine whether it is permitted to create
	// clusters with a dualstack TKG_IP_FAMILY.  There are separate flags for
	// each primary, "ipv4,ipv6" vs "ipv6,ipv4", and flags for management vs
	// workload cluster plugins.
	FeatureFlagManagementClusterDualStackIPv4Primary = "features.management-cluster.dual-stack-ipv4-primary"
	FeatureFlagManagementClusterDualStackIPv6Primary = "features.management-cluster.dual-stack-ipv6-primary"
	FeatureFlagClusterDualStackIPv4Primary           = "features.cluster.dual-stack-ipv4-primary"
	FeatureFlagClusterDualStackIPv6Primary           = "features.cluster.dual-stack-ipv6-primary"
	// Custom Nameserver feature flags determine whether it is permitted to
	// provide the CONTROL_PLANE_NODE_NAMESERVERS and WORKER_NODE_NAMESERVERS
	// when creating a cluster.
	FeatureFlagManagementClusterCustomNameservers = "features.management-cluster.custom-nameservers"
	FeatureFlagClusterCustomNameservers           = "features.cluster.custom-nameservers"
	// AWS Instance Types Exclude ARM feature flags determine whether instance types with processor architecture
	// support of ARM should be included when discovering available AWS instance types. Setting feature flag to true
	// filters out ARM supporting instance types; false allows ARM instance types to be included in results.
	FeatureFlagAwsInstanceTypesExcludeArm = "features.management-cluster.aws-instance-types-exclude-arm"
	// PackageBasedLCM feature flag determines whether to use package based lifecycle management of management component
	// or legacy way of managing management components. This is also used for clusterclass based management and workload
	// cluster provisioning
	FeatureFlagPackageBasedLCM = "features.global.package-based-lcm-beta"
	// TKR version v1alpha3 feature flag determines whether to use Tanzu Kubernetes Release API version v1alpha3. Setting
	// feature flag to true will allow to use the TKR version v1alpha3; false allows to use legacy TKR version v1alpha1
	FeatureFlagTKRVersionV1Alpha3 = "features.global.tkr-version-v1alpha3-beta"
	// Package Plugin Kctrl Command Tree determines whether to use the command tree from kctrl. Setting feature flag to
	// true will allow to use the package command tree from kctrl for package plugin
	FeatureFlagPackagePluginKctrlCommandTree = "features.package.kctrl-package-command-tree"
	// FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration feature flag determines whether to auto-apply the generated ClusterClass
	// based configuration after converting legacy configration to ClusterClass based config or not
	// Note: This is a hidden feature-flag that doesn't get persisted to config.yaml by default
	FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration = "features.cluster.auto-apply-generated-clusterclass-based-configuration"
	// FeatureFlagForceDeployClusterWithClusterClass if this feature flag is set CLI will try to deploy ClusterClass
	// based cluster even if user has done any customization to the provider templates
	// Note: This is a hidden feature-flag that doesn't get persisted to config.yaml by default
	FeatureFlagForceDeployClusterWithClusterClass = "features.cluster.force-deploy-cluster-with-clusterclass"
)

// DefaultCliFeatureFlags is used to populate an initially empty config file with default values for feature flags.
// The keys MUST be in the format "features.<plugin>.<feature>" or initialization
// will fail. Note that "global" is a special value for <plugin> to be used for CLI-wide features.
//
// If a developer expects that their feature will be ready to release, they should create an entry here with a true
// value.
// If a developer has a beta feature they want to expose, but leave turned off by default, they should create
// an entry here with a false value. WE HIGHLY RECOMMEND the use of a SEPARATE flag for beta use; one that ends in "-beta".
// Thus, if you plan to eventually release a feature with a flag named "features.cluster.foo-bar", you should consider
// releasing the beta version with "features.cluster.foo-bar-beta". This will make it much easier when it comes time for
// mainstreaming the feature (with a default true value) under the flag name "features.cluster.foo-bar", as there will be
// no conflict with previous installs (that have a false value for the entry "features.cluster.foo-bar-beta").
var (
	DefaultCliFeatureFlags = map[string]bool{
		FeatureContextAwareCLIForPlugins:                      ContextAwareDiscoveryEnabled(),
		FeatureContextCommand:                                 false,
		"features.management-cluster.import":                  false,
		"features.management-cluster.export-from-confirm":     true,
		"features.management-cluster.standalone-cluster-mode": false,
		FeatureFlagManagementClusterDualStackIPv4Primary:      false,
		FeatureFlagManagementClusterDualStackIPv6Primary:      false,
		FeatureFlagClusterDualStackIPv4Primary:                false,
		FeatureFlagClusterDualStackIPv6Primary:                false,
		FeatureFlagManagementClusterCustomNameservers:         false,
		FeatureFlagClusterCustomNameservers:                   false,
		FeatureFlagAwsInstanceTypesExcludeArm:                 true,
		FeatureFlagTKRVersionV1Alpha3:                         false,
		FeatureFlagPackagePluginKctrlCommandTree:              false,
	}
)

// ContextAwareDiscoveryEnabled returns true if the IsContextAwareDiscoveryEnabled
// is set to true during build time
func ContextAwareDiscoveryEnabled() bool {
	return strings.EqualFold(IsContextAwareDiscoveryEnabled, "true")
}
