// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package client provides access to the providers template data
package client

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/client/manifest"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providerinterface"
)

type provider struct {
}

// New returns provider client which implements provider interface
func New() providerinterface.ProviderInterface {
	return &provider{}
}

func (p *provider) GetProviderBundle() ([]byte, error) {
	return manifest.ProviderZipBundle, nil
}
