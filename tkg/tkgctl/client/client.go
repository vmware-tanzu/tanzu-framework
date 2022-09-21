// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package client provides access to the providers template data
// Copied over from github.com/vmware-tanzu/tanzu-framework/providers/client
package client

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/providerinterface"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl/client/manifest"
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
