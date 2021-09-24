// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
)

var imagePullSecretUpdateCmd = &cobra.Command{
	Use:   "update SECRET_NAME --username USERNAME --password PASSWORD",
	Short: "Updates the v1/Secret resource of type kubernetes.io/dockerconfigjson. In case of specifying the --export-to-all-namespaces flag, the SecretExport resource will also get updated. Otherwise, there will be no changes in the SecretExport resource",
	Example: `
    # Update an image pull secret. There will be no changes in the associated SecretExport resource
    tanzu imagepullsecret update test-secret --username test-user --password-file test-file

    # Update an image pull secret with 'export-to-all-namespaces' flag being set
    tanzu imagepullsecret update test-secret --username test-user --password test-pass --export-to-all-namespaces

    # Update an image pull secret with 'export-to-all-namespaces' flag being clear. In this case, the associated SecretExport resource will get deleted
    tanzu imagepullsecret update test-secret --username test-user --password test-pass --export-to-all-namespaces=false`,
	PreRunE: secretGenAvailabilityCheck,
	RunE:    imagePullSecretUpdate,
}

func init() {
	imagePullSecretUpdateCmd.Flags().StringVarP(&imagePullSecretOp.Username, "username", "", "", "Username for authenticating to the private registry")
	imagePullSecretUpdateCmd.Flags().StringVarP(&imagePullSecretOp.PasswordInput, "password", "", "", "Password for authenticating to the private registry")
	imagePullSecretUpdateCmd.Flags().StringVarP(&imagePullSecretOp.PasswordFile, "password-file", "", "", "File containing the password for authenticating to the private registry")
	imagePullSecretUpdateCmd.Flags().StringVarP(&imagePullSecretOp.PasswordEnvVar, "password-env-var", "", "", "Environment variable containing the password for authenticating to the private registry")
	imagePullSecretUpdateCmd.Flags().StringVarP(&imagePullSecretOp.Namespace, "namespace", "n", "default", "Target namespace for the image pull secret, optional")
	imagePullSecretUpdateCmd.Flags().StringVarP(&imagePullSecretOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	imagePullSecretUpdateCmd.Flags().BoolVarP(&imagePullSecretOp.PasswordStdin, "password-stdin", "", false, "When provided, password for authenticating to the private registry would be taken from the standard input")
	imagePullSecretUpdateCmd.Flags().VarPF(&imagePullSecretOp.Export, "export-to-all-namespaces", "", "If set to true, the secret gets available across all namespaces. If set to false, the secret will get unexported from ALL namespaces in which it was previously exported to. In case of not specifying this flag, no changes will be made in the existing SecretExport resource. optional").NoOptDefVal = "true"
	imagePullSecretUpdateCmd.Args = cobra.ExactArgs(1)
}

func imagePullSecretUpdate(cmd *cobra.Command, args []string) error {
	imagePullSecretOp.SecretName = args[0]

	password, err := extractPassword()
	if err != nil {
		return err
	}
	imagePullSecretOp.Password = password

	if imagePullSecretOp.Username == "" && imagePullSecretOp.Password == "" && imagePullSecretOp.Export.ExportToAllNamespaces == nil {
		return errors.New("no changes made in the image pull secret, as neither of username, password or export-to-all-namespaces flag options were provided")
	}

	cmd.SilenceUsage = true

	if imagePullSecretOp.Export.ExportToAllNamespaces != nil {
		if *imagePullSecretOp.Export.ExportToAllNamespaces {
			log.Warning("Warning: By specifying --export-to-all-namespaces as true, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope.\n")
		} else {
			log.Warning("Warning: By specifying --export-to-all-namespaces as false, the secret contents will get unexported from ALL namespaces in which it was previously available to.\n")
		}
		if err := cli.AskForConfirmation("Are you sure you want to proceed?"); err != nil {
			return errors.New("update of the secret got aborted")
		}
		log.Info("\n")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(imagePullSecretOp.KubeConfig)
	if err != nil {
		return err
	}

	if _, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Updating image pull secret '%s'...", imagePullSecretOp.SecretName), true); err != nil {
		return err
	}

	if err := pkgClient.UpdateImagePullSecret(imagePullSecretOp); err != nil {
		return err
	}

	log.Infof("\n Updated image pull secret '%s' in namespace '%s'", imagePullSecretOp.SecretName, imagePullSecretOp.Namespace)
	if imagePullSecretOp.Export.ExportToAllNamespaces != nil {
		if *imagePullSecretOp.Export.ExportToAllNamespaces {
			log.Infof(" Exported image pull secret '%s' to all namespaces", imagePullSecretOp.SecretName)
		} else {
			log.Infof(" Unexported image pull secret '%s' from all namespaces", imagePullSecretOp.SecretName)
		}
	}
	return nil
}
