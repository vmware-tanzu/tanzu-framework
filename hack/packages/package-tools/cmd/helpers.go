// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/constants"
)

func readPackageValues(projectRootDir string) (PackageValues, error) {
	var packageValues PackageValues

	packageValuesData, err := os.ReadFile(filepath.Join(projectRootDir, constants.PackageValuesFilePath))
	if err != nil {
		return PackageValues{}, fmt.Errorf("couldn't read file %s: %w", packageValuesFile, err)
	}

	if err := yaml.Unmarshal(packageValuesData, &packageValues); err != nil {
		return PackageValues{}, fmt.Errorf("error while unmarshaling: %w", err)
	}

	return packageValues, nil
}

// getPackageFromPackageValues returns the package definition from the package-values.yaml file.
func getPackageFromPackageValues(projectRootDir, packageName string) (Package, error) {
	packageValues, err := readPackageValues(projectRootDir)
	if err != nil {
		return Package{}, err
	}

	for i := range packageValues.Repositories {
		packages := packageValues.Repositories[i].Packages
		for _, pkg := range packages {
			if pkg.Name == packageName {
				return pkg, nil
			}
		}
	}
	return Package{}, nil
}
