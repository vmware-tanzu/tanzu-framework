// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/cmd"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepares your environment for an offline installation",
	Long: cmd.LongDesc(`
This command prepares your environment for an internet-restricted or air-gapped
installation of Tanzu Kubernetes Grid. It must be run from a machine with internet
access.

This includes:

	- Archiving Tanzu CLI metadata (e.g. plugins and bill-of-materials files), and
	- Uploading public images into your private registry.`),
	Example: `
	TKG_OFFLINE_REGISTRY_USERNAME=registry-username \
	TKG_OFFLINE_REGISTRY_PASSWORD=registry-password \
	tanzu offline prepare --container-registry registry.example.com:7999
	# Uploads public TKG images into registry.example.com:7999 and archives
	# your Tanzu CLI configuration into $XDG_DATA_DIR/tanzu-offline.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return prepare()
	},
}

func prepare() error {
	// WIP.
	return nil
}
