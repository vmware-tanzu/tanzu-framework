// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type updateCredentialsOptions struct {
	vSphereUser     string
	vSpherePassword string
	isCascading     bool
}

var updateCredentialsOpts = updateCredentialsOptions{}

var credentialsUpdateCmd = &cobra.Command{
	Use:   "update CLUSTER_NAME",
	Short: "Update credentials for management cluster",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var clusterName string
		if len(args) == 0 {
			clusterName = ""
		} else {
			clusterName = args[0]
		}

		return updateClusterCredentials(clusterName)
	},
}

func init() {
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSphereUser, "vsphere-user", "", "", "Username for vSphere provider")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSpherePassword, "vsphere-password", "", "", "Password for vSphere provider")
	credentialsUpdateCmd.Flags().BoolVarP(&updateCredentialsOpts.isCascading, "cascading", "", false, "Update credentials for all workload clusters under the management cluster")

	credentialsCmd.AddCommand(credentialsUpdateCmd)
}

func updateClusterCredentials(clusterName string) error {
	if clusterName == "" {
		server, err := config.GetCurrentServer()
		if err != nil {
			return err
		}

		clusterName = server.Name
	}

	var promptOpts []component.PromptOpt

	tkgctlClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	if updateCredentialsOpts.vSphereUser == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Enter vSphere username",
			},
			&updateCredentialsOpts.vSphereUser,
			promptOpts...,
		)
		if err != nil {
			return err
		}
	}

	if updateCredentialsOpts.vSpherePassword == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message:   "Enter vSphere password",
				Sensitive: true,
			},
			&updateCredentialsOpts.vSpherePassword,
			promptOpts...,
		)
		if err != nil {
			return err
		}
	}

	options := tkgctl.UpdateCredentialsRegionOptions{
		ClusterName:     clusterName,
		VSphereUsername: updateCredentialsOpts.vSphereUser,
		VSpherePassword: updateCredentialsOpts.vSpherePassword,
		IsCascading:     updateCredentialsOpts.isCascading,
	}

	return tkgctlClient.UpdateCredentialsRegion(options)
}
