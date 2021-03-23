// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var setAWSPermissionsCmd = &cobra.Command{
	Use:   "set",
	Short: "Configure permissions on AWS",
	Long:  `Configure permissions on AWS`,
	RunE:  setAWSPermissions,
}

type setAWSPermissionsOptions struct {
	clusterConfigFile string
}

var setAWSPermissionsOps setAWSPermissionsOptions

func init() {
	setAWSPermissionsCmd.Flags().StringVarP(&setAWSPermissionsOps.clusterConfigFile, "file", "f", "", "Optional, configuration file from which to read the aws credentials. Falls back to using the default AWS credentials chain if not provided.")
	awsPermissionsCmd.AddCommand(setAWSPermissionsCmd)
}

func setAWSPermissions(cmd *cobra.Command, args []string) error {
	tkgctlClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}
	return tkgctlClient.CreateAWSCloudFormationStack(setAWSPermissionsOps.clusterConfigFile)
}
