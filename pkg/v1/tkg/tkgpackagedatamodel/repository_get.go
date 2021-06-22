// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// RepositoryGetOptions includes fields for repository get
type RepositoryGetOptions struct {
	RepositoryName string
	KubeConfig     string
	Namespace      string
}

// NewRepositoryGetOptions instantiates RepositoryGetOptions
func NewRepositoryGetOptions() *RepositoryGetOptions {
	return &RepositoryGetOptions{}
}
