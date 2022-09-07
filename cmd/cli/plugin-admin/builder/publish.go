// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/command"
)

var publishArgs = &command.PublishArgs{
	OSArch: common.DefaultOSArch,
}

// NewPublishCmd creates a new command for plugin publishing.
func NewPublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.PublishPlugins(publishArgs)
		},
	}

	cmd.Flags().StringVar(&publishArgs.DistroType, "type", "", "Type of discovery and distribution for publishing plugins. Supported: local, oci")
	cmd.Flags().StringVar(&publishArgs.PluginsString, "plugins", "", "List of plugin names. Example: 'login management-cluster cluster'")
	cmd.Flags().StringVar(&publishArgs.InputArtifactDir, "input-artifact-dir", "", "Artifact directory which is a output of 'tanzu builder cli compile' command.")

	cmd.Flags().StringVar(&publishArgs.OSArch, "os-arch", publishArgs.OSArch, "List of OS architectures.")
	cmd.Flags().StringVar(&publishArgs.RecommendedVersion, "version", "", "Recommended version of the plugins.")

	cmd.Flags().StringVar(&publishArgs.LocalOutputDiscoveryDir, "local-output-discovery-dir", "", "Local output directory where CLIPlugin resource yamls for discovery will be placed. Applicable to 'local' type.")
	cmd.Flags().StringVar(&publishArgs.LocalOutputDistribtionDir, "local-output-distribution-dir", "", "Local output directory where plugin binaries will be placed. Applicable to 'local' type.")

	cmd.Flags().StringVar(&publishArgs.OCIDiscoverImage, "oci-discovery-image", "", "Image path to publish oci image with CLIPlugin resource yamls. Applicable to 'oci' type.")
	cmd.Flags().StringVar(&publishArgs.OCIDistributionImageRepository, "oci-distribution-image-repository", "", "Image path prefix to publish oci image for plugin binaries. Applicable to 'oci' type.")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("plugins")
	_ = cmd.MarkFlagRequired("input-artifact-dir")

	return cmd
}
