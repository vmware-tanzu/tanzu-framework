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
var all, thick bool

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
	packageBundleGenerateCmd.Flags().BoolVar(&thick, "thick", false, "Include thick tarball(s) in package bundle(s)")
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

		pkgs, err := getPackageFromPackageValues(projectRootDir, packageName)
		if err != nil {
			return err
		}

		for i := range pkgs {
			packagePath := filepath.Join(projectRootDir, "packages", packageName)
			if err := generateSingleImgpkgLockOutput(toolsBinDir, packagePath, getEnvArrayFromMap(pkgs[i].Env)...); err != nil {
				return fmt.Errorf("couldn't generate imgpkg lock output file: %w", err)
			}

			if err := generatePackageBundle(&pkgs[i], projectRootDir, toolsBinDir, packageName, packagePath); err != nil {
				return fmt.Errorf("couldn't generate the package bundle: %w", err)
			}
			buildPkgDir := filepath.Join(projectRootDir, "build", "packages")
			pkgValsDir := filepath.Join(projectRootDir, constants.PackageValuesFilePath)
			if err := generatePackageCR(projectRootDir, toolsBinDir, registry, buildPkgDir, pkgValsDir, &pkgs[i]); err != nil {
				return err
			}
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

func generateSingleImgpkgLockOutput(toolsBinDir, packagePath string, envArray ...string) error {
	if err := utils.RunMakeTarget(packagePath, "configure-package", envArray...); err != nil {
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

	kbldArgs := []string{
		"-f", "-",
		"--imgpkg-lock-output", filepath.Join(imgpkgLockOutputDir, "images.yml"),
	}
	if _, err := os.Stat(filepath.Join(packagePath, "kbld-config.yaml")); err == nil {
		kbldArgs = append(kbldArgs, "-f", filepath.Join(packagePath, "kbld-config.yaml"))
	}

	kbldCmd := exec.Command(filepath.Join(toolsBinDir, "kbld"), kbldArgs...) // #nosec G204

	pipe, err := yttCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate imgpkg lock output file: %w", err)
	}
	defer pipe.Close()
	var kbldCmdErrBytes, yttCmdErrBytes bytes.Buffer
	kbldCmd.Stdin = pipe
	kbldCmd.Stderr = &kbldCmdErrBytes
	yttCmd.Stderr = &yttCmdErrBytes

	fmt.Println(yttCmd.String())
	if err := yttCmd.Start(); err != nil {
		return fmt.Errorf("couldn't run ytt command: %w", err)
	}

	fmt.Println(kbldCmd.String())
	if err := kbldCmd.Run(); err != nil {
		return fmt.Errorf("couldn't run kbld command to generate imgpkg lock output file: %s", kbldCmdErrBytes.String())
	}

	if yttCmdErrBytes.String() != "" {
		return fmt.Errorf("couldn't run ytt command: %s", yttCmdErrBytes.String())
	}

	if err := utils.RunMakeTarget(packagePath, "reset-package"); err != nil {
		return err
	}
	return nil
}

func generatePackageBundle(pkg *Package, projectRootDir, toolsBinDir, packageName, packagePath string) error {
	if err := utils.RunMakeTarget(packagePath, "configure-package", getEnvArrayFromMap(pkg.Env)...); err != nil {
		return err
	}

	formattedVer := formatVersion(pkg, "_")
	imagePackageVersion := formattedVer.concat

	// create tarball of package bundle
	tarBallPath := filepath.Join(projectRootDir, constants.PackageBundlesDir)
	tarBallFileName := packageName + "-" + imagePackageVersion + ".tar.gz"
	pathToContents := filepath.Join(packagePath, "bundle")
	if err := utils.CreateTarball(tarBallPath, tarBallFileName, pathToContents); err != nil {
		return fmt.Errorf("couldn't generate package bundle: %w", err)
	}

	// create thick tarball
	if thick {
		fmt.Println("Including thick tarball...")
		var cmdErr bytes.Buffer

		packageURL := fmt.Sprintf("%s/%s:%s", constants.LocalRegistryURL, packageName, imagePackageVersion)
		imgpkgPushCmd := exec.Command(
			filepath.Join(toolsBinDir, "imgpkg"),
			"push",
			"-b", packageURL,
			"--file", filepath.Join(packagePath, "bundle"),
		) // #nosec G204
		imgpkgPushCmd.Stderr = &cmdErr
		if err := imgpkgPushCmd.Run(); err != nil {
			fmt.Println("cmd:", imgpkgPushCmd.String())
			fmt.Println("err:", err)
			return fmt.Errorf("pushing package bundle to local registry: %s", cmdErr.String())
		}

		tarBallFileName = packageName + "-" + imagePackageVersion + "-thick.tar.gz"
		imgpkgCopyCmd := exec.Command(
			filepath.Join(toolsBinDir, "imgpkg"),
			"copy",
			"-b", packageURL,
			"--to-tar", filepath.Join(tarBallPath, tarBallFileName),
		) // #nosec G204
		imgpkgCopyCmd.Stderr = &cmdErr
		if err := imgpkgCopyCmd.Run(); err != nil {
			return fmt.Errorf("generating thick tarball: %s", cmdErr.String())
		}
	}

	if err := utils.RunMakeTarget(packagePath, "reset-package"); err != nil {
		return err
	}
	return nil
}

func generatePackageBundles(projectRootDir, toolsBinDir string) error {
	pkgVals, err := readPackageValues(projectRootDir)
	if err != nil {
		return err
	}
	packageRepos, err := filterPackageRepos(pkgVals)
	if err != nil {
		return err
	}

	for _, repo := range packageRepos {
		for i := range pkgVals.Repositories[repo].Packages {
			fmt.Printf("Generating %q package bundle...\n", pkgVals.Repositories[repo].Packages[i].Name)

			packagePath := filepath.Join(projectRootDir, "packages", pkgVals.Repositories[repo].Packages[i].Name)
			if err := utils.RunMakeTarget(packagePath, "configure-package", getEnvArrayFromMap(pkgVals.Repositories[repo].Packages[i].Env)...); err != nil {
				return err
			}

			// generate package bundle imgpkg lock output file
			if err := generateSingleImgpkgLockOutput(toolsBinDir, packagePath, getEnvArrayFromMap(pkgVals.Repositories[repo].Packages[i].Env)...); err != nil {
				return fmt.Errorf("couldn't generate imgpkg lock output file: %w", err)
			}

			// push the imgpkg bundle to local registry
			imagePackageVersion := formatVersion(&pkgVals.Repositories[repo].Packages[i], "_").concat
			lockOutputFile := pkgVals.Repositories[repo].Packages[i].Name + "-" + imagePackageVersion + "-lock-output.yaml"
			imgpkgCmd := exec.Command(
				filepath.Join(toolsBinDir, "imgpkg"),
				"push", "-b", constants.LocalRegistryURL+"/"+pkgVals.Repositories[repo].Packages[i].Name+":"+imagePackageVersion,
				"--file", filepath.Join(packagePath, "bundle"),
				"--lock-output", lockOutputFile,
			) // #nosec G204

			var imgpkgCmdErrBytes bytes.Buffer
			imgpkgCmd.Stderr = &imgpkgCmdErrBytes
			fmt.Println(imgpkgCmd.String())
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

			pkgVals.Repositories[repo].Packages[i].Version = formatVersion(&pkgVals.Repositories[repo].Packages[i], "_").version
			pkgVals.Repositories[repo].Packages[i].Sha256 = utils.AfterString(
				bundleLock.Bundle.Image,
				constants.LocalRegistryURL+"/"+pkgVals.Repositories[repo].Packages[i].Name+"@sha256:",
			)
			yamlData, err := yaml.Marshal(&pkgVals)
			if err != nil {
				return fmt.Errorf("error while marshaling: %w", err)
			}

			comments := []byte("#@data/values\n---\n")
			yamlData = append(comments, yamlData...)
			if err := os.WriteFile(filepath.Join(projectRootDir, constants.PackageValuesSha256FilePath), yamlData, 0755); err != nil {
				return err
			}

			err = generatePackageBundle(
				&pkgVals.Repositories[repo].Packages[i],
				projectRootDir,
				toolsBinDir,
				pkgVals.Repositories[repo].Packages[i].Name,
				packagePath,
			)
			if err != nil {
				return fmt.Errorf("couldn't generate package bundle: %w", err)
			}

			buildPkgsDir := filepath.Join(projectRootDir, "build", "packages")
			pkgValsPath := filepath.Join(projectRootDir, constants.PackageValuesFilePath)
			err = generatePackageCR(
				projectRootDir,
				toolsBinDir,
				registry,
				buildPkgsDir,
				pkgValsPath,
				&pkgVals.Repositories[repo].Packages[i],
			)
			if err != nil {
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
