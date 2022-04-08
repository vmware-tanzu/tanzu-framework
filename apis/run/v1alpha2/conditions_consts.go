// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

// Conditions and conditions Reasons for the TanzuKubernetesCluster object

// WaitingForConditionsReason documents TanzuKubernetesCluster Ready Condition is false
// because there are missing conditions from different controllers
const WaitingForConditionsReason = "WaitingForConditions"

const (
	// ControlPlaneReadyCondition should report the control plane nodes are
	// ready and functional for this cluster.
	ControlPlaneReadyCondition clusterv1.ConditionType = "ControlPlaneReady"

	// WaitingForClusterInfrastructureFallbackReason indicates that the cluster is waiting for
	// prerequisites that are necessary for running machines. Examples might include networking, load balancers
	// and so on.
	// NOTE: This reason is used only as a fallback when the InfrastructureCluster
	// is not reporting its own ready condition.
	WaitingForClusterInfrastructureFallbackReason = "WaitingForClusterInfrastructure"

	// WaitingForClusterInfrastructureFallbackMessage indicates that the cluster is waiting for
	// prerequisites that are necessary for running machines. Examples might include networking, load balancers
	// and so on.
	WaitingForClusterInfrastructureFallbackMessage = "Waiting for cluster infrastructure to be ready"

	// WaitingForControlPlaneInitializedReason documents that the first control
	// plane node is initializing.
	WaitingForControlPlaneInitializedReason = "WaitingForControlPlaneInitialized"

	// WaitingForControlPlaneFallbackReason reflects the condition of KubeadmControlPlane.
	// NOTE: This reason is used only as a fallback when the KubeadmControlPlane
	// is not reporting its own ready condition.
	WaitingForControlPlaneFallbackReason = "WaitingForControlPlane"

	// WaitingForControlPlaneFallbackMessage sets default message for WaitingForControlPlaneReason.
	WaitingForControlPlaneFallbackMessage = "Waiting for control planes to be ready"

	// NodesHealthyCondition documents the status of TanzuKubernetesCluster Nodes
	NodesHealthyCondition clusterv1.ConditionType = "NodesHealthy"

	// WaitingForNodesHealthy documents that not all the Nodes are healthy.
	WaitingForNodesHealthy = "WaitingForNodesHealthy"

	// NodePoolsReadyCondition should report the worker nodes are ready and functional for this cluster.
	NodePoolsReadyCondition clusterv1.ConditionType = "NodePoolsReady"

	// NodePoolsReadyConditionUnknownReason indicates the worker nodes condition is unknown.
	NodePoolsReadyConditionUnknownReason = "NodePoolsUnknown"
	// NodePoolsReadyConditionUpdatingReason indicates the worker nodes are updating or scaling.
	NodePoolsReadyConditionUpdatingReason = "NodePoolsUpdating"
	// NodePoolsReadyConditionFailedReason indicates the worker nodes provision failed.
	NodePoolsReadyConditionFailedReason = "NodePoolsFailed"
)

// Conditions for add-ons
const (
	// The condition is set for TanzuKubernetesCluster

	// AddonsReadyCondition is a summary of add-ons(CoreDNS, KubeProxy, CSP, CPI, CNI, AuthSvc) conditions
	AddonsReadyCondition clusterv1.ConditionType = "AddonsReady"

	// AddonsReconciliationFailedReason is a summarized reason for all add-ons reconciliation failure
	AddonsReconciliationFailedReason = "AddonsReconciliationFailed"

	// The below condition is set for AddonStatus

	// ProvisionedCondition documents the status of TanzuKubernetesCluster addon
	ProvisionedCondition clusterv1.ConditionType = "Provisioned"

	// ProvisioningFailedReason (Severity=Warning) documents addon failed to create or update
	ProvisioningFailedReason = "ProvisioningFailed"

	// AddonUnmanagedReason (Severity=Info) documents an addon-on is not managed by controller
	AddonUnmanagedReason = "AddonUnManaged"
)

// Conditions for StorageClass and RoleBinding synchronization, ProviderServiceAccount related resource reconciliation
// and ServiceDiscovery set up
// The conditions are set for TanzuKubernetesCluster
const (
	// StorageClassSyncedCondition documents the status of StorageClass synchronization from supervisor cluster to workload cluster
	StorageClassSyncedCondition clusterv1.ConditionType = "StorageClassSynced"

	// StorageClassSyncFailedReason reports the StorageClass synchronization failed
	StorageClassSyncFailedReason = "StorageClassSyncFailed"

	// RoleBindingSyncedCondition documents the status of RoleBinding synchronization from supervisor cluster to workload cluster
	RoleBindingSyncedCondition clusterv1.ConditionType = "RoleBindingSynced"

	// RoleBindingSyncFailedReason reports the RoleBinding synchronization failed
	RoleBindingSyncFailedReason = "RoleBindingSyncFailed"

	// ProviderServiceAccountsReadyCondition documents the status of provider service accounts
	// and related Roles, RoleBindings and Secrets are created
	ProviderServiceAccountsReadyCondition clusterv1.ConditionType = "ProviderServiceAccountsReady"

	// ProviderServiceAccountsReconciliationFailedReason reports that provider service accounts related resources reconciliation failed
	ProviderServiceAccountsReconciliationFailedReason = "ProviderServiceAccountsReconciliationFailed"

	// ServiceDiscoveryReadyCondition documents the status of service discoveries
	ServiceDiscoveryReadyCondition clusterv1.ConditionType = "ServiceDiscoveryReady"

	// SupervisorHeadlessServiceSetupFailedReason documents the headless service setup for svc api server failed
	SupervisorHeadlessServiceSetupFailedReason = "SupervisorHeadlessServiceSetupFailed"
)

const (
	// ConditionTanzuKubernetesReleaseCompatible mirrors the TanzuKubernetesRelease Compatible condition to the TanzuKubernetesCluster.
	ConditionTanzuKubernetesReleaseCompatible = "TanzuKubernetesReleaseCompatible"
)

// All conditions we want to consider to determine the Ready Condition of TanzuKubernetesCluster
// NOTE: The order of conditions types defines the priority for determining the Reason and Message for the target condition.
var ConditionTypes = []clusterv1.ConditionType{
	ControlPlaneReadyCondition,
	NodePoolsReadyCondition,
	AddonsReadyCondition,
	NodesHealthyCondition,
	ProviderServiceAccountsReadyCondition,
	RoleBindingSyncedCondition,
	StorageClassSyncedCondition,
	ServiceDiscoveryReadyCondition,
}
