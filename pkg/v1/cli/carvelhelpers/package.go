// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"os"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// ProcessCarvelPackage processes a carvel package and returns a configuration YAML
// Downloads package to temporary directory and processes the package by
// implementing equivalent functionality as the command: `ytt -f <path> | kbld -f -`
func ProcessCarvelPackage(image string) ([]byte, error) {
	configDir, err := DownloadImageBundleAndSaveFilesToTempDir(image)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get resource files from discovery")
	}
	defer os.RemoveAll(configDir)

	bytes, err := ProcessYTTPackage(configDir)
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

	return ResolveImagesInPackage([]string{f.Name()})
}
