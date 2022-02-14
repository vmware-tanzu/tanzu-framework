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

// ClusterBootstrapControllerConfig contains the configuration for clusterbootstrap_controller to reconcile resources
type ClusterBootstrapControllerConfig struct {
	// ServiceAccount name that will be used by kapp-controller to install underlying package contents
	PkgiServiceAccount string
	// The name that will be used to create ClusterRole contains all required rules for PkgiServiceAccount
	PkgiClusterRole string
	// The name of ClusterRoleBinding that will be used to bind PkgiClusterRole and PkgiServiceAccount
	PkgiClusterRoleBinding string
	// The namespace where the bootstrap objects will be created, i.e., tkg-system
	BootstrapSystemNamespace string
}
