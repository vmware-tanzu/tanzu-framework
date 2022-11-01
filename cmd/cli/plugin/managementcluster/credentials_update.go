// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type updateCredentialsOptions struct {
	vSphereUser         string
	vSpherePassword     string
	azureTenantID       string
	azureSubscriptionID string
	azureClientID       string
	azureClientSecret   string
	isCascading         bool
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
	SilenceUsage: true,
}

func init() {
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSphereUser, "vsphere-user", "", "", "Username for vSphere provider")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSpherePassword, "vsphere-password", "", "", "Password for vSphere provider")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.azureTenantID, "azure-tenant-id", "", "", "ID for Azure Active Directory in which the app for Tanzu Kubernetes Grid is created")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.azureSubscriptionID, "azure-subscription-id", "", "", "GUID that uniquely identifies the subscription to use Azure services")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.azureClientID, "azure-client-id", "", "", "Client ID of the app for Tanzu Kubernetes Grid that you registered with Azure")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.azureClientSecret, "azure-client-secret", "", "", "Client Password of the app for Tanzu Kubernetes Grid that you registered with Azure")
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

	forceUpdateTKGCompatibilityImage := false
	tkgctlClient, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	if err != nil {
		return err
	}

	provider := ""

	if updateCredentialsOpts.vSphereUser != "" {
		provider = "vsphere"
	} else if updateCredentialsOpts.azureClientID != "" {
		provider = "azure"
	} else {
		err := component.Prompt(
			&component.PromptConfig{
				Message: "Specify provider \"vsphere\" or \"azure\"",
			},
			&provider,
			promptOpts...,
		)
		if err != nil {
			return err
		}
	}

	if provider == "vsphere" {
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
	} else if provider == "azure" {
		if updateCredentialsOpts.azureClientID == "" {
			err = component.Prompt(
				&component.PromptConfig{
					Message: "Enter azure client id",
				},
				&updateCredentialsOpts.azureClientID,
				promptOpts...,
			)
			if err != nil {
				return err
			}
		}

		if updateCredentialsOpts.azureClientSecret == "" {
			err = component.Prompt(
				&component.PromptConfig{
					Message:   "Enter azure client secret",
					Sensitive: true,
				},
				&updateCredentialsOpts.azureClientSecret,
				promptOpts...,
			)
			if err != nil {
				return err
			}
		}

		if updateCredentialsOpts.azureTenantID == "" {
			err = component.Prompt(
				&component.PromptConfig{
					Message: "Enter azure tenant id",
				},
				&updateCredentialsOpts.azureTenantID,
				promptOpts...,
			)
			if err != nil {
				return err
			}
		}

		if updateCredentialsOpts.azureSubscriptionID == "" {
			err = component.Prompt(
				&component.PromptConfig{
					Message:   "Enter azure subscription id",
				},
				&updateCredentialsOpts.azureSubscriptionID,
				promptOpts...,
			)
			if err != nil {
				return err
			}
		}
	}

	options := tkgctl.UpdateCredentialsRegionOptions{
		ClusterName:         clusterName,
		VSphereUsername:     updateCredentialsOpts.vSphereUser,
		VSpherePassword:     updateCredentialsOpts.vSpherePassword,
		AzureTenantID:       updateCredentialsOpts.azureTenantID,
		AzureSubscriptionID: updateCredentialsOpts.azureSubscriptionID,
		AzureClientID:       updateCredentialsOpts.azureClientID,
		AzureClientSecret:   updateCredentialsOpts.azureClientSecret,
		IsCascading:         updateCredentialsOpts.isCascading,
	}

	return tkgctlClient.UpdateCredentialsRegion(options)
}
