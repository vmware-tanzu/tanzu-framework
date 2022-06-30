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
		return PackageValues{}, fmt.Errorf("reading %s: %w", constants.PackageValuesFilePath, err)
	}

	if err := yaml.Unmarshal(packageValuesData, &packageValues); err != nil {
		return PackageValues{}, fmt.Errorf("unmarshalling package values data: %w", err)
	}

	return packageValues, nil
}

// getPackageFromPackageValues returns the list of package definition from the package-values.yaml file.
func getPackageFromPackageValues(projectRootDir, packageName string) ([]Package, error) {
	packageValues, err := readPackageValues(projectRootDir)
	if err != nil {
		return []Package{}, err
	}

	matchingPackages := []Package{}
	for i := range packageValues.Repositories {
		packages := packageValues.Repositories[i].Packages
		for i := range packages {
			if packages[i].Name == packageName {
				if packages[i].Version == version {
					return []Package{packages[i]}, nil
				}
				matchingPackages = append(matchingPackages, packages[i])
			}
		}
	}

	// Found packages but no matching version found, returning all matching package
	if len(matchingPackages) >= 1 {
		return matchingPackages, nil
	}

	return []Package{}, fmt.Errorf("package %q not found in %s", packageName, constants.PackageValuesFilePath)
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
	for i := range packagesList {
		if packagesList[i].Name == pkg {
			return true
		}
	}
	return false
}

type formattedVersion struct {
	concatenator             string
	version, noV, subVersion string
	concat, concatNoV        string
}

func formatVersion(pkg *Package, concatenator string) formattedVersion {
	if pkg != nil && pkg.SkipVersionOverride {
		fv := formattedVersion{
			version:      pkg.Version,
			noV:          getPackageVersion(pkg.Version),
			concatenator: concatenator,
			subVersion:   pkg.PackageSubVersion,
		}

		fv.concat = fv.version
		fv.concatNoV = fv.noV
		if fv.subVersion != "" {
			fv.concat = fv.version + concatenator + fv.subVersion
			fv.concatNoV = fv.noV + concatenator + fv.subVersion
		}
		return fv
	}

	fv := formattedVersion{
		version:      version,
		noV:          getPackageVersion(version),
		concatenator: concatenator,
	}

	if pkg != nil {
		fv.subVersion = pkg.PackageSubVersion
	}
	if subVersion != "" {
		// subVersion flag overrides package values subversion.
		fv.subVersion = subVersion
	}

	fv.concat = fv.version
	fv.concatNoV = fv.noV
	if fv.subVersion != "" {
		fv.concat = fv.version + concatenator + fv.subVersion
		fv.concatNoV = fv.noV + concatenator + fv.subVersion
	}

	return fv
}

func getPackageVersion(version string) string {
	pkgVersion := version
	if string(version[0]) == "v" {
		pkgVersion = version[1:]
	}
	return pkgVersion
}

func getEnvArrayFromMap(env map[string]string) []string {
	var arrEnv []string
	for k, v := range env {
		arrEnv = append(arrEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return arrEnv
}

func getTempPackageHelpersLib(lib string) (string, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "package-helpers.lib-")
	if err != nil {
		return "", fmt.Errorf("cannot create temporary directory: %w", err)
	}
	tempFileName := filepath.Join(tempDir, "package-helpers.lib.yaml")
	if err := os.WriteFile(tempFileName, []byte(lib), 0755); err != nil {
		return "", err
	}
	return tempFileName, nil
}
