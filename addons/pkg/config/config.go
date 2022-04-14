// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config implements configuration settings information.
package config

import (
	"time"
)

// AddonControllerConfig contains addons controller configuration information.
type AddonControllerConfig struct {
	AppSyncPeriod           time.Duration
	AppWaitTimeout          time.Duration
	AddonNamespace          string
	AddonServiceAccount     string
	AddonClusterRole        string
	AddonClusterRoleBinding string
	AddonImagePullPolicy    string
	CorePackageRepoName     string
}

// ClusterBootstrapControllerConfig contains configuration information related to ClusterBootstrap
type ClusterBootstrapControllerConfig struct {
	IPFamilyClusterClassVarName string
	// The length of time to wait before kapp-controller's reconciliation
	PkgiSyncPeriod time.Duration
	// ServiceAccount name that will be used by kapp-controller to install underlying package contents
	PkgiServiceAccount string
	// The name that will be used to create ClusterRole contains all required rules for PkgiServiceAccount
	PkgiClusterRole string
	// The name of ClusterRoleBinding that will be used to bind PkgiClusterRole and PkgiServiceAccount
	PkgiClusterRoleBinding string
	// The namespace where the bootstrap objects will be created, i.e., tkg-system
	SystemNamespace string
}

// PackageInstallStatusControllerConfig contains configuration information related to PackageInstallStatus
type PackageInstallStatusControllerConfig struct {
	// The namespace where the bootstrap objects will be created, i.e., tkg-system
	SystemNamespace string
}
