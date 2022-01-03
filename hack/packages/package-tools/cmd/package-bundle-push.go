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
	"gopkg.in/yaml.v2"

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
	packageBundlePushCmd.MarkFlagRequired("repository") //nolint: errcheck
	packageBundlePushCmd.MarkFlagRequired("registry")   //nolint: errcheck
	packageBundlePushCmd.MarkFlagRequired("version")    //nolint: errcheck
}

func runPackageBundlePush(cmd *cobra.Command, args []string) error {
	if err := validatePackageBundlePushFlags(); err != nil {
		return err
	}
	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}
	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)

	packageValuesData, err := os.ReadFile(filepath.Join(projectRootDir, constants.PackageValuesFilePath))
	if err != nil {
		return fmt.Errorf("couldn't read file %s: %w", packageValuesFile, err)
	}

	packageValues := PackageValues{}
	if err := yaml.Unmarshal(packageValuesData, &packageValues); err != nil {
		return fmt.Errorf("error while unmarshalling: %w", err)
	}

	repository, found := packageValues.Repositories[packageRepository]
	if !found {
		return fmt.Errorf("%s repository not found", packageRepository)
	}

	for _, pkg := range repository.Packages {
		fmt.Printf("Pushing %q package bundle...\n", pkg.Name)
		imagePackageVersion := version
		if subVersion != "" {
			imagePackageVersion = version + "_" + subVersion
		}

		packageBundlePath := filepath.Join(projectRootDir, constants.PackageBundlesDir, packageRepository, pkg.Name+"-"+imagePackageVersion)
		if err := utils.CreateDir(packageBundlePath); err != nil {
			return err
		}

		// untar the package bundle
		tarBallFilePath := filepath.Join(projectRootDir, constants.PackageBundlesDir, packageRepository, pkg.Name+"-"+imagePackageVersion+".tar.gz")
		r, err := os.Open(tarBallFilePath)
		if err != nil {
			return fmt.Errorf("couldn't open tar file %s: %w", tarBallFilePath, err)
		}
		if err := utils.Untar(packageBundlePath, r); err != nil {
			return fmt.Errorf("couldn't untar package bundle: %w", err)
		}

		// push the package bundle to remote registry
		imgpkgCmd := exec.Command(filepath.Join(toolsBinDir, "imgpkg"),
			"push", "-b", registry+"/"+pkg.Name+":"+imagePackageVersion,
			"--file", packageBundlePath) // #nosec G204

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
	return nil
}

func validatePackageBundlePushFlags() error {
	if utils.IsStringEmpty(packageRepository) {
		return fmt.Errorf("repository flag cannot be empty")
	}
	if utils.IsStringEmpty(registry) {
		return fmt.Errorf("registry flag cannot be empty")
	}
	if utils.IsStringEmpty(version) {
		return fmt.Errorf("version flag cannot be empty")
	}
	return nil
}
