// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

func (c *client) GetUpdateStatus() (providersNeedsUpdate, bomsNeedUpdate, tkgConfigNeedsUpdate bool, err error) {
	providersNeedsUpdate, err = c.CheckProvidersNeedUpdate()
	if err != nil {
		return
	}

	bomsNeedUpdate, err = c.CheckBOMsNeedUpdate()
	if err != nil {
		return
	}

	tkgConfigNeedsUpdate = false
	if !providersNeedsUpdate && os.Getenv(constants.SuppressUpdateEnvar) == "" {
		tkgConfigNeedsUpdate, _, err = c.CheckTkgConfigNeedUpdate()
		if err != nil {
			return
		}
	}
	return
}

// CheckProvidersNeedUpdate checks if .tkg/providers/config.yaml is up-to-date.
func (c *client) CheckProvidersNeedUpdate() (bool, error) {
	tkgDir, err := c.tkgConfigPathsClient.GetTKGDirectory()
	if err != nil {
		return false, err
	}

	providerDir := filepath.Join(tkgDir, constants.LocalProvidersFolderName)
	if _, err := os.Stat(providerDir); os.IsNotExist(err) {
		return false, nil
	}

	currentChecksumPath := filepath.Join(tkgDir, constants.LocalProvidersFolderName, constants.LocalProvidersChecksumFileName)
	if _, err := os.Stat(currentChecksumPath); os.IsNotExist(err) {
		return true, nil
	}

	zipPath := filepath.Join(tkgDir, constants.LocalProvidersZipFileName)
	err = c.saveTemplatesZipFile(zipPath)
	if err != nil {
		return false, errors.Wrap(err, "failed to check providers template update")
	}

	defer func() {
		if err := os.Remove(zipPath); err != nil {
			log.Infof("Unable to remove temporary providers.zip file %s, Error: %s", zipPath, err.Error())
		}
	}()

	currentChecksumBytes, err := os.ReadFile(currentChecksumPath)
	if err != nil {
		return false, errors.Wrap(err, "cannot read the original providers config.yaml")
	}

	newChecksumBytes, err := getBundledProvidersChecksum(zipPath)
	if err != nil {
		return false, errors.Wrap(err, "cannot read the bundled providersconfig.yaml")
	}

	return !bytes.Equal(currentChecksumBytes, newChecksumBytes), nil
}

// CheckBOMsNeedUpdate checks if bom files are up-to-date.
// returns true if $HOME/.tkg/bom directory exists, not empty and doesn't contain the defaultBoM file
func (c *client) CheckBOMsNeedUpdate() (bool, error) {
	var err error
	bomsDir, err := c.tkgConfigPathsClient.GetTKGBoMDirectory()
	if err != nil {
		return false, err
	}
	if _, err = os.Stat(bomsDir); os.IsNotExist(err) {
		return false, nil
	}

	isBOMDirectoryEmpty, err := isDirectoryEmpty(bomsDir)
	if err != nil {
		return false, errors.Wrap(err, "failed to check BOM directory is empty")
	}

	// if directory is empty we don't need update
	if isBOMDirectoryEmpty {
		return false, nil
	}

	// if defaultBOMfile doesn't exist we need to to update BOMs
	defaultBOMFile, _ := c.tkgBomClient.GetDefaultBoMFilePath()
	if _, err = os.Stat(defaultBOMFile); os.IsNotExist(err) {
		return true, nil
	}

	defaultTKRVersion, err := c.tkgBomClient.GetDefaultTKRVersion()
	if err != nil {
		return false, errors.Wrap(err, "failed to get default TKr version")
	}
	defaultTKRBOMFileName := fmt.Sprintf("tkr-bom-%s.yaml", defaultTKRVersion)
	defaultTKRBOMFilePath := filepath.Join(bomsDir, defaultTKRBOMFileName)
	if _, err = os.Stat(defaultTKRBOMFilePath); os.IsNotExist(err) {
		return true, nil
	}

	// for any other error return error
	if err != nil {
		return false, errors.Wrap(err, "failed to check BOMs need update")
	}
	return false, nil
}
