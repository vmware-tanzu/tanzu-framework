// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	_ "embed" // required to embed file
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

var packageValuesFile, registry string

// repoBundleGenerateCmd is for generating package repo bundle
var repoBundleGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate package repo bundle",
	RunE:  runRepoBundleGenerate,
}

func init() {
	repositoryBundleCmd.AddCommand(repoBundleGenerateCmd)
	repoBundleGenerateCmd.Flags().StringVar(&packageRepository, "repository", "", "Package repository of the package bundles")
	repoBundleGenerateCmd.Flags().StringVar(&registry, "registry", "", "OCI registry where the package repo bundle image needs to be stored")
	repoBundleGenerateCmd.Flags().StringVar(&version, "version", "", "Package version of a package in repo bundle")
	repoBundleGenerateCmd.Flags().StringVar(&subVersion, "sub-version", "", "Package subversion of a package in repo bundle")
	repoBundleGenerateCmd.Flags().StringVar(&packageValuesFile, "package-values-file", "", "File containing the packages configuration")
	repoBundleGenerateCmd.MarkFlagRequired("repository") //nolint: errcheck
	repoBundleGenerateCmd.MarkFlagRequired("registry")   //nolint: errcheck
	repoBundleGenerateCmd.MarkFlagRequired("version")    //nolint: errcheck
}

//go:embed templates/images-tmpl.yaml
var imagesLockTemplate string

//go:embed templates/package-helpers.lib.yaml
var packageHelpersLib string

//go:embed templates/package-cr-overlay.yaml
var packageCrOverlay string

//go:embed templates/package-metadata-cr-overlay.yaml
var packageMetadataCrOverlay string

//go:embed templates/packagerepo-tmpl.yaml
var packageRepoTemplate string

func runRepoBundleGenerate(cmd *cobra.Command, args []string) error {
	if err := validateRepoBundleGenerateFlags(); err != nil {
		return err
	}
	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}

	if packageValuesFile == "" {
		if err := generatePackageBundlesSha256(projectRootDir, constants.LocalRegistryURL); err != nil {
			return fmt.Errorf("couldn't generate package-values-sha256.yaml: %w", err)
		}
		packageValuesFile = filepath.Join(projectRootDir, constants.PackageValuesSha256FilePath)
	}

	if err := generateRepoBundle(projectRootDir); err != nil {
		return fmt.Errorf("couldn't generate package repo bundle: %w", err)
	}

	return nil
}

func validateRepoBundleGenerateFlags() error {
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

func generatePackageBundlesSha256(projectRootDir, localRegistry string) error {
	packageValuesData, err := os.ReadFile(filepath.Join(projectRootDir, constants.PackageValuesFilePath))
	if err != nil {
		return fmt.Errorf("couldn't read package-values.yaml: %w ", err)
	}

	packageValues := PackageValues{}
	if err := yaml.Unmarshal(packageValuesData, &packageValues); err != nil {
		return err
	}

	repository, found := packageValues.Repositories[packageRepository]
	if !found {
		return fmt.Errorf("repository not found %s", packageRepository)
	}

	for i := range repository.Packages {
		pkg := repository.Packages[i]
		formattedVer := formatVersion(&repository.Packages[i], "_")
		packagePath := filepath.Join(projectRootDir, "packages", pkg.Name)
		toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)

		if err := utils.RunMakeTarget(packagePath, "configure-package", getEnvArrayFromMap(pkg.Env)...); err != nil {
			return err
		}

		// push the imgpkg bundle to local registry
		lockOutputFile := pkg.Name + "-" + formattedVer.concat + "-lock-output.yaml"
		imgpkgCmd := exec.Command(
			filepath.Join(toolsBinDir, "imgpkg"),
			"push",
			"-b", localRegistry+"/"+pkg.Name+":"+formattedVer.concat,
			"--file", filepath.Join(packagePath, "bundle"),
			"--lock-output", lockOutputFile,
		) // #nosec G204

		var errBytes bytes.Buffer
		imgpkgCmd.Stderr = &errBytes
		if err := imgpkgCmd.Run(); err != nil {
			return fmt.Errorf("couldn't push the imgpkg bundle: %s", errBytes.String())
		}

		// update the package version and sha256
		lockOutputData, err := os.ReadFile(lockOutputFile)
		if err != nil {
			return fmt.Errorf("couldn't read lock output file %s: %w", lockOutputFile, err)
		}

		bundleLock := lockconfig.BundleLock{}
		if err := yaml.Unmarshal(lockOutputData, &bundleLock); err != nil {
			return fmt.Errorf("error while unmarshaling: %w", err)
		}

		packageValues.Repositories[packageRepository].Packages[i].Version = formattedVer.noV
		packageValues.Repositories[packageRepository].Packages[i].Sha256 = utils.AfterString(
			bundleLock.Bundle.Image,
			localRegistry+"/"+pkg.Name+"@sha256:",
		)
		packageValues.Repositories[packageRepository].Packages[i].PackageSubVersion = formattedVer.subVersion
		yamlData, err := yaml.Marshal(&packageValues)
		if err != nil {
			return fmt.Errorf("error while marshaling: %w", err)
		}

		comments := []byte("#@data/values\n---\n")
		yamlData = append(comments, yamlData...)
		if err := os.WriteFile(filepath.Join(projectRootDir, constants.PackageValuesSha256FilePath), yamlData, 0755); err != nil {
			return err
		}

		// remove lock output files
		if err := os.Remove(lockOutputFile); err != nil {
			return fmt.Errorf("couldn't remove file %s: %w", lockOutputFile, err)
		}

		if err := utils.RunMakeTarget(packagePath, "reset-package"); err != nil {
			return err
		}
	}
	return nil
}

func generateRepoBundle(projectRootDir string) error {
	fmt.Printf("Generating %q repo bundle...\n", packageRepository)
	if err := utils.CreateDir(filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, ".imgpkg")); err != nil {
		return err
	}

	if err := utils.CreateDir(filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, "packages")); err != nil {
		return err
	}

	toolsBinDir := filepath.Join(projectRootDir, constants.ToolsBinDirPath)

	// write the ytt lib to a temp file and delete it later
	packageHelpersLibFile, err := getTempPackageHelpersLib(packageHelpersLib)
	if err != nil {
		return err
	}
	defer os.Remove(packageHelpersLibFile)

	// generate repo bundle image lock output file
	yttCmd := exec.Command(
		filepath.Join(toolsBinDir, "ytt"),
		"-f-",
		"-f", packageHelpersLibFile,
		"-f", packageValuesFile,
		"-v", "packageRepository="+packageRepository,
		"-v", "registry="+registry,
	) // #nosec G204

	outFilePath := filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, ".imgpkg", "images.yml")
	outfile, err := os.Create(outFilePath)
	if err != nil {
		return fmt.Errorf("error creating file %s : %w", outFilePath, err)
	}
	defer outfile.Close()
	var errBytes bytes.Buffer
	yttCmd.Stdout = outfile
	yttCmd.Stderr = &errBytes

	yttCmdStdin, err := yttCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate imgpkg lock output file for repo bundle: %w", err)
	}
	_, err = yttCmdStdin.Write([]byte(imagesLockTemplate))
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate imgpkg lock output file for repo bundle: %w", err)
	}
	yttCmdStdin.Close()

	if err = yttCmd.Run(); err != nil {
		return fmt.Errorf("couldn't generate the image lock output file: %s", errBytes.String())
	}

	packageValuesData, err := os.ReadFile(packageValuesFile)
	if err != nil {
		return fmt.Errorf("couldn't read file %s: %w", packageValuesFile, err)
	}

	packageValues := PackageValues{}
	if err := yaml.Unmarshal(packageValuesData, &packageValues); err != nil {
		return fmt.Errorf("error while unmarshaling: %w", err)
	}

	repository, found := packageValues.Repositories[packageRepository]
	if !found {
		return fmt.Errorf("%s repository not found", packageRepository)
	}

	pkgRepoPkgsDir := filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository, "packages")
	for i := range repository.Packages {
		if err := generatePackageCR(projectRootDir, toolsBinDir, registry, pkgRepoPkgsDir, packageValuesFile, &repository.Packages[i]); err != nil {
			return fmt.Errorf("couldn't generate the package: %w", err)
		}
	}

	// create tarball of repo bundle
	tarballVersion := formatVersion(nil, "_").concat
	tarBallPath := filepath.Join(projectRootDir, constants.RepoBundlesDir, packageRepository)
	tarBallFileName := packageRepository + "-repo-" + tarballVersion + ".tar.gz"
	if err := utils.CreateTarball(tarBallPath, tarBallFileName, tarBallPath); err != nil {
		return fmt.Errorf("couldn't generate package bundle: %w", err)
	}

	return nil
}

func generatePackageCR(projectRootDir, toolsBinDir, registry, packageArtifactDirectory, packageValuesFile string, pkg *Package) error {
	// package values file
	fmt.Printf("Generating Package CR for package '%s:%s'...\n", pkg.Name, pkg.Version)
	if err := utils.CreateDir(filepath.Join(packageArtifactDirectory, pkg.Name+"."+pkg.Domain)); err != nil {
		return err
	}

	formattedVer := formatVersion(pkg, "+")

	// write the ytt lib to a temp file and delete it later
	packageHelpersLibFile, err := getTempPackageHelpersLib(packageHelpersLib)
	if err != nil {
		return err
	}
	defer os.Remove(packageHelpersLibFile)

	// generate Package CR and write it to a file
	packageYttCmd := exec.Command(
		filepath.Join(toolsBinDir, "ytt"),
		"-f-",
		"-f", filepath.Join(projectRootDir, "packages", pkg.Name, "package.yaml"),
		"-f", packageHelpersLibFile,
		"-f", packageValuesFile,
		"-v", "packageRepository="+packageRepository,
		"-v", "packageName="+pkg.Name,
		"-v", "registry="+registry,
		"-v", "timestamp="+utils.GetFormattedCurrentTime(),
		"-v", "version="+formattedVer.noV,
		"-v", "subVersion="+formattedVer.subVersion,
	) // #nosec G204

	packageFileName := formattedVer.concatNoV + ".yml"
	packageFilePath := filepath.Join(packageArtifactDirectory, pkg.Name+"."+pkg.Domain, packageFileName)
	packageFile, err := os.Create(packageFilePath)
	if err != nil {
		return fmt.Errorf("couldn't create file %s: %w", packageFilePath, err)
	}
	defer packageFile.Close()
	var packageYttCmdErrBytes bytes.Buffer
	packageYttCmd.Stdout = packageFile
	packageYttCmd.Stderr = &packageYttCmdErrBytes

	packageYttCmdStdin, err := packageYttCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate Package CR for repo bundle: %w", err)
	}
	_, err = packageYttCmdStdin.Write([]byte(packageCrOverlay))
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate Package CR for repo bundle: %w", err)
	}
	packageYttCmdStdin.Close()

	err = packageYttCmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't generate Package CR %s: %s", pkg.Name, packageYttCmdErrBytes.String())
	}

	// generate PacakageMetadata CR and write it to a file
	packageMetadataYttCmd := exec.Command(
		filepath.Join(toolsBinDir, "ytt"), "-f", filepath.Join(projectRootDir, "packages", pkg.Name, "metadata.yaml"),
		"-f-",
		"-f", packageHelpersLibFile,
		"-f", packageValuesFile,
		"-v", "packageRepository="+packageRepository,
		"-v", "packageName="+pkg.Name,
		"-v", "registry="+registry,
	) // #nosec G204

	packageMetadataFilePath := filepath.Join(packageArtifactDirectory, pkg.Name+"."+pkg.Domain, "metadata.yml")
	metadataFile, err := os.Create(packageMetadataFilePath)
	if err != nil {
		return fmt.Errorf("couldn't create file %s: %w", packageMetadataFilePath, err)
	}
	defer metadataFile.Close()
	var packageMetadataYttCmdErrBytes bytes.Buffer
	packageMetadataYttCmd.Stdout = metadataFile
	packageMetadataYttCmd.Stderr = &packageMetadataYttCmdErrBytes

	packageMetadataYttCmdStdin, err := packageMetadataYttCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate PackageMetadata CR for repo bundle: %w", err)
	}
	_, err = packageMetadataYttCmdStdin.Write([]byte(packageMetadataCrOverlay))
	if err != nil {
		return fmt.Errorf("couldn't run ytt command to generate PackageMetadata CR for repo bundle: %w", err)
	}
	packageMetadataYttCmdStdin.Close()
	err = packageMetadataYttCmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't generate PackageMetadata CR %s: %s", pkg.Name, packageMetadataYttCmdErrBytes.String())
	}
	return nil
}
