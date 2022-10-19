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
	// The maximum amount of time that will be spent trying to clean resources before cluster deletion is allowed to proceed.
	ClusterDeleteTimeout time.Duration
}

// PackageInstallStatusControllerConfig contains configuration information related to PackageInstallStatus
type PackageInstallStatusControllerConfig struct {
	// The namespace where the bootstrap objects will be created, i.e., tkg-system
	SystemNamespace string
}

// ConfigControllerConfig contains common configuration information of config controller
type ConfigControllerConfig struct {
	// The namespace where the template config objects will be created, i.e., tkg-system
	SystemNamespace string
}

// AntreaConfigControllerConfig contains configuration information of AntreaConfig controller
type AntreaConfigControllerConfig struct {
	ConfigControllerConfig
}

// CalicoConfigControllerConfig contains configuration information of CalicoConfig controller
type CalicoConfigControllerConfig struct {
	ConfigControllerConfig
}

// KappControllerConfigControllerConfig contains configuration information of KappControllerConfig controller
type KappControllerConfigControllerConfig struct {
	ConfigControllerConfig
}

// VSphereCPIConfigControllerConfig contains configuration information of VSphereCPIConfig controller
type VSphereCPIConfigControllerConfig struct {
	ConfigControllerConfig
}

// VSphereCSIConfigControllerConfig contains configuration information of VSphereCSIConfig controller
type VSphereCSIConfigControllerConfig struct {
	ConfigControllerConfig
}

// AwsEbsCSIConfigControllerConfig contains configuration information of VSphereCSIConfig controller
type AwsEbsCSIConfigControllerConfig struct {
	ConfigControllerConfig
}
