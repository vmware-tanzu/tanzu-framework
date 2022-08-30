// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
)

func (c *client) GetAzureVMImageInfo(tkrVersion string) (*tkgconfigbom.AzureInfo, error) {
	imageInfo, err := c.getAzureVMImageInfoFromUserConfiguration(tkrVersion)
	if err == nil && imageInfo != nil {
		return imageInfo, nil
	}
	log.V(9).Infof("unable to find azure-image config in user configuration file, using it from BoM files")

	bomConfiguration, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get bom configuration for TKr version %s", tkrVersion)
	}

	imageInfo = tkgconfighelper.SelectAzureImageBasedonOSOptions(bomConfiguration.Azure, c.TKGConfigReaderWriter())
	if imageInfo == nil {
		osInfo := tkgconfighelper.GetUserProvidedOsOptions(c.TKGConfigReaderWriter())
		return nil, errors.Errorf("unable to find azure-image for TKr version: %v, os options: %v in BoM files", tkrVersion, osInfo)
	}

	return imageInfo, nil
}

func (c *client) getAzureVMImageInfoFromUserConfiguration(tkrVersion string) (*tkgconfigbom.AzureInfo, error) {
	var azureImageMap map[string][]tkgconfigbom.AzureInfo
	err := c.TKGConfigReaderWriter().UnmarshalKey("azure-image", &azureImageMap)
	if err != nil {
		return nil, errors.New("unable to read azure-image config from user configuration file")
	}

	azureImages, exists := azureImageMap[tkrVersion]
	if exists && len(azureImages) != 0 {
		userConfiguredAzureImageInfo := tkgconfighelper.SelectAzureImageBasedonOSOptions(azureImages, c.TKGConfigReaderWriter())
		if userConfiguredAzureImageInfo != nil {
			log.V(3).Infof("using Azure Image based on the user settings, TKr version: '%v', azureImageInfo: '%v'", tkrVersion, userConfiguredAzureImageInfo)
			return userConfiguredAzureImageInfo, nil
		}
		return nil, errors.Errorf("unable to configure azure image information from user configuration file")
	}
	return nil, errors.New("no azure images found for kubernetes version: %v in user configure configuration file")
}
