// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

// ProcessCarvelPackage processes a carvel package and returns a configuration YAML
// Downloads package to temporary directory and processes the package by
// implementing equivalent functionality as the command: `ytt -f <path> [-f <values-files>] | kbld -f -`
func ProcessCarvelPackage(image string, valuesFiles ...string) ([]byte, error) {
	pkgDir, err := DownloadImageBundleAndSaveFilesToTempDir(image)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get resource files from discovery")
	}
	defer os.RemoveAll(pkgDir)
	return CarvelPackageProcessor(pkgDir, image, valuesFiles...)
}

// CarvelPackageProcessor processes a carvel package and returns a configuration YAML file
func CarvelPackageProcessor(pkgDir, image string, valuesFiles ...string) ([]byte, error) {
	// Each package contains `config` and `.imgpkg` directory
	// `config` directory contains ytt files
	// `.imgpkg` directory contains ImageLock configuration for ImageResolution
	configDir := filepath.Join(pkgDir, "config")
	files := append([]string{configDir}, valuesFiles...)
	bytes, err := ProcessYTTPackage(files...)
	if err != nil {
		return nil, errors.Wrap(err, "error while running ytt")
	}

	f, err := os.CreateTemp("", "ytt-processed")
	if err != nil {
		return nil, errors.Wrap(err, "error while creating temp directory")
	}
	defer os.Remove(f.Name())

	err = utils.SaveFile(f.Name(), bytes)
	if err != nil {
		return nil, errors.Wrap(err, "error while saving file")
	}

	inputFilesForImageResolution := []string{f.Name()}

	// Use `.imgpkg` directory if exists for ImageResolution
	imgpkgDir := filepath.Join(pkgDir, ".imgpkg")
	if utils.PathExists(imgpkgDir) {
		inputFilesForImageResolution = append(inputFilesForImageResolution, imgpkgDir)
	}

	return ResolveImagesInPackage(inputFilesForImageResolution)
}
