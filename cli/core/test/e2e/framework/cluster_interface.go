// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

// ClusterOps has helper operations to perform on cluster
type ClusterOps interface {
	CreateCluster(name string, args []string) (output string, err error)
	DeleteCluster(name string, args []string) (output string, err error)
	ClusterStatus(name string, args []string) (output string, err error)
}
