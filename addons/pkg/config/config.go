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
	CNISelectionClusterVariableName string
	HTTPProxyClusterClassVarName    string
	HTTPSProxyClusterClassVarName   string
	NoProxyClusterClassVarName      string
	ProxyCACertClusterClassVarName  string
	IPFamilyClusterClassVarName     string
}
