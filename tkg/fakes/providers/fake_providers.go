// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package providers ...
package providers

import (
	_ "embed"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providerinterface"
)

//go:embed providers.zip
var FakeProviderZip []byte

type fakeProviderGetter struct {
}

func (f *fakeProviderGetter) GetProviderBundle() ([]byte, error) {
	return FakeProviderZip, nil
}

func FakeProviderGetter() providerinterface.ProviderInterface {
	return &fakeProviderGetter{}
}
