// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

type imageForConfigFile struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag,omitempty"`
}

func (c *client) EnsureImages(needUpdate bool, tkgConfigNode *yaml.Node) error {
	imageIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)

	// Always update the image section to handle users using different BOM files. //-x clean up?
	// if imageIndex != -1 && !needUpdate {
	// 	return nil
	// }

	if imageIndex == -1 {
		tkgConfigNode.Content[0].Content = append(tkgConfigNode.Content[0].Content, createMappingNode(constants.ImagesConfigKey)...)
		imageIndex = GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
	}

	images, err := c.getImageMapForConfigFile()
	if err != nil {
		return errors.Wrap(err, "unable to get image map for config file")
	}

	defaultImagesBytes, err := yaml.Marshal(images)
	if err != nil {
		return errors.Wrap(err, "unable to get a list of default images")
	}

	imageListNode := yaml.Node{}
	err = yaml.Unmarshal(defaultImagesBytes, &imageListNode)
	if err != nil {
		return errors.Wrap(err, "unable to get a list of default images")
	}

	tkgConfigNode.Content[0].Content[imageIndex] = imageListNode.Content[0]
	return nil
}

func (c *client) getImageMapForConfigFile() (map[string]imageForConfigFile, error) {
	bomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get default BOM file")
	}

	baseImageRepository := bomConfig.ImageConfig.ImageRepository

	type component struct {
		ComponentName string
		ImageName     string
	}
	configKeyTObomImageKeyMap := map[string]component{
		"cert-manager": {
			"jetstack_cert-manager",
			"certMgrControllerImage",
		},
		"cluster-api": {
			"cluster_api",
			"capiControllerImage",
		},
		"bootstrap-kubeadm": {
			"cluster_api",
			"cabpkControllerImage",
		},
		"control-plane-kubeadm": {
			"cluster_api",
			"kcpControllerImage",
		},
		"infrastructure-aws": {
			"cluster_api_aws",
			"capaControllerImage",
		},
		"infrastructure-vsphere": {
			"cluster_api_vsphere",
			"capvControllerImage",
		},
		"infrastructure-docker": {
			"cluster_api",
			"capdManagerImage",
		},
		"infrastructure-azure": {
			"cluster-api-provider-azure",
			"capzControllerImage",
		},
	}

	defaultBaseRepositoryForClusterAPI := baseImageRepository + "/cluster-api"

	images := map[string]imageForConfigFile{}
	images["all"] = imageForConfigFile{Repository: defaultBaseRepositoryForClusterAPI}

	for configKey, bomImageComponent := range configKeyTObomImageKeyMap {
		imageInfoFromBOM, exists := bomConfig.Components[bomImageComponent.ComponentName][0].Images[bomImageComponent.ImageName]
		if !exists {
			log.V(7).Infof("unable to find component %s, image %s in BOM file", bomImageComponent.ComponentName, bomImageComponent.ImageName)
			continue
		}

		repository, err := getRepositoryFromImagePathWithBaseRepositoryPatch(imageInfoFromBOM, baseImageRepository)
		if err != nil {
			log.V(7).Infof("unable to construct repository information for component %s, image %s in BOM file", bomImageComponent.ComponentName, bomImageComponent.ImageName)
			continue
		}

		if repository == defaultBaseRepositoryForClusterAPI {
			continue
		}

		image := imageForConfigFile{}
		image.Repository = repository
		if configKey == "cert-manager" {
			image.Tag = imageInfoFromBOM.Tag
		}
		images[configKey] = image
	}

	return images, nil
}

// getRepositoryFromImagePathWithBaseRepositoryPatch return repository string from imagePath
// example: imagePath=cert-manager/cert-manager-controller, imageRepo: custom-repository, baseImageRepository: registry.tkg.vmware.run
// returns: custom-repository/cert-manager
//
// example: imagePath=cert-manager/cert-manager-controller, imageRepo: "", baseImageRepository: registry.tkg.vmware.run
// returns: registry.tkg.vmware.run/cert-manager
func getRepositoryFromImagePathWithBaseRepositoryPatch(image *tkgconfigbom.ImageInfo, baseImageRepository string) (string, error) {
	imagePath := tkgconfigbom.GetFullImagePath(image, baseImageRepository)
	arr := strings.Split(imagePath, "/")
	if len(arr) < 1 {
		return "", errors.New("invalid image patch, path does not contain repository")
	}
	return strings.Join(arr[:len(arr)-1], "/"), nil
}
