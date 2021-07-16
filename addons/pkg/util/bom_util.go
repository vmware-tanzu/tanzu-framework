// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	bomtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
)

// GetBOMByTKRName returns the bom associated with the TKR
func GetBOMByTKRName(ctx context.Context, c client.Client, tkrName string) (*bomtypes.Bom, error) {
	configMapList := &corev1.ConfigMapList{}
	var bomConfigMap *corev1.ConfigMap
	if err := c.List(ctx, configMapList, client.InNamespace(constants.TKGBomNamespace), client.MatchingLabels{constants.TKRLabel: tkrName}); err != nil {
		return nil, err
	}

	if len(configMapList.Items) == 0 {
		return nil, nil
	}

	bomConfigMap = &configMapList.Items[0]
	bomData, ok := bomConfigMap.BinaryData[constants.TKGBomContent]
	if !ok {
		bomDataString, ok := bomConfigMap.Data[constants.TKGBomContent]
		if !ok {
			return nil, nil
		}
		bomData = []byte(bomDataString)
	}

	bom, err := bomtypes.NewBom(bomData)
	if err != nil {
		return nil, err
	}

	return &bom, nil
}

// GetTKRNameFromBOMConfigMap returns tkr name given a bom configmap
func GetTKRNameFromBOMConfigMap(bomConfigMap *corev1.ConfigMap) string {
	return bomConfigMap.Labels[constants.TKRLabel]
}

// GetAddonImageRepository returns imageRepository from configMap `tkr-controller-config` in namespace `tkr-system` if exists else use BOM
func GetAddonImageRepository(ctx context.Context, c client.Client, bom *bomtypes.Bom) (string, error) {
	bomConfigMap := &corev1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: constants.TKGBomNamespace,
		Name:      constants.TKRConfigmapName,
	}, bomConfigMap)

	// if the configmap exists, try get repository from the config map
	if err == nil {
		if imgRepo, ok := bomConfigMap.Data[constants.TKRRepoKey]; ok && imgRepo != "" {
			return imgRepo, nil
		}
	} else if !errors.IsNotFound(err) {
		// return the error to controller if err is not IsNotFound
		return "", err
	}
	// if the configmap doesn't exist, or there is no `imageRepository` field, get repository from BOM
	return bom.GetImageRepository()
}

// GetCorePackageRepositoryImageFromBom generates the core PackageRepository Object
func GetCorePackageRepositoryImageFromBom(bom *bomtypes.Bom) (*bomtypes.ImageInfo, error) {
	repositoryImage, err := bom.GetImageInfo(constants.TKGCorePackageRepositoryComponentName, "", constants.TKGCorePackageRepositoryImageName)
	if err != nil {
		return nil, err
	}
	return &repositoryImage, nil
}

// GetTemplateImageURLFromBom gets the image template image url of an addon
// This method first checks if packageName is present in addonConfig
// If packageName is present, it will use it to find imgpkg bundle in tkg-core-packages
// If packageName is not present, it will look for addonTemplatesImage
// addonTemplatesImage should be present in a 1.3.1 cluster, and will be used to find images in tanzu_core_addons
// If addonTemplatesImage is not present, it will fall back to using templatesImagePath and templatesImageTag to find template images
func GetTemplateImageURLFromBom(addonConfig *bomtypes.Addon, imageRepository string, bom *bomtypes.Bom) (string, error) {
	/*example addon section in BOM:
	  kapp-controller:
	    category: addons-management
	    clusterTypes:
	    - management
	    - workload
	    templatesImagePath: tanzu_core/addons/kapp-controller-templates (1.3.0-)
	    templatesImageTag: v1.3.0 (1.3.0-)
	    addonTemplatesImage: (1.3.1)
	    - componentRef: tanzu_core_addons (1.3.1)
	      imageRefs: (1.3.1)
	      - kappControllerTemplatesImage (1.3.1)
	    addonContainerImages: (1.3.1)
	    - componentRef: kapp-controller (1.3.1)
	      imageRefs: (1.3.1)
	      - kappControllerImage (1.3.1)
	    packageName: addons-manager.tkg-core.tanzu.vmware (1.4.0+)
	*/
	var templateImagePath, templateImageTag string
	if addonConfig.PackageName != "" {
		addonPackageImage, err := bom.GetImageInfo(constants.TKGCorePackageRepositoryComponentName, "", addonConfig.PackageName)
		if err != nil {
			return "", err
		}
		templateImagePath = addonPackageImage.ImagePath
		templateImageTag = addonPackageImage.Tag
	} else if len(addonConfig.AddonTemplatesImage) < 1 || len(addonConfig.AddonTemplatesImage[0].ImageRefs) < 1 {
		// if AddonTemplatesImage and AddonTemplatesImage are not present, use the older BOM format
		templateImagePath = addonConfig.TemplatesImagePath
		templateImageTag = addonConfig.TemplatesImageTag
	} else {
		templateImageComponentName := addonConfig.AddonTemplatesImage[0].ComponentRef
		templateImageName := addonConfig.AddonTemplatesImage[0].ImageRefs[0]

		templateImage, err := bom.GetImageInfo(templateImageComponentName, "", templateImageName)
		if err != nil {
			return "", err
		}
		templateImagePath = templateImage.ImagePath
		templateImageTag = templateImage.Tag
	}

	if templateImagePath == "" || templateImageTag == "" {
		err := fmt.Errorf("unable to get template image")
		return "", err
	}

	return fmt.Sprintf("%s/%s:%s", imageRepository, templateImagePath, templateImageTag), nil
}
