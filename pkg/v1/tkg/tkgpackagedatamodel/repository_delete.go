// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// RepositoryDeleteOptions includes fields for repository delete
type RepositoryDeleteOptions struct {
	RepositoryName string
	IsForce        bool
	KubeConfig     string
	Namespace      string
}

// NewRepositoryDeleteOptions instantiates RepositoryDeleteOptions
func NewRepositoryDeleteOptions() *RepositoryDeleteOptions {
	return &RepositoryDeleteOptions{}
}
