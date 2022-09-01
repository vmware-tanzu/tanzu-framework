// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build embedproviders
// +build embedproviders

// Package manifest...
package manifest

import _ "embed"

//go:embed providers.zip
var ProviderZipBundle []byte
