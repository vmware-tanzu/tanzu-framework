// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"time"
)

// default value constants
const (
	DefaultCNIType = "antrea"

	DefaultDevControlPlaneMachineCount              = 1
	DefaultProdControlPlaneMachineCount             = 3
	DefaultWorkerMachineCountForManagementCluster   = 1
	DefaultDevWorkerMachineCountForWorkloadCluster  = 1
	DefaultProdWorkerMachineCountForWorkloadCluster = 3

	DefaultOperationTimeout            = 30 * time.Second
	DefaultLongRunningOperationTimeout = 30 * time.Minute

	DefaultCertmanagerDeploymentTimeout = 40 * time.Minute

	DefaultNamespace = "default"
)
