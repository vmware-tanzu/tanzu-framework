// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type updateCredentialsOptions struct {
	vSphereUser       string
	vSpherePassword   string
	azureTenantID     string
	azureClientID     string
	azureClientSecret string
	isCascading       bool
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

const (
	vsphereProvider = "vsphere"
	azureProvider   = "azure"
)

func init() {
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSphereUser, "vsphere-user", "", "", "Username for vSphere provider")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSpherePassword, "vsphere-password", "", "", "Password for vSphere provider")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.azureTenantID, "azure-tenant-id", "", "", "ID for Azure Active Directory in which the app for Tanzu Kubernetes Grid is created")
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
		provider = vsphereProvider
	} else if updateCredentialsOpts.azureClientID != "" {
		provider = azureProvider
	} else {
		err := component.Prompt(
			&component.PromptConfig{
				Message: fmt.Sprintf("Specify provider %q or %q", vsphereProvider, azureProvider),
				Default: "vsphere",
			},
			&provider,
			promptOpts...,
		)
		if err != nil {
			return err
		}
	}

	if provider == vsphereProvider {
		vsphereVariables := [2]*string{&updateCredentialsOpts.vSphereUser, &updateCredentialsOpts.vSpherePassword}
		vsphereMessages := [2]string{"Enter vSphere username", "Enter vSphere password"}
		vsphereSensitive := [2]bool{false, true}
		for i := 0; i < 2; i++ {
			if *vsphereVariables[i] == "" {
				err = component.Prompt(
					&component.PromptConfig{
						Message:   vsphereMessages[i],
						Sensitive: vsphereSensitive[i],
					},
					vsphereVariables[i],
					promptOpts...,
				)
				if err != nil {
					return err
				}
			}
		}
	} else if provider == azureProvider {
		azureVariables := [3]*string{&updateCredentialsOpts.azureTenantID, &updateCredentialsOpts.azureClientID, &updateCredentialsOpts.azureClientSecret}
		azureMessages := [3]string{"Enter azure tenant id", "Enter azure client id", "Enter azure client secret"}
		azureSensitive := [3]bool{false, false, true}
		for i := 0; i < 3; i++ {
			if *azureVariables[i] == "" {
				err = component.Prompt(
					&component.PromptConfig{
						Message:   azureMessages[i],
						Sensitive: azureSensitive[i],
					},
					azureVariables[i],
					promptOpts...,
				)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return errors.New("please specify supported provider name: vsphere or azure")
	}

	options := tkgctl.UpdateCredentialsRegionOptions{
		ClusterName:       clusterName,
		VSphereUsername:   updateCredentialsOpts.vSphereUser,
		VSpherePassword:   updateCredentialsOpts.vSpherePassword,
		AzureTenantID:     updateCredentialsOpts.azureTenantID,
		AzureClientID:     updateCredentialsOpts.azureClientID,
		AzureClientSecret: updateCredentialsOpts.azureClientSecret,
		IsCascading:       updateCredentialsOpts.isCascading,
	}

	return tkgctlClient.UpdateCredentialsRegion(options)
}
