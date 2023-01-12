// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package imgpkgutil

import (
	"bytes"
	"strings"

	"sigs.k8s.io/yaml"
)

type imagesLock struct {
	Images []*image `json:"images"`
}

type image struct {
	Image       string            `json:"image"`
	Annotations map[string]string `json:"annotations"`
}

const annotKbldID = "kbld.carvel.dev/id"

func ParseImagesLock(bundleImageName string, imagesLockBytes []byte) (map[string]string, error) {
	if imagesLockBytes == nil {
		return nil, nil
	}

	bundleImageNamePrefix := bundleImagePrefix(bundleImageName)

	var imagesLock imagesLock
	if err := yaml.Unmarshal(imagesLockBytes, &imagesLock); err != nil {
		return nil, err
	}

	imageMap := make(map[string]string, len(imagesLock.Images))
	for _, image := range imagesLock.Images {
		targetImage := replaceImagePrefix(bundleImageNamePrefix, image.Image)
		imageMap[image.Annotations[annotKbldID]] = targetImage
	}
	return imageMap, nil
}

func replaceImagePrefix(imagePrefix string, image string) string {
	shaTagIndex := strings.Index(image, "@sha256:")
	return imagePrefix + image[shaTagIndex:]
}

func bundleImagePrefix(bundleImageName string) string {
	lastIndex := strings.LastIndex(bundleImageName, ":")
	if lastIndex == -1 {
		return bundleImageName
	}
	return bundleImageName[:lastIndex]
}

func ResolveImages(imageMap map[string]string, bundleContent map[string][]byte) {
	for origImage, targetImage := range imageMap {
		for path, bs := range bundleContent {
			bundleContent[path] = bytes.ReplaceAll(bs, []byte(origImage), []byte(targetImage))
		}
	}
}
