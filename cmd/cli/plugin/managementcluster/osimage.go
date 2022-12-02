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
	// the output tkr path
	outputDirectory string
)

func init() {
	osImageCmd.PersistentFlags().StringVar(&tkrRegistryPath, "tkr-path", "", "The input tkr package image path")
	osImageCmd.PersistentFlags().StringVarP(&outputDirectory, "output-directory", "d", "", "The output TKR path")
	osImageCmd.AddCommand(oracleCmd)
}
