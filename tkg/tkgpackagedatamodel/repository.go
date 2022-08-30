// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

import "time"

// RepositoryOptions includes fields for repository operations
type RepositoryOptions struct {
	Namespace        string
	RepositoryName   string
	RepositoryURL    string
	PollInterval     time.Duration
	PollTimeout      time.Duration
	AllNamespaces    bool
	CreateRepository bool
	CreateNamespace  bool
	IsForceDelete    bool
	SkipPrompt       bool
	Wait             bool
}

// NewRepositoryOptions instantiates RepositoryOptions
func NewRepositoryOptions() *RepositoryOptions {
	return &RepositoryOptions{}
}
