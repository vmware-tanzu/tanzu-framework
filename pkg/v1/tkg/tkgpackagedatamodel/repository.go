// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// RepositoryOptions includes fields for repository operations
type RepositoryOptions struct {
	KubeConfig       string
	Namespace        string
	RepositoryName   string
	RepositoryURL    string
	AllNamespaces    bool
	CreateRepository bool
	CreateNamespace  bool
	IsForceDelete    bool
}

// NewRepositoryOptions instantiates RepositoryOptions
func NewRepositoryOptions() *RepositoryOptions {
	return &RepositoryOptions{}
}
