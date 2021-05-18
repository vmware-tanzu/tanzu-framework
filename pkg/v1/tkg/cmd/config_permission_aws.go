// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type configPermissionsAWSOptions struct {
	clusterConfigFile string
}

var cpa = &configPermissionsAWSOptions{}

var configPermissionsAWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "Configure permissions on AWS",
	Long: LongDesc(`Configure permissions on AWS. This is done via creating a new AWS CloudFormation stack with the correct IAM resources.
	This command assumes that the following configuration values are set in the tkg configuration file or in the AWS default credential provider chain https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html.
	AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY.`),
	Example: Examples(`
	# Configure permissions on AWS
	tkg config permissions aws`),
	Run: func(cmd *cobra.Command, args []string) {
		err := runConfigPermissionsAWS()
		verifyCommandError(err)
	},
}

func init() {
	configPermissionsCmd.Flags().StringVarP(&cpa.clusterConfigFile, "file", "", "", "The cluster configuration file (default \"$HOME/.tkg/cluster-config.yaml\")")

	configPermissionsCmd.AddCommand(configPermissionsAWSCmd)
}

func runConfigPermissionsAWS() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}
	return tkgClient.CreateAWSCloudFormationStack(cpa.clusterConfigFile)
}
