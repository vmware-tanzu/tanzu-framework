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

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/lockconfig"
	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/constants"
	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/utils"
)

var packageRepository, version, subVersion string
var all bool

// packageBundleGenerateCmd is for generating package bundle
var packageBundleGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate package bundle",
	RunE:  runPackageBundleGenerate,
}

func init() {
	packageBundleCmd.AddCommand(packageBundleGenerateCmd)
	packageBundleGenerateCmd.Flags().StringVar(&packageRepository, "repository", "", "Package repository of the package bundle being created")
	packageBundleGenerateCmd.Flags().StringVar(&registry, "registry", "", "OCI registry where the package bundle image needs to be stored")
	packageBundleGenerateCmd.Flags().StringVar(&version, "version", "", "Package bundle version")
	packageBundleGenerateCmd.Flags().StringVar(&subVersion, "sub-version", "", "Package bundle subversion")
	packageBundleGenerateCmd.Flags().BoolVar(&all, "all", false, "Generate all package bundles in a repository")
	packageBundleGenerateCmd.MarkFlagRequired("version") //nolint: errcheck
}

func runPackageBundleGenerate(cmd *cobra.Command, args []string) error {
	if err := validatePackageBundleGenerateFlags(); err != nil {
		return err
	}
	packageName := ""
	if len(args) == 1 {
		packageName = args[0]
	}

	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}

	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)

	if all {
		if err := generatePackageBundles(projectRootDir, toolsBinDir); err != nil {
			return fmt.Errorf("couldn't generate imgpkg lock output file: %w", err)
		}
	} else {
		fmt.Printf("Generating %q package bundle...\n", packageName)
		packagePath := filepath.Join(projectRootDir, "packages", packageName)
		if err := generateSingleImgpkgLockOutput(projectRootDir, toolsBinDir, packagePath); err != nil {
			return fmt.Errorf("couldn't generate imgpkg lock output file: %w", err)
		}
		if err := generatePackageBundle(projectRootDir, packageName, packagePath); err != nil {
			return fmt.Errorf("couldn't generate the package bundle: %w", err)
		}
		pkg, err := getPackageFromPackageValues(projectRootDir, packageName)
		if err != nil {
			return err
		}
		if err := generatePackageCR(projectRootDir,
			toolsBinDir,
			registry,
			filepath.Join(projectRootDir, "build", "packages"),
			filepath.Join(projectRootDir, constants.PackageValuesFilePath),
			&pkg); err != nil {
			return err
		}
	}
	return nil
}

func validatePackageBundleGenerateFlags() error {
	if utils.IsStringEmpty(version) {
		return fmt.Errorf("version flag cannot be empty")
	}
	return nil
}

func generateSingleImgpkgLockOutput(projectRootDir, toolsBinDir, packagePath string) error {
	if err := utils.RunMakeTarget(packagePath, "configure-package"); err != nil {
		return err
	}

	imgpkgLockOutputDir := filepath.Join(packagePath, "bundle", ".imgpkg")
	if err := utils.CreateDir(imgpkgLockOutputDir); err != nil {
		return err
	}

	// run the ytt command on the package config and pipe the output to kbld command to generate imgpkg lock output file
	yttCmd := exec.Command(filepath.Join(toolsBinDir, "ytt"),
		"--ignore-unknown-comments",
		"-f", filepath.Join(packagePath, "bundle", "config")) // #nosec G204
	kbldCmd := exec.Command(filepath.Join(toolsBinDir, "kbld"),
		"-f", "-", // kbld interprets this as reading a file from stdin
		"-f", filepath.Join(projectRootDir, constants.KbldConfigFilePath),
		"--imgpkg-lock-output", filepath.Join(imgpkgLockOutputDir, "images.yml")) // #nosec G204

	pipe, err := yttCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate imgpkg lock output file: %w", err)
	}
	defer pipe.Close()
	var kbldCmdErrBytes bytes.Buffer
	kbldCmd.Stdin = pipe
	kbldCmd.Stderr = &kbldCmdErrBytes

	if err := yttCmd.Start(); err != nil {
		return fmt.Errorf("couldn't run ytt command: %w", err)
	}
	if err := kbldCmd.Run(); err != nil {
		return fmt.Errorf("couldn't run kbld command to generate imgpkg lock output file: %s", kbldCmdErrBytes.String())
	}

	if err := utils.RunMakeTarget(packagePath, "reset-package"); err != nil {
		return err
	}
	return nil
}

func generatePackageBundle(projectRootDir, packageName, packagePath string) error {
	if err := utils.RunMakeTarget(packagePath, "configure-package"); err != nil {
		return err
	}

	imagePackageVersion := version
	if subVersion != "" {
		imagePackageVersion = version + "_" + subVersion
	}

	// create tarball of package bundle
	tarBallPath := filepath.Join(projectRootDir, constants.PackageBundlesDir)
	tarBallFileName := packageName + "-" + imagePackageVersion + ".tar.gz"
	pathToContents := filepath.Join(packagePath, "bundle")
	if err := utils.CreateTarball(tarBallPath, tarBallFileName, pathToContents); err != nil {
		return fmt.Errorf("couldn't generate package bundle: %w", err)
	}

	if err := utils.RunMakeTarget(packagePath, "reset-package"); err != nil {
		return err
	}
	return nil
}

func generatePackageBundles(projectRootDir, toolsBinDir string) error {
	packageValues, err := readPackageValues(projectRootDir)
	if err != nil {
		return err
	}

	packageRepos, err := filterPackageRepos(packageValues)
	if err != nil {
		return err
	}

	for _, repo := range packageRepos {
		for i, pkg := range packageValues.Repositories[repo].Packages {
			fmt.Printf("Generating %q package bundle...\n", pkg.Name)
			imagePackageVersion := version
			if subVersion != "" {
				imagePackageVersion = version + "_" + subVersion
			}
			packagePath := filepath.Join(projectRootDir, "packages", pkg.Name)
			if err := utils.RunMakeTarget(packagePath, "configure-package"); err != nil {
				return err
			}

			// generate package bundle imgpkg lock output file
			if err := generateSingleImgpkgLockOutput(projectRootDir, toolsBinDir, packagePath); err != nil {
				return fmt.Errorf("couldn't generate imgpkg lock output file: %w", err)
			}

			// push the imgpkg bundle to local registry
			lockOutputFile := pkg.Name + "-" + imagePackageVersion + "-lock-output.yaml"
			imgpkgCmd := exec.Command(
				filepath.Join(toolsBinDir, "imgpkg"),
				"push", "-b", constants.LocalRegistryURL+"/"+pkg.Name+":"+imagePackageVersion,
				"--file", filepath.Join(packagePath, "bundle"),
				"--lock-output", lockOutputFile,
			) // #nosec G204

			var imgpkgCmdErrBytes bytes.Buffer
			imgpkgCmd.Stderr = &imgpkgCmdErrBytes
			if err := imgpkgCmd.Run(); err != nil {
				return fmt.Errorf("couldn't push the imgpkg bundle: %s", imgpkgCmdErrBytes.String())
			}

			// update the package version and sha256 in package-values-sha256.yaml file
			lockOutputData, err := os.ReadFile(lockOutputFile)
			if err != nil {
				return fmt.Errorf("couldn't read lock output file %s: %w", lockOutputFile, err)
			}

			bundleLock := lockconfig.BundleLock{}
			if err := yaml.Unmarshal(lockOutputData, &bundleLock); err != nil {
				return fmt.Errorf("error while unmarshaling: %w", err)
			}

			packageValues.Repositories[repo].Packages[i].Version = getPackageVersion(version)
			packageValues.Repositories[repo].Packages[i].Sha256 = utils.AfterString(bundleLock.Bundle.Image, constants.LocalRegistryURL+"/"+pkg.Name+"@sha256:")
			yamlData, err := yaml.Marshal(&packageValues)
			if err != nil {
				return fmt.Errorf("error while marshaling: %w", err)
			}

			comments := []byte("#@data/values\n---\n")
			yamlData = append(comments, yamlData...)
			if err := os.WriteFile(filepath.Join(projectRootDir, constants.PackageValuesSha256FilePath), yamlData, 0755); err != nil {
				return err
			}

			if err := generatePackageBundle(projectRootDir, pkg.Name, packagePath); err != nil {
				return fmt.Errorf("couldn't generate package bundle: %w", err)
			}

			if err := generatePackageCR(projectRootDir,
				toolsBinDir,
				registry,
				filepath.Join(projectRootDir, "build", "packages"),
				filepath.Join(projectRootDir, constants.PackageValuesFilePath),
				&pkg); err != nil {
				return err
			}

			// remove lock output files
			os.Remove(lockOutputFile)

			if err := utils.RunMakeTarget(packagePath, "reset-package"); err != nil {
				return err
			}
		}
	}
	return nil
}

func getPackageVersion(version string) string {
	pkgVersion := version
	if string(version[0]) == "v" {
		pkgVersion = version[1:]
	}
	return pkgVersion
}
