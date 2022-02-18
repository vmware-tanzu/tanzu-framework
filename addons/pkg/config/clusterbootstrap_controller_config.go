// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

// ClusterBootstrapControllerConfig contains configuration information related to ClusterBootstrap
type ClusterBootstrapControllerConfig struct {
	CNISelectionClusterVariableName string
	HTTPProxyClusterClassVarName    string
	HTTPSProxyClusterClassVarName   string
	NoProxyClusterClassVarName      string
	ProxyCACertClusterClassVarName  string
	IPFamilyClusterClassVarName     string
}
