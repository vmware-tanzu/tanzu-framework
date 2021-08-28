// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"time"
)

// default value constants
const (
	DefaultCNIType = "antrea"

	DefaultDevControlPlaneMachineCount  = 1
	DefaultProdControlPlaneMachineCount = 3
	DefaultDevWorkerMachineCount        = 1
	DefaultProdWorkerMachineCount       = 3

	DefaultOperationTimeout            = 30 * time.Second
	DefaultLongRunningOperationTimeout = 30 * time.Minute

	DefaultCertmanagerDeploymentTimeout = 40 * time.Minute

	DefaultNamespace = "default"

	// de-facto defaults initially chosen by kops: https://github.com/kubernetes/kops
	DefaultIPv4ClusterCIDR = "100.96.0.0/11"
	DefaultIPv4ServiceCIDR = "100.64.0.0/13"

	// chosen to match our IPv4 defaults
	// use /48 for cluster CIDR because each node gets a /64 by default in IPv6
	DefaultIPv6ClusterCIDR = "fd00:100:96::/48"
	// use /108 is the max allowed for IPv6
	DefaultIPv6ServiceCIDR = "fd00:100:64::/108"

	// dual stack IPv4,IPv6 defaults
	DefaultDualStackPrimaryIPv4ClusterCIDR = DefaultIPv4ClusterCIDR + "," + DefaultIPv6ClusterCIDR
	DefaultDualStackPrimaryIPv4ServiceCIDR = DefaultIPv4ServiceCIDR + "," + DefaultIPv6ServiceCIDR

	DefaultDualStackPrimaryIPv6ClusterCIDR = DefaultIPv6ClusterCIDR + "," + DefaultIPv4ClusterCIDR
	DefaultDualStackPrimaryIPv6ServiceCIDR = DefaultIPv6ServiceCIDR + "," + DefaultIPv4ServiceCIDR
	// DefaultIsWindowsWorkloadCluster is false, indicating that the normal thing to do is, is to make linux clusters.
	DefaultIsWindowsWorkloadCluster = false
)
