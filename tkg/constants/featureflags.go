// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// This block is for global feature constants, to allow them to be used more broadly
const (

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
	// PackageBasedCC feature flag determines whether to use package based lifecycle management of management component
	// or legacy way of managing management components. This is also used for clusterclass based management cluster provisioning
	FeatureFlagPackageBasedCC = "features.management-cluster.package-based-cc"
	// FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration feature flag determines whether to auto-apply the generated ClusterClass
	// based configuration after converting legacy configration to ClusterClass based config or not
	// Note: This is a hidden feature-flag that doesn't get persisted to config.yaml by default
	FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration = "features.cluster.auto-apply-generated-clusterclass-based-configuration"
	// FeatureFlagForceDeployClusterWithClusterClass if this feature flag is set CLI will try to deploy ClusterClass
	// based cluster even if user has done any customization to the provider templates
	// Note: This is a hidden feature-flag that doesn't get persisted to config.yaml by default
	FeatureFlagForceDeployClusterWithClusterClass = "features.cluster.force-deploy-cluster-with-clusterclass"
	// FeatureFlagSingleNodeClusters is to enable Single Node Cluster deployment via tanzu CLI.
	// Setting the feature flag to true will allow the creation of Single Node Clusters.
	FeatureFlagSingleNodeClusters = "features.cluster.single-node-clusters"
	// FeatureFlagManagementClusterDeployInClusterIPAMProvider feature flag
	// determines whether to apply the In-Cluster IPAM provider to the
	// management cluster.
	FeatureFlagManagementClusterDeployInClusterIPAMProvider = "features.management-cluster.deploy-in-cluster-ipam-provider-beta"
	// FeatureFlagAllowLegacyCluster is used to decide the workload cluster is clusterclass based or legayc based.
	// By default, it's false. If it's true, then workload cluster is legacy based.
	FeatureFlagAllowLegacyCluster = "features.cluster.allow-legacy-cluster"
)
