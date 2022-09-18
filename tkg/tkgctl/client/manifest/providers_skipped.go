// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !embedproviders
// +build !embedproviders

// Package manifest ...
package manifest

// ProviderZipBundle variable defined when embedproviders tag is not passed during build
// Meaning provider will be downloaded based on TKG BoM file
var ProviderZipBundle []byte
