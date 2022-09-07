// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/command/publish"
)

type PublishArgs struct {
	DistroType                     string
	PluginsString                  string
	OSArch                         string
	InputArtifactDir               string
	LocalOutputDiscoveryDir        string
	LocalOutputDistribtionDir      string
	OCIDiscoverImage               string
	OCIDistributionImageRepository string
	RecommendedVersion             string
}

func PublishPlugins(publishArgs *PublishArgs) error {
	plugins := strings.Split(publishArgs.PluginsString, " ")
	osArch := strings.Split(publishArgs.OSArch, " ")

	if publishArgs.LocalOutputDiscoveryDir == "" {
		publishArgs.LocalOutputDiscoveryDir = filepath.Join(common.DefaultLocalPluginDistroDir, "discovery", "oci")
	}

	var publisherInterface publish.Publisher
	var err error

	switch strings.ToLower(publishArgs.DistroType) {
	case "local":
		publisherInterface, err = publish.NewLocalPublisher(publishArgs.LocalOutputDistribtionDir)
	case "oci":
		publisherInterface, err = publish.NewOCIPublisher(publishArgs.OCIDiscoverImage, publishArgs.OCIDistributionImageRepository, publishArgs.LocalOutputDiscoveryDir)
	default:
		return errors.Errorf("publish plugins with type %s is not yet supported", publishArgs.DistroType)
	}
	if err != nil {
		return err
	}

	publishMetadata := publish.Metadata{
		Plugins:            plugins,
		OSArch:             osArch,
		RecommendedVersion: publishArgs.RecommendedVersion,
		InputArtifactDir:   publishArgs.InputArtifactDir,
		LocalDiscoveryPath: publishArgs.LocalOutputDiscoveryDir,
		PublisherInterface: publisherInterface,
	}

	return publish.PublishPlugins(&publishMetadata)
}
