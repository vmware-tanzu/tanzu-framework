// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Release contains the release name.
type Release struct {
	Version string `yaml:"version"`
}

// ImageConfig contains the path of the image registry
type ImageConfig struct {
	ImageRepository string `yaml:"imageRepository"`
}

// ComponentInfo contains component information
type ComponentInfo struct {
	Version string `yaml:"version"`
	// Each component can optionally have container images associated with it
	Images map[string]ImageInfo `yaml:"images,omitempty"`
	// Metadata section can be anything for the component
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

// ImageInfo contains the image information
type ImageInfo struct {
	ImagePath string `yaml:"imagePath"`
	Tag       string `yaml:"tag"`
}

// OSInfo contains the OS information
type OSInfo struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Arch    string `yaml:"arch"`
}

// String gets a string representation of the BOM
func (os OSInfo) String() string {
	return fmt.Sprintf("%s-%s-%s", os.Name, os.Version, os.Arch)
}

// OVAInfo gets the Open Virtual Appliance (OVA) information
type OVAInfo struct {
	Name string `yaml:"name"`
	OSInfo
	Version  string                 `yaml:"version"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// AMIInfo defines information about an AMI shipped
type AMIInfo struct {
	ID string `yaml:"id"`
	OSInfo
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// AzureInfo defines information about an Azure Image shipped
type AzureInfo struct {
	Sku             string                 `yaml:"sku"`
	Publisher       string                 `yaml:"publisher"`
	Offer           string                 `yaml:"offer"`
	Version         string                 `yaml:"version"`
	ThirdPartyImage bool                   `yaml:"thirdPartyImage,omitempty"`
	Metadata        map[string]interface{} `yaml:"metadata,omitempty"`
	OSInfo
}

// AzureInfos is an array of AzureInfo
type AzureInfos []AzureInfo

// AMIInfos is a lookup of AMIInfo
type AMIInfos map[string][]AMIInfo

// OVAInfos is an array of OVAInfo
type OVAInfos []OVAInfo

// Addons represents map of Addons
type Addons map[string]Addon

// ComponentReference references the images in a component
type ComponentReference struct {
	ComponentRef string   `yaml:"componentRef"`
	ImageRefs    []string `yaml:"imageRefs"`
}

// Addon contains addon info
type Addon struct {
	Category             string               `yaml:"category,omitempty"`
	ClusterTypes         []string             `yaml:"clusterTypes,omitempty"`
	Version              string               `yaml:"version,omitempty"`
	TemplatesImagePath   string               `yaml:"templatesImagePath,omitempty"`
	TemplatesImageTag    string               `yaml:"templatesImageTag,omitempty"`
	AddonTemplatesImage  []ComponentReference `yaml:"addonTemplatesImage,omitempty"`
	AddonContainerImages []ComponentReference `yaml:"addonContainerImages,omitempty"`
	PackageName          string               `yaml:"packageName,omitempty"`
}

// BomContent contains the content of a BOM file
type BomContent struct {
	Release     Release                    `yaml:"release"`
	Components  map[string][]ComponentInfo `yaml:"components"`
	ImageConfig ImageConfig                `yaml:"imageConfig"`
	OVA         OVAInfos                   `yaml:"ova"`
	AMI         AMIInfos                   `yaml:"ami,omitempty"`
	Azure       AzureInfos                 `yaml:"azure,omitempty"`
	Addons      Addons                     `yaml:"addons,omitempty"`
}

// Bom represents a BOM file
type Bom struct {
	bom        BomContent
	initialzed bool
}

// NewBom creates a new Bom from raw data
func NewBom(content []byte) (Bom, error) {
	var bc BomContent
	err := yaml.Unmarshal(content, &bc)
	if err != nil {
		return Bom{}, errors.Wrap(err, "error parsing the BOM file content")
	}

	if bc.Release.Version == "" {
		return Bom{}, errors.New("bom does not contain proper release information")
	}
	if len(bc.Components) == 0 {
		return Bom{}, errors.New("bom does not contain release component information")
	}

	if bc.ImageConfig.ImageRepository == "" {
		return Bom{}, errors.New("bom does not contain image repository information")
	}

	return Bom{
		bom:        bc,
		initialzed: true,
	}, nil
}

// GetReleaseVersion gets the Tanzu release version
func (b *Bom) GetReleaseVersion() (string, error) {
	if !b.initialzed {
		return "", errors.New("the BOM is not initialized")
	}
	return b.bom.Release.Version, nil
}

// GetComponent gets a release component
func (b *Bom) GetComponent(name string) ([]ComponentInfo, error) {
	if !b.initialzed {
		return []ComponentInfo{}, errors.New("the BOM is not initialized")
	}
	if component, ok := b.bom.Components[name]; ok {
		return component, nil
	}
	return []ComponentInfo{}, errors.Errorf("unable to find the component %s", name)
}

// GetImageInfo gets a image in a component of specific version.
// If version is "" or not found, get the image of the first version in the component array
func (b *Bom) GetImageInfo(componentName, componentVersion, imageName string) (ImageInfo, error) {
	if !b.initialzed {
		return ImageInfo{}, errors.New("the BOM is not initialized")
	}
	components, ok := b.bom.Components[componentName]
	if !ok {
		return ImageInfo{}, errors.Errorf("unable to find the component %s", componentName)
	}
	if len(components) < 1 {
		return ImageInfo{}, errors.Errorf("Empty component list in BOM for %s", componentName)
	}

	versionIndex := 0
	if componentVersion != "" {
		for idx, component := range components {
			if component.Version == componentVersion {
				versionIndex = idx
				break
			}
		}
	}
	if image, ok := components[versionIndex].Images[imageName]; ok {
		return image, nil
	}
	return ImageInfo{}, errors.Errorf("unable to find image %s in component %s", imageName, componentName)
}

// Components gets all release components in the BOM file
func (b *Bom) Components() (map[string][]ComponentInfo, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}
	return b.bom.Components, nil
}

// GetImageRepository gets the image repository
func (b *Bom) GetImageRepository() (string, error) {
	if !b.initialzed {
		return "", errors.New("the BOM is not initialized")
	}
	return b.bom.ImageConfig.ImageRepository, nil
}

// Addons gets all the addons in the BOM
func (b *Bom) Addons() (Addons, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}
	return b.bom.Addons, nil
}

// GetAddon gets an addon info from BOM
func (b *Bom) GetAddon(name string) (Addon, error) {
	if !b.initialzed {
		return Addon{}, errors.New("the BOM is not initialized")
	}

	if addon, ok := b.bom.Addons[name]; ok {
		return addon, nil
	}

	return Addon{}, errors.Errorf("unable to find the Addon %s", name)
}

// GetAzureInfo gets azure os image info
func (b *Bom) GetAzureInfo() ([]AzureInfo, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}
	return b.bom.Azure, nil
}

// GetAMIInfo gets ami info
func (b *Bom) GetAMIInfo() (map[string][]AMIInfo, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}
	return b.bom.AMI, nil
}

// GetOVAInfo gets vsphere ova info
func (b *Bom) GetOVAInfo() ([]OVAInfo, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}
	return b.bom.OVA, nil
}
