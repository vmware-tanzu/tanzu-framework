// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
)

func (c *client) GetAWSAMIInfo(tkrBoMConfiguration *tkgconfigbom.BOMConfiguration, awsRegion string) (*tkgconfigbom.AMIInfo, error) {
	userConfiguredAMIInfo := c.getUserConfiguredAMIBasedonOSOptions(tkrBoMConfiguration, awsRegion)
	if userConfiguredAMIInfo != nil {
		log.V(3).Infof("using AMI '%v' based on the user settings, aws-region '%v', amiInfo: %v", userConfiguredAMIInfo.ID, awsRegion, userConfiguredAMIInfo.ID)
		return userConfiguredAMIInfo, nil
	}

	// use new BoM file format to configure AMI ID  and OS options based on BOM and OS Options
	var mappedAMIs []tkgconfigbom.AMIInfo
	var exists bool
	// check if the AMI exists for given aws region
	if mappedAMIs, exists = tkrBoMConfiguration.AMI[awsRegion]; !exists {
		return nil, errors.Errorf("no AMI found in region %s for TKr version %s", awsRegion, tkrBoMConfiguration.Release.Version)
	}

	ami := tkgconfighelper.SelectAWSImageBasedonOSOptions(mappedAMIs, c.TKGConfigReaderWriter())
	if ami == nil {
		osInfo := tkgconfighelper.GetUserProvidedOsOptions(c.TKGConfigReaderWriter())
		return nil, errors.Errorf("no AMI found in region %s for TKr version: '%s' and os options: '(%v,%v,%v)'", awsRegion, tkrBoMConfiguration.Release.Version, osInfo.Name, osInfo.Version, osInfo.Arch)
	}

	return ami, nil
}

func (c *client) getUserConfiguredAMIBasedonOSOptions(tkrBoMConfiguration *tkgconfigbom.BOMConfiguration, awsRegion string) *tkgconfigbom.AMIInfo {
	var amiMapFromUserConfig map[string]tkgconfigbom.AWSVMImages
	err := c.TKGConfigReaderWriter().UnmarshalKey("aws-image", &amiMapFromUserConfig)
	if err != nil {
		return nil
	}

	mapRegionToAMIInfo, exists := amiMapFromUserConfig[tkrBoMConfiguration.Release.Version]
	if !exists {
		return nil
	}

	listAMIInfo, exists := mapRegionToAMIInfo[awsRegion]
	if !exists {
		return nil
	}

	// find match based on specified configVariable osName, osVersion and osArch
	// if no match found, return nil
	return tkgconfighelper.SelectAWSImageBasedonOSOptions(listAMIInfo, c.TKGConfigReaderWriter())
}
