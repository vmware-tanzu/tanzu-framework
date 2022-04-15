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

// filterPackageRepos returns a list of repos that should be generated.
func filterPackageRepos(pkgVals PackageValues) ([]string, error) {
	var filteredRepos []string

	for repo := range pkgVals.Repositories {
		if packageRepository == "" {
			// --repository flag was not provided and is optional, so don't
			// filter out any repos.
			filteredRepos = append(filteredRepos, repo)
			continue
		}

		_, found := pkgVals.Repositories[packageRepository]
		if !found {
			return nil, fmt.Errorf("%s repository not found", packageRepository)
		}

		if packageRepository == repo {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos, nil
}

// packagesContains checks if a package is in the given collection of packages.
func packagesContains(packagesList []Package, pkg string) bool {
	for _, p := range packagesList {
		if p.Name == pkg {
			return true
		}
	}
	return false
}

// getPackageFromPackageValues returns the Package definition from package-values.yaml file
func getPackageFromPackageValues(projectRootDir, packageName string) (Package, error) {
	packageValues, err := readPackageValues(projectRootDir)
	if err != nil {
		return Package{}, err
	}

	for _, repository := range packageValues.Repositories {
		packages := repository.Packages
		for _, pkg := range packages {
			if pkg.Name == packageName {
				return pkg, nil
			}
		}
	}
	return Package{}, nil
}
