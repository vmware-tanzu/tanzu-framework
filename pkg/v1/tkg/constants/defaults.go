/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
