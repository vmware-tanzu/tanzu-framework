// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/builder/command/publish"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

var (
	distroType, pluginsString, oa, inputArtifactDir     string
	localOutputDiscoveryDir, localOutputDistributionDir string
	ociDiscoveryImage, ociDistributionImageRepository   string
	recommendedVersion                                  string
)

// PublishCmd publishes plugin resources
var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish operations",
	RunE:  publishPlugins,
}

func init() {
	PublishCmd.Flags().StringVar(&distroType, "type", "", "type of discovery and distribution for publishing plugins. Supported: local")
	PublishCmd.Flags().StringVar(&pluginsString, "plugins", "", "list of plugin names. Example: 'login management-cluster cluster'")
	PublishCmd.Flags().StringVar(&inputArtifactDir, "input-artifact-dir", "", "artifact directory which is a output of 'tanzu builder cli compile' command")

	PublishCmd.Flags().StringVar(&oa, "os-arch", common.DefaultOSArch, "list of os-arch")
	PublishCmd.Flags().StringVar(&recommendedVersion, "version", "", "recommended version of the plugins")

	PublishCmd.Flags().StringVar(&localOutputDiscoveryDir, "local-output-discovery-dir", "", "local output directory where CLIPlugin resource yamls for discovery will be placed. Applicable to 'local' type")
	PublishCmd.Flags().StringVar(&localOutputDistributionDir, "local-output-distribution-dir", "", "local output directory where plugin binaries will be placed. Applicable to 'local' type")

	PublishCmd.Flags().StringVar(&ociDiscoveryImage, "oci-discovery-image", "", "image path to publish oci image with CLIPlugin resource yamls. Applicable to 'oci' type")
	PublishCmd.Flags().StringVar(&ociDistributionImageRepository, "oci-distribution-image-repository", "", "image path prefix to publish oci image for plugin binaries. Applicable to 'oci' type")

	_ = PublishCmd.MarkFlagRequired("type")
	_ = PublishCmd.MarkFlagRequired("version")
	_ = PublishCmd.MarkFlagRequired("plugins")
	_ = PublishCmd.MarkFlagRequired("input-artifact-dir")
}

func publishPlugins(cmd *cobra.Command, args []string) error {
	plugins := strings.Split(pluginsString, " ")
	osArch := strings.Split(oa, " ")

	if localOutputDiscoveryDir == "" {
		localOutputDiscoveryDir = filepath.Join(common.DefaultLocalPluginDistroDir, "discovery", "oci")
	}

	var publisherInterface publish.Publisher
	var err error

	switch strings.ToLower(distroType) {
	case "local":
		publisherInterface, err = publish.NewLocalPublisher(localOutputDistributionDir)
	case "oci":
		publisherInterface, err = publish.NewOCIPublisher(ociDiscoveryImage, ociDistributionImageRepository, localOutputDiscoveryDir)
	default:
		return errors.Errorf("publish plugins with type %s is not yet supported", distroType)
	}
	if err != nil {
		return err
	}

	publishMetadata := publish.Metadata{
		Plugins:            plugins,
		OSArch:             osArch,
		RecommendedVersion: recommendedVersion,
		InputArtifactDir:   inputArtifactDir,
		LocalDiscoveryPath: localOutputDiscoveryDir,
		PublisherInterface: publisherInterface,
	}

	return publish.PublishPlugins(&publishMetadata)
}
