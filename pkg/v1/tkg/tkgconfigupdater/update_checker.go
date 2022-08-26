// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

// CheckProviderTemplatesNeedUpdate checks if .tkg/providers/config.yaml is up-to-date.
func (c *client) CheckProviderTemplatesNeedUpdate() (bool, error) {
	// Do not update provider templates if `SUPPRESS_PROVIDERS_UPDATE` env variable is set
	if isSuppressProviderUpdateEnvSet() {
		return false, nil
	}

	// If local develeopment and providers are embedded then always update providers based
	if c.isProviderTemplatesEmbedded() {
		return true, nil
	}

	providerDir, err := c.tkgConfigPathsClient.GetTKGProvidersDirectory()
	if err != nil {
		return true, err
	}

	// check the version info with BoM file's tag
	// if it matches no need to update anything
	// if not providers need to be updated

	tkgBomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return true, errors.Wrap(err, "error reading TKG BoM configuration")
	}

	providerTemplateImage, err := getProviderTemplateImageFromBoM(tkgBomConfig)
	if err != nil {
		return true, err
	}

	imageTag := providerTemplateImage.Tag
	if _, err := os.Stat(filepath.Join(providerDir, imageTag)); os.IsNotExist(err) {
		return true, nil
	}

	return false, nil
}

// CheckBOMsNeedUpdate checks if bom files are up-to-date.
// returns true if $HOME/.tkg/bom directory exists, not empty and doesn't contain the defaultBoM file
func (c *client) CheckBOMsNeedUpdate() (bool, error) {
	var err error

	compatibilityFileExists, err := c.checkTKGCompatibilityFileExists()
	if err != nil {
		return false, err
	}
	// tkg-compatibility file contents determines default BOM files, if compatibility file doesn't exists it should trigger BOM file updates
	if !compatibilityFileExists {
		return true, nil
	}

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

func getProviderTemplateImageFromBoM(tkgBomConfig *tkgconfigbom.BOMConfiguration) (*tkgconfigbom.ImageInfo, error) {
	if _, exists := tkgBomConfig.Components["tanzu-framework"]; !exists {
		return nil, errors.New("unable to find tanzu-framework component in TKG BoM file")
	}

	if _, exists := tkgBomConfig.Components["tanzu-framework"][0].Images["providerTemplateImage"]; !exists {
		return nil, errors.New("unable to find providerTemplateImage in TKG BoM file")
	}

	return tkgBomConfig.Components["tanzu-framework"][0].Images["providerTemplateImage"], nil
}

func isSuppressProviderUpdateEnvSet() bool {
	return os.Getenv(constants.SuppressProvidersUpdate) != ""
}

// checkTKGCompatibilityFileExists checks if TKG compatibility file exists.
// returns true if <TKGConfigDir>/compatibility directory exists and contains the TKG Compatibility file
func (c *client) checkTKGCompatibilityFileExists() (bool, error) {
	var err error
	compatibilityConfigFile, err := c.tkgConfigPathsClient.GetTKGCompatibilityConfigPath()
	if err != nil {
		return false, errors.Wrap(err, "unable to get TKG compatibility file path")
	}
	_, err = os.Stat(compatibilityConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		// for any other error return error
		return false, errors.Wrap(err, "failed to check TKG compatibility file exists")
	}
	return true, nil
}
