// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package providerinterface ...
package providerinterface

// ProviderInterface implements provider interface functions
type ProviderInterface interface {
	// GetProviderBundle should return provider bundle zip file bytes
	GetProviderBundle() ([]byte, error)
}
