// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
)

var registrySecretUpdateCmd = &cobra.Command{
	Use:   "update SECRET_NAME --username USERNAME --password PASSWORD",
	Short: "Updates the v1/Secret resource",
	Long:  "Updates the v1/Secret resource of type kubernetes.io/dockerconfigjson. In case of specifying the --export-to-all-namespaces flag, the SecretExport resource will also get updated. Otherwise, there will be no changes in the SecretExport resource.",
	Example: `
    # Update a registry secret. There will be no changes in the associated SecretExport resource
    tanzu registry secret update test-secret --username test-user --password-file test-file

    # Update a registry secret with 'export-to-all-namespaces' flag being set
    tanzu registry secret update test-secret --username test-user --password test-pass --export-to-all-namespaces

    # Update a registry secret with 'export-to-all-namespaces' flag being clear. In this case, the associated SecretExport resource will get deleted
    tanzu registry secret update test-secret --username test-user --password test-pass --export-to-all-namespaces=false`,
	RunE:         registrySecretUpdate,
	SilenceUsage: true,
}

func init() {
	registrySecretUpdateCmd.Flags().StringVarP(&registrySecretOp.Username, "username", "", "", "Username for authenticating to the private registry")
	registrySecretUpdateCmd.Flags().StringVarP(&registrySecretOp.PasswordInput, "password", "", "", "Password for authenticating to the private registry")
	registrySecretUpdateCmd.Flags().StringVarP(&registrySecretOp.PasswordFile, "password-file", "", "", "File containing the password for authenticating to the private registry")
	registrySecretUpdateCmd.Flags().StringVarP(&registrySecretOp.PasswordEnvVar, "password-env-var", "", "", "Environment variable containing the password for authenticating to the private registry")
	registrySecretUpdateCmd.Flags().BoolVarP(&registrySecretOp.PasswordStdin, "password-stdin", "", false, "When provided, password for authenticating to the private registry would be taken from the standard input")
	registrySecretUpdateCmd.Flags().BoolVarP(&registrySecretOp.SkipPrompt, "yes", "y", false, "In case the --export-to-all-namespaces flag was provided, export/un-export of the secret will be performed without asking for confirmation, optional")
	registrySecretUpdateCmd.Flags().VarPF(&registrySecretOp.Export, "export-to-all-namespaces", "", "If set to true, the secret gets available across all namespaces. If set to false, the secret will get unexported from ALL namespaces in which it was previously exported to. In case of not specifying this flag, no changes will be made in the existing SecretExport resource. optional").NoOptDefVal = "true"
	registrySecretCmd.AddCommand(registrySecretUpdateCmd)
	registrySecretUpdateCmd.Args = cobra.ExactArgs(1)
}

func registrySecretUpdate(cmd *cobra.Command, args []string) error {
	registrySecretOp.SecretName = args[0]

	password, err := extractPassword()
	if err != nil {
		return err
	}
	registrySecretOp.Password = password

	if registrySecretOp.Username == "" && registrySecretOp.Password == "" && registrySecretOp.Export.ExportToAllNamespaces == nil {
		return errors.New("no changes made in the registry secret, as neither of username, password or export-to-all-namespaces flag options were provided")
	}

	if registrySecretOp.Export.ExportToAllNamespaces != nil {
		if *registrySecretOp.Export.ExportToAllNamespaces {
			log.Warning("Warning: By specifying --export-to-all-namespaces as true, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope.\n")
		} else {
			log.Warning("Warning: By specifying --export-to-all-namespaces as false, the secret contents will get unexported from ALL namespaces in which it was previously available to.\n")
		}
		if !registrySecretOp.SkipPrompt {
			if err := cli.AskForConfirmation("Are you sure you want to proceed?"); err != nil {
				return errors.New("update of the secret got aborted")
			}
		}
		log.Info("\n")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	if _, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Updating registry secret '%s'...", registrySecretOp.SecretName), true); err != nil {
		return err
	}

	if err := pkgClient.UpdateRegistrySecret(registrySecretOp); err != nil {
		return err
	}

	log.Infof("\n Updated registry secret '%s' in namespace '%s'", registrySecretOp.SecretName, registrySecretOp.Namespace)
	if registrySecretOp.Export.ExportToAllNamespaces != nil {
		if *registrySecretOp.Export.ExportToAllNamespaces {
			log.Infof(" Exported registry secret '%s' to all namespaces", registrySecretOp.SecretName)
		} else {
			log.Infof(" Unexported registry secret '%s' from all namespaces", registrySecretOp.SecretName)
		}
	}
	return nil
}
