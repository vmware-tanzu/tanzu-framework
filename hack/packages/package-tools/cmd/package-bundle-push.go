// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/constants"
	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/utils"
)

// packageBundlePushCmd is for pushing package bundles
var packageBundlePushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push package bundle",
	RunE:  runPackageBundlePush,
}

func init() {
	packageBundleCmd.AddCommand(packageBundlePushCmd)
	packageBundlePushCmd.Flags().StringVar(&packageRepository, "repository", "", "Package repository of the package bundle being pushed")
	packageBundlePushCmd.Flags().StringVar(&registry, "registry", "", "OCI registry where the package bundle image needs to be stored")
	packageBundlePushCmd.Flags().StringVar(&version, "version", "", "Package bundle version")
	packageBundlePushCmd.Flags().StringVar(&subVersion, "sub-version", "", "Package bundle subversion")
	packageBundlePushCmd.Flags().BoolVar(&all, "all", false, "Push all package bundles in given package repository to an image repository")
	packageBundlePushCmd.MarkFlagRequired("registry") //nolint: errcheck
	packageBundlePushCmd.MarkFlagRequired("version")  //nolint: errcheck
}

func runPackageBundlePush(cmd *cobra.Command, args []string) error {
	if err := validatePackageBundlePushFlags(args); err != nil {
		return err
	}
	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}
	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)

	packageValues, err := readPackageValues(projectRootDir)
	if err != nil {
		return err
	}

	repos, err := filterPackageRepos(packageValues)
	if err != nil {
		return err
	}

	if !all {
		// The first argument is expected to be a comma-separated list of
		// package bundles.
		if err := prunePackages(packageValues.Repositories, args[0]); err != nil {
			return err
		}
	}

	for _, repo := range repos {
		for i := range packageValues.Repositories[repo].Packages {
			pkg := packageValues.Repositories[repo].Packages[i]
			fmt.Printf("Pushing %q package bundle...\n", pkg.Name)
			imagePackageVersion := formatVersion(&packageValues.Repositories[repo].Packages[i], "_").concat

			packageBundlePath := filepath.Join(projectRootDir, constants.PackageBundlesDir, pkg.Name+"-"+imagePackageVersion)
			if err := utils.CreateDir(packageBundlePath); err != nil {
				return err
			}

			// untar the package bundle
			tarBallFilePath := filepath.Join(projectRootDir, constants.PackageBundlesDir, pkg.Name+"-"+imagePackageVersion+".tar.gz")
			r, err := os.Open(tarBallFilePath)
			if err != nil {
				return fmt.Errorf("couldn't open tar file %s: %w", tarBallFilePath, err)
			}
			if err := utils.Untar(packageBundlePath, r); err != nil {
				return fmt.Errorf("couldn't untar package bundle: %w", err)
			}

			// push the package bundle to remote registry
			imgpkgCmd := exec.Command(
				filepath.Join(toolsBinDir, "imgpkg"),
				"push", "-b", registry+"/"+pkg.Name+":"+imagePackageVersion,
				"--file", packageBundlePath,
			) // #nosec G204

			var errBytes bytes.Buffer
			imgpkgCmd.Stderr = &errBytes
			if err := imgpkgCmd.Run(); err != nil {
				return fmt.Errorf("couldn't push the package bundle: %s", errBytes.String())
			}

			// remove the untared package bundle
			if err := os.RemoveAll(packageBundlePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func validatePackageBundlePushFlags(args []string) error {
	// At least one argument is expected to be passed in if --all is not specified.
	if !all && len(args) == 0 {
		return fmt.Errorf("at least one package bundle name is required to be specified")
	}

	if utils.IsStringEmpty(registry) {
		return fmt.Errorf("registry flag cannot be empty")
	}
	if utils.IsStringEmpty(version) {
		return fmt.Errorf("version flag cannot be empty")
	}
	return nil
}

// prunePackages will update in place the given map of repository packages to
// contain only the bundle packages listed in csvBundles.
func prunePackages(repos map[string]Repository, csvBundles string) error {
	bundles := strings.Split(csvBundles, ",")

	prunedPkgs := make(map[string][]Package)
	for repoName := range repos {
		prunedPkgs[repoName] = []Package{}
	}

	for _, bundle := range bundles {
		var bundleFound bool

	RepoLoop:
		for repoName := range repos {
			for i := range repos[repoName].Packages {
				pkg := repos[repoName].Packages[i]
				if pkg.Name == bundle {
					bundleFound = true
					prunedPkgs[repoName] = append(prunedPkgs[repoName], pkg)
					break RepoLoop
				}
			}
		}

		if !bundleFound {
			return fmt.Errorf("unable to find package bundle %q", bundle)
		}
	}

	for repoName, pkgs := range prunedPkgs {
		tmpRepo := repos[repoName]
		tmpRepo.Packages = pkgs
		repos[repoName] = tmpRepo
	}

	return nil
}
