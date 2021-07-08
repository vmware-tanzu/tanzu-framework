// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfighelper

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
)

const (
	osName    = "name"
	osVersion = "version"
	osArch    = "arch"
)

// GetDefaultOsOptions returns default OS option based on providerType
func GetDefaultOsOptions(providerType string) tkgconfigbom.OSInfo {
	switch providerType {
	case constants.InfrastructureProviderVSphere:
		return tkgconfigbom.OSInfo{Name: "ubuntu", Version: "20.04", Arch: "amd64"}

	case constants.InfrastructureProviderAWS:
		return tkgconfigbom.OSInfo{Name: "ubuntu", Version: "20.04", Arch: "amd64"}

	case constants.InfrastructureProviderAzure:
		return tkgconfigbom.OSInfo{Name: "ubuntu", Version: "20.04", Arch: "amd64"}
	}
	return tkgconfigbom.OSInfo{}
}

// GetOSOptionsForProviders returns OS options for the providers
// If user has configured any options, it will have higher precedence
// user provided settings gets merged with default OS options for the given provider
func GetOSOptionsForProviders(providerType string, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) tkgconfigbom.OSInfo {
	osInfo := GetDefaultOsOptions(providerType)

	osName, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableOSName)
	osVersion, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableOSVersion)
	osArch, _ := tkgConfigReaderWriter.Get(constants.ConfigVariableOSArch)

	if osName != "" {
		osInfo.Name = osName
	}
	if osVersion != "" {
		osInfo.Version = osVersion
	}
	if osArch != "" {
		osInfo.Arch = osArch
	}
	return osInfo
}

// GetUserProvidedOsOptions returns user provided os options
func GetUserProvidedOsOptions(tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) tkgconfigbom.OSInfo {
	osInfo := tkgconfigbom.OSInfo{}

	osInfo.Name, _ = tkgConfigReaderWriter.Get(constants.ConfigVariableOSName)
	osInfo.Version, _ = tkgConfigReaderWriter.Get(constants.ConfigVariableOSVersion)
	osInfo.Arch, _ = tkgConfigReaderWriter.Get(constants.ConfigVariableOSArch)

	return osInfo
}

// ############################### vSphere specific helper functions ###############################

// SelectTemplateForVsphereProviderBasedonOSOptions selects template among all for vsphere provider
func SelectTemplateForVsphereProviderBasedonOSOptions(vms []*types.VSphereVirtualMachine, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) *types.VSphereVirtualMachine {
	if len(vms) == 0 { // no vms provided skipping selection
		return nil
	}

	userProvidedOSOptions := GetUserProvidedOsOptions(tkgConfigReaderWriter)

	if userProvidedOSOptions.Name != "" {
		vms = filterVMTemplatesByOSOption(vms, osName, userProvidedOSOptions.Name)
		if userProvidedOSOptions.Version != "" {
			vms = filterVMTemplatesByOSOption(vms, osVersion, userProvidedOSOptions.Version)
			if userProvidedOSOptions.Arch != "" {
				vms = filterVMTemplatesByOSOption(vms, osArch, userProvidedOSOptions.Arch)
			}
		}
		if len(vms) == 0 {
			return nil
		}
		if len(vms) == 1 {
			return vms[0]
		}
		if len(vms) > 1 {
			log.V(6).Infof("multiple vm template found for given kubernetes version and OS option: %v, selecting %v", userProvidedOSOptions, vms[0].Name)
			return vms[0]
		}
	}

	log.V(6).Info("no os options provided, selecting based on default os options")

	if len(vms) == 1 { // only one vm found, return that vm
		return vms[0]
	}

	vm := filterVMTemplatesByAllOsOptions(vms, GetDefaultOsOptions(constants.InfrastructureProviderVSphere))
	if vm != nil {
		return vm
	}

	// TODO: remove this and return nil
	// currently added this line because we do not have v1.19.3 vm templates with OS information present under vAppConfig
	log.Infof("No vm template found matching user provided os configuration or default os configuration, selecting '%v'", vms[0].Name)
	return vms[0]
}

func filterVMTemplatesByOSOption(vms []*types.VSphereVirtualMachine, osOption, osOptionValue string) []*types.VSphereVirtualMachine {
	filteredVms := []*types.VSphereVirtualMachine{}
	for _, vm := range vms {
		switch osOption {
		case osName:
			if vm.DistroName == osOptionValue {
				filteredVms = append(filteredVms, vm)
			}
		case osVersion:
			if vm.DistroVersion == osOptionValue {
				filteredVms = append(filteredVms, vm)
			}
		case osArch:
			if vm.DistroArch == osOptionValue {
				filteredVms = append(filteredVms, vm)
			}
		}
	}
	return filteredVms
}

func filterVMTemplatesByAllOsOptions(vms []*types.VSphereVirtualMachine, osOption tkgconfigbom.OSInfo) *types.VSphereVirtualMachine {
	for _, vm := range vms {
		if vm.DistroName == osOption.Name && vm.DistroVersion == osOption.Version && vm.DistroArch == osOption.Arch {
			return vm
		}
	}
	return nil
}

// ############################### Azure specific helper functions ###############################

// SelectAzureImageBasedonOSOptions selects template among all for azure images
func SelectAzureImageBasedonOSOptions(azureImages []tkgconfigbom.AzureInfo, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) *tkgconfigbom.AzureInfo { // nolint:dupl
	if len(azureImages) == 0 { // no image provided skipping selection
		return nil
	}

	userProvidedOSOptions := GetUserProvidedOsOptions(tkgConfigReaderWriter)

	if userProvidedOSOptions.Name != "" {
		azureImages = filterAzureImagesByOSOption(azureImages, osName, userProvidedOSOptions.Name)
		if userProvidedOSOptions.Version != "" {
			azureImages = filterAzureImagesByOSOption(azureImages, osVersion, userProvidedOSOptions.Version)
			if userProvidedOSOptions.Arch != "" {
				azureImages = filterAzureImagesByOSOption(azureImages, osArch, userProvidedOSOptions.Arch)
			}
		}
		if len(azureImages) == 0 {
			return nil
		}
		if len(azureImages) == 1 {
			return &azureImages[0]
		}
		if len(azureImages) > 1 {
			log.V(9).Infof("multiple azure images found for given kubernetes version and OS option: %v, selecting %v", userProvidedOSOptions, azureImages[0])
			return &azureImages[0]
		}
	}

	log.V(6).Info("no os options provided, selecting based on default os options")

	if len(azureImages) == 1 { // only one image found, return that vm
		return &azureImages[0]
	}

	azureImage := filterAzureImagesByAllOsOptions(azureImages, GetDefaultOsOptions(constants.InfrastructureProviderAzure))
	if azureImage != nil {
		return azureImage
	}

	return nil
}

func filterAzureImagesByOSOption(azureImages []tkgconfigbom.AzureInfo, osOption, osOptionValue string) []tkgconfigbom.AzureInfo {
	filteredAzureImages := []tkgconfigbom.AzureInfo{}
	for i := range azureImages {
		switch osOption {
		case osName:
			if azureImages[i].OSInfo.Name == osOptionValue {
				filteredAzureImages = append(filteredAzureImages, azureImages[i])
			}
		case osVersion:
			if azureImages[i].OSInfo.Version == osOptionValue {
				filteredAzureImages = append(filteredAzureImages, azureImages[i])
			}
		case osArch:
			if azureImages[i].OSInfo.Arch == osOptionValue {
				filteredAzureImages = append(filteredAzureImages, azureImages[i])
			}
		}
	}
	return filteredAzureImages
}

func filterAzureImagesByAllOsOptions(azureImages []tkgconfigbom.AzureInfo, osOption tkgconfigbom.OSInfo) *tkgconfigbom.AzureInfo {
	for i := range azureImages {
		if azureImages[i].OSInfo.Name == osOption.Name && azureImages[i].OSInfo.Version == osOption.Version && azureImages[i].OSInfo.Arch == osOption.Arch {
			return &azureImages[i]
		}
	}
	return nil
}

// ############################### AWS specific helper functions ###############################

// SelectAWSImageBasedonOSOptions selects template among all for azure images
func SelectAWSImageBasedonOSOptions(amis []tkgconfigbom.AMIInfo, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) *tkgconfigbom.AMIInfo { //nolint:dupl
	if len(amis) == 0 { // no image provided skipping selection
		return nil
	}

	userProvidedOSOptions := GetUserProvidedOsOptions(tkgConfigReaderWriter)

	if userProvidedOSOptions.Name != "" {
		amis = filterAMIsByOSOption(amis, osName, userProvidedOSOptions.Name)
		if userProvidedOSOptions.Version != "" {
			amis = filterAMIsByOSOption(amis, osVersion, userProvidedOSOptions.Version)
			if userProvidedOSOptions.Arch != "" {
				amis = filterAMIsByOSOption(amis, osArch, userProvidedOSOptions.Arch)
			}
		}
		if len(amis) == 0 {
			return nil
		}
		if len(amis) == 1 {
			return &amis[0]
		}
		if len(amis) > 1 {
			log.V(6).Infof("multiple aws images found for given kubernetes version and OS option: %v, selecting %v", userProvidedOSOptions, amis[0])
			return &amis[0]
		}
	}

	log.V(6).Info("no os options provided, selecting based on default os options")

	if len(amis) == 1 { // only one image found, return that vm
		return &amis[0]
	}

	ami := filterAMIsByAllOsOptions(amis, GetDefaultOsOptions(constants.InfrastructureProviderAWS))
	if ami != nil {
		return ami
	}

	return nil
}

func filterAMIsByOSOption(amis []tkgconfigbom.AMIInfo, osOption, osOptionValue string) []tkgconfigbom.AMIInfo {
	filteredAMIs := []tkgconfigbom.AMIInfo{}
	for _, ami := range amis {
		switch osOption {
		case osName:
			if ami.OSInfo.Name == osOptionValue {
				filteredAMIs = append(filteredAMIs, ami)
			}
		case osVersion:
			if ami.OSInfo.Version == osOptionValue {
				filteredAMIs = append(filteredAMIs, ami)
			}
		case osArch:
			if ami.OSInfo.Arch == osOptionValue {
				filteredAMIs = append(filteredAMIs, ami)
			}
		}
	}
	return filteredAMIs
}

func filterAMIsByAllOsOptions(amis []tkgconfigbom.AMIInfo, osOption tkgconfigbom.OSInfo) *tkgconfigbom.AMIInfo {
	for _, ami := range amis {
		if ami.OSInfo.Name == osOption.Name && ami.OSInfo.Version == osOption.Version && ami.OSInfo.Arch == osOption.Arch {
			return &ami
		}
	}
	return nil
}

// GetDefaultOsOptionsForTKG12 returns default OS option based on providerType
func GetDefaultOsOptionsForTKG12(providerType string) tkgconfigbom.OSInfo {
	switch providerType {
	case constants.InfrastructureProviderVSphere:
		return tkgconfigbom.OSInfo{Name: "photon", Version: "3", Arch: "amd64"}

	case constants.InfrastructureProviderAWS:
		return tkgconfigbom.OSInfo{Name: "amazon", Version: "2", Arch: "amd64"}

	case constants.InfrastructureProviderAzure:
		return tkgconfigbom.OSInfo{Name: "ubuntu", Version: "18.04", Arch: "amd64"}
	}
	return tkgconfigbom.OSInfo{}
}
