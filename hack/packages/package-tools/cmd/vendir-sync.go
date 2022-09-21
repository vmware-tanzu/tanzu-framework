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
	"golang.org/x/sync/errgroup"

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
	pkgRepos, err := filterPackageRepos(packageValues)
	if err != nil {
		return err
	}

	packagesPath := filepath.Join(projectRootDir, "packages")
	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)
	files, err := os.ReadDir(packagesPath)
	if err != nil {
		return fmt.Errorf("couldn't read packages directory: %w", err)
	}

	var g errgroup.Group

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		for _, repo := range pkgRepos {
			if packagesContains(packageValues.Repositories[repo].Packages, file.Name()) {
				packagePath := filepath.Join(packagesPath, file.Name())

				// Skip if vendir.yml doesn't exist in package.
				if _, err := os.Stat(filepath.Join(packagePath, "vendir.yml")); err != nil {
					if os.IsNotExist(err) {
						fmt.Printf("No vendir.yml found in package %q. Skipping vendir sync...\n", file.Name())
						break
					} else {
						return err
					}
				}

				fmt.Printf("Syncing package %q\n", file.Name())
				g.Go(func() error {
					return syncPackage(packagePath, toolsBinDir)
				})
			}
		}
	}
	return g.Wait()
}

func syncPackage(packagePath, toolsBinDir string) error {
	cmd := exec.Command(filepath.Join(toolsBinDir, "vendir"), "sync", "--chdir", packagePath) // #nosec G204
	var errBytes bytes.Buffer
	cmd.Stderr = &errBytes
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("couldn't vendir sync package %q: %s", filepath.Base(packagePath), errBytes.String())
	}
	return nil
}
