// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

const (
	// LocalRegistryURL is the url of the local docker registry
	LocalRegistryURL = "localhost:5001"

	// ToolsBinDirPath is the tools bin directory path
	ToolsBinDirPath = "hack/tools/bin"

	// KbldConfigFilePath is the path to kbld-config.yaml file
	KbldConfigFilePath = "packages/kbld-config.yaml"

	// PackageBundlesDir is the path to generated package bundles
	PackageBundlesDir = "build/package-bundles"

	// PackageValuesFilePath is the path to package-values.yaml file
	PackageValuesFilePath = "packages/package-values.yaml"

	// PackageValuesSha256FilePath is the path to package-values-sha256.yaml file
	PackageValuesSha256FilePath = "packages/package-values-sha256.yaml"

	// RepoBundlesDir is the path to the generated repo bundles
	RepoBundlesDir = "build/package-repo-bundles"
)
