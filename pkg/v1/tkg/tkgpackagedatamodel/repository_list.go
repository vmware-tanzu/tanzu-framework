// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// RepositoryListOptions includes fields for repository list
type RepositoryListOptions struct {
	KubeConfig    string
	Namespace     string
	AllNamespaces bool
}

// NewRepositoryListOptions instantiates RepositoryListOptions
func NewRepositoryListOptions() *RepositoryListOptions {
	return &RepositoryListOptions{}
}
