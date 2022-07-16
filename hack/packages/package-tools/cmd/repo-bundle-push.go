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
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/lockconfig"
	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/constants"
	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/utils"
)

// repoBundlePushCmd is for pushing package repo bundle
var repoBundlePushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push package repo bundle",
	RunE:  runRepoBundlePush,
}

func init() {
	repositoryBundleCmd.AddCommand(repoBundlePushCmd)
	repoBundlePushCmd.Flags().StringVar(&packageRepository, "repository", "", "Package repository of the package repo bundle being pushed")
	repoBundlePushCmd.Flags().StringVar(&registry, "registry", "", "OCI registry where the package repo bundle image needs to be stored")
	repoBundlePushCmd.Flags().StringVar(&version, "version", "", "Package repo bundle version")
	repoBundlePushCmd.Flags().StringVar(&subVersion, "sub-version", "", "Package subversion of a package in repo bundle")
	repoBundlePushCmd.MarkFlagRequired("repository") //nolint: errcheck
	repoBundlePushCmd.MarkFlagRequired("registry")   //nolint: errcheck
	repoBundlePushCmd.MarkFlagRequired("version")    //nolint: errcheck
}

func runRepoBundlePush(cmd *cobra.Command, args []string) error {
	fmt.Printf("Pushing package repo bundle %q...\n", packageRepository)
	if err := validateRepoBundlePushFlags(); err != nil {
		return err
	}
	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}

	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)
	repoVersion := formatVersion(nil, "+").concat

	repoBundlePath := filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, packageRepository+"-repo-"+repoVersion)
	if err := utils.CreateDir(repoBundlePath); err != nil {
		return err
	}

	// untar the repo bundle
	tarBallVersion := formatVersion(nil, "_").concat
	tarBallFilePath := filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, packageRepository+"-repo-"+tarBallVersion+".tar.gz")
	r, err := os.Open(tarBallFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open tar file %s: %w", tarBallFilePath, err)
	}
	if err := utils.Untar(repoBundlePath, r); err != nil {
		return fmt.Errorf("couldn't untar package bundle: %w", err)
	}

	// push the repo bundle
	lockOutputFile := filepath.Join(repoBundlePath, packageRepository+"-repo-"+repoVersion+"-lock-output.yaml")
	imgpkgCmd := exec.Command(
		filepath.Join(toolsBinDir, "imgpkg"),
		"push",
		"-b",
		registry+"/"+packageRepository+":"+tarBallVersion,
		"--file", repoBundlePath,
		"--lock-output", lockOutputFile,
	) // #nosec G204

	var imgpkgCmdErrBytes bytes.Buffer
	imgpkgCmd.Stderr = &imgpkgCmdErrBytes
	if err := imgpkgCmd.Run(); err != nil {
		return fmt.Errorf("couldn't push the package repo bundle: %s", imgpkgCmdErrBytes.String())
	}

	// generate PackageRepository CR
	lockOutputData, err := os.ReadFile(lockOutputFile)
	if err != nil {
		return fmt.Errorf("couldn't read lock output file %s: %w", lockOutputFile, err)
	}

	bundleLock := lockconfig.BundleLock{}
	if err := yaml.Unmarshal(lockOutputData, &bundleLock); err != nil {
		return fmt.Errorf("error while unmarshaling: %w", err)
	}

	sha256 := strings.Split(bundleLock.Bundle.Image, ":")[1]

	// write the ytt lib to a temp file and delete it later
	packageHelpersLibFile, err := getTempPackageHelpersLib(packageHelpersLib)
	if err != nil {
		return err
	}
	defer os.Remove(packageHelpersLibFile)

	yttCmd := exec.Command(
		filepath.Join(toolsBinDir, "ytt"),
		"-f-",
		"-f", packageHelpersLibFile,
		"-f", filepath.Join(projectRootDir, constants.PackageValuesSha256FilePath),
		"-v", "packageRepository="+packageRepository,
		"-v", "registry="+registry,
		"-v", "sha256="+sha256,
	) // #nosec G204

	packageRepositoryCRFilePath := filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, packageRepository+"-repo-"+repoVersion+".yaml")
	outfile, err := os.Create(packageRepositoryCRFilePath)
	if err != nil {
		return fmt.Errorf("error creating file %s : %w", packageRepositoryCRFilePath, err)
	}
	defer outfile.Close()
	var errBytes bytes.Buffer
	yttCmd.Stdout = outfile
	yttCmd.Stderr = &errBytes
	yttCmdStdin, err := yttCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate PackageRepository CR: %w", err)
	}
	_, err = yttCmdStdin.Write([]byte(packageRepoTemplate))
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate PackageRepository CR: %w", err)
	}
	yttCmdStdin.Close()

	if err = yttCmd.Run(); err != nil {
		return fmt.Errorf("couldn't generate PackageRepository CR: %s", errBytes.String())
	}

	// remove untared package repo bundle
	if err := os.RemoveAll(repoBundlePath); err != nil {
		return err
	}

	return nil
}

func validateRepoBundlePushFlags() error {
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
