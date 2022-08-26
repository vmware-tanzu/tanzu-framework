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

var registrySecretDeleteCmd = &cobra.Command{
	Use:   "delete SECRET_NAME",
	Short: "Deletes a v1/Secret resource",
	Long:  "Deletes a v1/Secret resource of type kubernetes.io/dockerconfigjson and the associated SecretExport from the cluster",
	Example: `
    # Delete a registry secret
    tanzu registry secret delete test-secret`,
	RunE:         registrySecretDelete,
	SilenceUsage: true,
}

func init() {
	registrySecretDeleteCmd.Flags().BoolVarP(&registrySecretOp.SkipPrompt, "yes", "y", false, "Delete the registry secret without asking for confirmation, optional")
	registrySecretCmd.AddCommand(registrySecretDeleteCmd)
	registrySecretDeleteCmd.Args = cobra.ExactArgs(1)
}

func registrySecretDelete(cmd *cobra.Command, args []string) error {
	registrySecretOp.SecretName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	if !registrySecretOp.SkipPrompt {
		if err := cli.AskForConfirmation(fmt.Sprintf("Deleting registry secret '%s' from namespace '%s'. Are you sure?",
			registrySecretOp.SecretName, registrySecretOp.Namespace)); err != nil {
			return errors.New("deletion of the secret got aborted")
		}
	}

	if _, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Deleting registry secret '%s'...", registrySecretOp.SecretName), true); err != nil {
		return err
	}

	found, err := pkgClient.DeleteRegistrySecret(registrySecretOp)
	if !found {
		log.Warningf("\n registry secret '%s' does not exist in namespace '%s'", registrySecretOp.SecretName, registrySecretOp.Namespace)
		return nil
	}
	if err != nil {
		return err
	}

	log.Infof("\n Deleted registry secret '%s' from namespace '%s'", registrySecretOp.SecretName, registrySecretOp.Namespace)
	return nil
}
