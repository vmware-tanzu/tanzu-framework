// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config implements configuration settings information.
package config

import (
	"time"
)

// Config contains configuration information.
type Config struct {
	AppSyncPeriod           time.Duration
	AppWaitTimeout          time.Duration
	AddonNamespace          string
	AddonServiceAccount     string
	AddonClusterRole        string
	AddonClusterRoleBinding string
	AddonImagePullPolicy    string
	CorePackageRepoName     string
}
