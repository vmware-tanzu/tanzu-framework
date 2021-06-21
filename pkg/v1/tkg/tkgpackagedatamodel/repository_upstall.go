// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// RepositoryOptions includes fields for repository install/update
type RepositoryOptions struct {
	RepositoryName   string
	RepositoryURL    string
	KubeConfig       string
	Namespace        string
	CreateRepository bool
	CreateNamespace  bool
}

// NewRepositoryOptions instantiates RepositoryOptions
func NewRepositoryOptions() *RepositoryOptions {
	return &RepositoryOptions{}
}
