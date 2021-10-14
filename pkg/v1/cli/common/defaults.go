// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package common defines generic constants and structs
package common

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	// DefaultCacheDir is the default cache directory
	DefaultCacheDir = filepath.Join(xdg.CacheHome, "tanzu")

	// DefaultPluginRoot is the default plugin root.
	DefaultPluginRoot = filepath.Join(xdg.DataHome, "tanzu-cli")

	// DefaultDistro is the core set of plugins that should be included with the CLI.
	DefaultDistro = []string{"login", "pinniped-auth", "cluster", "management-cluster", "kubernetes-release"}
)
