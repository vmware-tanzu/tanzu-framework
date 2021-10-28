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
	DefaultCacheDir = filepath.Join(xdg.Home, ".cache", "tanzu")

	// DefaultPluginRoot is the default plugin root.
	DefaultPluginRoot = filepath.Join(xdg.DataHome, "tanzu-cli")

	// DefaultLocalPluginDistroDir is the default Local plugin distribution root directory
	// This directory will be used for local discovery and local distribute of plugins
	DefaultLocalPluginDistroDir = filepath.Join(xdg.Home, ".config", "tanzu-plugins")
)

const (
	// IsContextAwareDiscoveryEnabled defines default to use when the user has not configured a value
	IsContextAwareDiscoveryEnabled = false
)
