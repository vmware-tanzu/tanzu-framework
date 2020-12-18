package types

import (
"github.com/pkg/errors"
"gopkg.in/yaml.v2"
)

// Release contains the release name.
type Release struct {
	Version string `yaml:"version"`
}

// ImageConfig contains the path of the image registy
type ImageConfig struct {
	ImageRepository string `yaml:"imageRepository"`
}

// Image contains image info
type Image struct {
	ImagePath string `yaml:"imagePath"`
	Tag       string `yaml:"tag"`
}

// bomContent contains the content of a BOM file
type bomContent struct {
	TanzuRelease Release            `yaml:"release"`
	Components   map[string]Release `yaml:"components"`
	ImageConfig  ImageConfig        `yaml:"imageConfig"`
	Images       map[string]Image   `yaml:"images"`
	Addons       BomAddons          `yaml:"addons,omitempty"`
}

type BomAddons map[string]BomAddon

type BomAddon struct {
	Category      string   `yaml:"category,omitempty"`
	ClusterTypes  []string `yaml:"clusterTypes,omitempty"`
	Version       string   `yaml:"version,omitempty"`
	Image         string   `yaml:"image,omitempty"`
	ComponentName string   `yaml:"componentName,omitempty"`
}

// Bom represents a BOM file
type Bom struct {
	bom        bomContent
	initialzed bool
}

// NewBom creates a new Bom from raw data
func NewBom(content []byte) (Bom, error) {
	var bc bomContent
	err := yaml.Unmarshal(content, &bc)
	if err != nil {
		return Bom{}, errors.Wrap(err, "error parsing the BOM file content")
	}

	return Bom{
		bom:        bc,
		initialzed: true,
	}, nil
}

// GetTanzuReleaseVersion gets the Tanzu release version
func (b *Bom) GetTanzuReleaseVersion() (string, error) {
	if !b.initialzed {
		return "", errors.New("the BOM is not initialized")
	}
	return b.bom.TanzuRelease.Version, nil
}

// GetComponent gets a release component
func (b *Bom) GetComponent(name string) (Release, error) {
	if !b.initialzed {
		return Release{}, errors.New("the BOM is not initialized")
	}
	if release, ok := b.bom.Components[name]; ok {
		return release, nil
	}
	return Release{}, errors.Errorf("unable to find the component %s", name)
}

// GetImage gets a image
func (b *Bom) GetImage(name string) (Image, error) {
	if !b.initialzed {
		return Image{}, errors.New("the BOM is not initialized")
	}

	if image, ok := b.bom.Images[name]; ok {
		return image, nil
	}
	return Image{}, errors.Errorf("unable to find the Image %s", name)
}

// Images gets all images in the BOM file
func (b *Bom) Images() (map[string]Image, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}

	result := make(map[string]Image)
	for k, v := range b.bom.Images {
		result[k] = v
	}
	return result, nil
}

// Components gets all release components in the BOM file
func (b *Bom) Components() (map[string]Release, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}

	result := make(map[string]Release)
	for k, v := range b.bom.Components {
		result[k] = v
	}
	return result, nil
}

func (b *Bom) GetImageRepository() (string, error) {
	if !b.initialzed {
		return "", errors.New("the BOM is not initialized")
	}
	return b.bom.ImageConfig.ImageRepository, nil
}

func (b *Bom) Addons() (BomAddons, error) {
	if !b.initialzed {
		return nil, errors.New("the BOM is not initialized")
	}
	return b.bom.Addons, nil
}

func (b *Bom) GetAddon(name string) (BomAddon, error) {
	if !b.initialzed {
		return BomAddon{}, errors.New("the BOM is not initialized")
	}

	if addon, ok := b.bom.Addons[name]; ok {
		return addon, nil
	}

	return BomAddon{}, errors.Errorf("unable to find the Addon %s", name)
}
