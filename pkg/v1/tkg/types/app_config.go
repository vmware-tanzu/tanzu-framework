// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types package to store configs
package types

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providerinterface"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

// AppConfig stores configuration related to running app
type AppConfig struct {
	TKGConfigDir      string
	ProviderGetter    providerinterface.ProviderInterface
	CustomizerOptions CustomizerOptions
	TKGSettingsFile   string
}

// CustomizerOptions provides overrides for CreateAllClients that allows a
// user to customize the underying clients.
type CustomizerOptions struct {
	RegionManagerFactory region.ManagerFactory
}
