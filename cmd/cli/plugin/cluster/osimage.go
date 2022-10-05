// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var osImageCmd = &cobra.Command{
	Use:   "osimage",
	Short: "Tanzu TKR image operations for Kubernetes nodes.",
	Long: "Tanzu TKR image operations for Kubernetes nodes. This command serves as a temporary workaround" +
		"for the Bring-Your-Own-Image feature. It imports the image from a public url, into the specified infrastructure" +
		"account / compartment. It also alters the TKR and adds OSImage that represented the imported image.",
}

var (
	// the input tkr package image path
	// with tag, for example projects-stg.registry.vmware.com/tkg/tkr-oci:v1.23.5
	tkrRegistryPath string
	// the input public image endpoint URI
	imageEndpoint string
	// the output tkr path
	outputDirectory string
	// display name for the imported image and the created OSImage resource
	name string

	// Image Operating System Information
	osType    string
	osName    string
	osVersion string
	osArch    string
)

func init() {
	osImageCmd.PersistentFlags().StringVar(&tkrRegistryPath, "tkr-path", "", "The input tkr package image path")
	osImageCmd.PersistentFlags().StringVar(&imageEndpoint, "image", "", "The input public image endpoint URI")
	osImageCmd.PersistentFlags().StringVarP(&outputDirectory, "output-directory", "d", "", "The output TKR path")

	osImageCmd.PersistentFlags().StringVar(&name, "name", "", "Display name for the imported image and the created OSImage resource")
	osImageCmd.PersistentFlags().StringVar(&osType, "os-type", "linux", "OS type")
	osImageCmd.PersistentFlags().StringVar(&osName, "os-name", "ubuntu", "OS name")
	osImageCmd.PersistentFlags().StringVar(&osVersion, "os-version", "2004", "OS version")
	osImageCmd.PersistentFlags().StringVar(&osArch, "os-arch", "amd64", "OS amd64")

	osImageCmd.AddCommand(oracleCmd)
}
