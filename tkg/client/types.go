// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

// TKGClusterPhase is used to type the constants describing possible cluster status.
type TKGClusterPhase string

const (
	// TKGClusterPhaseCreating means the cluster control plane is under creation,
	// or that infrastructure has been created and configured but not yet with
	// an initialized control plane.
	TKGClusterPhaseCreating = TKGClusterPhase("creating")

	// TKGClusterPhaseCreationStalled indicates that the cluster is in the
	// process of creating control plane and the process is possibly
	// stalled and user intervention is required.
	TKGClusterPhaseCreationStalled = TKGClusterPhase("createStalled")

	// TKGClusterPhaseRunning means that the infrastructure has been created
	// and configured, and that the control plane has been fully initialized.
	TKGClusterPhaseRunning = TKGClusterPhase("running")

	// TKGClusterPhaseDeleting means that the cluster is being deleted.
	TKGClusterPhaseDeleting = TKGClusterPhase("deleting")

	// TKGClusterPhaseFailed means that cluster control plane
	// creation failed. Possible user intervention required.
	// This Phase to be used for TKC(pacific cluster).
	TKGClusterPhaseFailed = TKGClusterPhase("failed")

	// TKGClusterPhaseUpdating indicates that the cluster is in the
	// process of rolling out an update or scaling nodes
	TKGClusterPhaseUpdating = TKGClusterPhase("updating")

	// TKGClusterPhaseUpdateFailed indicates that the cluster's
	// rolling update failed and likely requires user intervention.
	// This Phase to be used for TKC(pacific cluster).
	TKGClusterPhaseUpdateFailed = TKGClusterPhase("updateFailed")

	// TKGClusterPhaseUpdateStalled indicates that the cluster is in the
	// process of rolling out an update and the update process is possibly
	// stalled and user intervention is required.
	TKGClusterPhaseUpdateStalled = TKGClusterPhase("updateStalled")

	// TKGClusterPhaseEmpty is useful for the initial reconcile,
	// before we even state the phase as creating.
	TKGClusterPhaseEmpty = TKGClusterPhase("")
)

// TKGClusterType specifies cluster type
type TKGClusterType string

// cluster types
const (
	ManagementCluster TKGClusterType = "Management"
	WorkloadCluster   TKGClusterType = "Workload"
)
