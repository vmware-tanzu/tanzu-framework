// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/constants"
	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/utils"
)

// packageVendirSyncCmd is for sync the package
var packageVendirSyncCmd = &cobra.Command{
	Use:   "vendir sync",
	Short: "Package vendir sync",
	RunE:  runPackageVendirSync,
}

func init() {
	rootCmd.AddCommand(packageVendirSyncCmd)
	packageVendirSyncCmd.Flags().StringVar(&packageRepository, "repository", "", "Package repository of the packages to be synced")
}

func runPackageVendirSync(cmd *cobra.Command, args []string) error {
	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}

	packageValues, err := readPackageValues(projectRootDir)
	if err != nil {
		return err
	}
	packagesToSync, err := selectPackages(packageValues, packageRepository)
	if err != nil {
		return err
	}

	packagesPath := filepath.Join(projectRootDir, "packages")
	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)
	files, err := os.ReadDir(packagesPath)
	if err != nil {
		return fmt.Errorf("couldn't read packages directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			_, ok := packagesToSync[file.Name()]
			if ok {
				fmt.Printf("Syncing package %s\n", file.Name())
				packagePath := filepath.Join(packagesPath, file.Name())
				if err := syncPackage(packagePath, toolsBinDir); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func syncPackage(packagePath, toolsBinDir string) error {
	err := os.Chdir(packagePath)
	if err != nil {
		return fmt.Errorf("couldn't change to directory %s: %w", packagePath, err)
	}

	cmd := exec.Command(filepath.Join(toolsBinDir, "vendir"), "sync") // #nosec G204
	var errBytes bytes.Buffer
	cmd.Stderr = &errBytes
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("couldn't vendir sync package: %s", errBytes.String())
	}
	return nil
}

// selectPackages will return a map of package names as keys in a given repo. If
// no repo is provided, all packages will be in the map.
func selectPackages(pkgVals PackageValues, repoName string) (map[string]struct{}, error) {
	selectPkgs := make(map[string]struct{})

	for repo := range pkgVals.Repositories {
		if repoName == "" {
			// --repository flag was not provided and is optional, so don't
			// filter out any packages.
			for _, pkg := range pkgVals.Repositories[repo].Packages {
				selectPkgs[pkg.Name] = struct{}{}
			}
			continue
		}

		_, found := pkgVals.Repositories[repoName]
		if !found {
			return nil, fmt.Errorf("%s repository not found", repoName)
		}

		if pkgVals.Repositories[repo].Name == repoName {
			for _, pkg := range pkgVals.Repositories[repo].Packages {
				selectPkgs[pkg.Name] = struct{}{}
			}
		}
	}

	return selectPkgs, nil
}
