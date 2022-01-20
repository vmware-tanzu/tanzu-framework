// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateAWSCloudFormationTemplateCmd = &cobra.Command{
	Use:   "generate-cloudformation-template",
	Short: "Generate AWS CloudFormation Template",
	Long:  `Generate AWS CloudFormation Template`,
	RunE:  generateCloudFormationTemplate,
}

func init() {
	generateAWSCloudFormationTemplateCmd.Flags().StringVarP(&setAWSPermissionsOps.clusterConfigFile, "file", "f", "", "Optional, configuration file from which to read the aws credentials. Falls back to using the default AWS credentials chain if not provided.")
	awsPermissionsCmd.AddCommand(generateAWSCloudFormationTemplateCmd)
}

func generateCloudFormationTemplate(cmd *cobra.Command, args []string) error {
	forceUpdateTKGCompatibilityImage := false
	tkgctlClient, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	if err != nil {
		return err
	}
	template, err := tkgctlClient.GenerateAWSCloudFormationTemplate(setAWSPermissionsOps.clusterConfigFile)
	if err != nil {
		return err
	}
	fmt.Println(template)
	return nil
}
