// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strings"
)

const (
	// BinNamePrefix is the prefix for tanzu plugin binary names.
	BinNamePrefix = "tanzu-plugin-"

	// TestBinNamePrefix is the prefix for tanzu plugin binary names.
	TestBinNamePrefix = "tanzu-plugin-test-"

	// ArtifactNamePrefix is the prefix for tanzu artifact names.
	ArtifactNamePrefix = "tanzu"
)

// PluginNameFromBin returns a plugin name from the binary name.
func PluginNameFromBin(binName string) string {
	if BuildArch().IsWindows() {
		binName = strings.TrimSuffix(binName, ".exe")
	}
	return strings.TrimPrefix(binName, BinNamePrefix)
}

// BinFromPluginName return a plugin binary name from its name.
func BinFromPluginName(name string) string {
	return BinNamePrefix + name
}

// PluginNameFromTestBin returns a plugin name from the test binary name.
func PluginNameFromTestBin(binName string) string {
	return strings.TrimPrefix(binName, TestBinNamePrefix)
}

// BinTestFromPluginName return a plugin binary name from its name.
func BinTestFromPluginName(name string) string {
	return TestBinNamePrefix + name
}

// MakeArtifactName returns an artifact name for a plugin name.
func MakeArtifactName(pluginName string, arch Arch) string {
	if arch.IsWindows() {
		return fmt.Sprintf("%s-%s-%s.exe", ArtifactNamePrefix, pluginName, arch)
	}
	return fmt.Sprintf("%s-%s-%s", ArtifactNamePrefix, pluginName, arch)
}

// MakeTestArtifactName returns a test artifact name for a plugin name.
func MakeTestArtifactName(pluginName string, arch Arch) string {
	if arch.IsWindows() {
		return fmt.Sprintf("%s-%s-test-%s.exe", ArtifactNamePrefix, pluginName, arch)
	}
	return fmt.Sprintf("%s-%s-test-%s", ArtifactNamePrefix, pluginName, arch)
}
