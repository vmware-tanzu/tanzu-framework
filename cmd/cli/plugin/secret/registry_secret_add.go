// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
)

const errInvalidPasswordFlags = "exactly one of --password, --password-file, --password-env-var flags should be provided"

var registrySecretAddCmd = &cobra.Command{
	Use:   "add SECRET_NAME --server-url REGISTRY_URL --username USERNAME --password PASSWORD",
	Short: "Creates a v1/Secret resource of type kubernetes.io/dockerconfigjson. In case of specifying the --export-to-all-namespaces flag, a SecretExport resource will also get created",
	Example: `
    # Add a registry secret
    tanzu secret registry add test-secret --server-url projects-stg.registry.vmware.com --username test-user --password-file test-file

    # Add a registry secret with 'export-to-all-namespaces' flag being set
    tanzu secret registry add test-secret --server-url projects-stg.registry.vmware.com --username test-user --password test-pass --export-to-all-namespaces`,
	RunE:         registrySecretAdd,
	SilenceUsage: true,
}

func init() {
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.ServerURL, "server-url", "", "", "URL of the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.Username, "username", "", "", "Username for authenticating to the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.PasswordInput, "password", "", "", "Password for authenticating to the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.PasswordFile, "password-file", "", "", "File containing the password for authenticating to the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.PasswordEnvVar, "password-env-var", "", "", "Environment variable containing the password for authenticating to the private registry")
	registrySecretAddCmd.Flags().BoolVarP(&registrySecretOp.PasswordStdin, "password-stdin", "", false, "When provided, password for authenticating to the private registry would be taken from the standard input")
	registrySecretAddCmd.Flags().BoolVarP(&registrySecretOp.ExportToAllNamespaces, "export-to-all-namespaces", "", false, "Make the registry secret available across all namespaces (i.e. create SecretExport with toNamespace=*), optional")
	registrySecretCmd.AddCommand(registrySecretAddCmd)
	registrySecretAddCmd.Args = cobra.ExactArgs(1)
	registrySecretAddCmd.MarkFlagRequired("server-url") //nolint
	registrySecretAddCmd.MarkFlagRequired("username")   //nolint
}

func registrySecretAdd(cmd *cobra.Command, args []string) error {
	registrySecretOp.SecretName = args[0]

	password, err := extractPassword()
	if err != nil {
		return err
	}
	if password == "" {
		return errors.New(errInvalidPasswordFlags)
	}
	registrySecretOp.Password = password

	if registrySecretOp.ExportToAllNamespaces {
		log.Warning("Warning: By choosing --export-to-all-namespaces, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope.\n\n")
		if err := cli.AskForConfirmation("Are you sure you want to proceed?"); err != nil {
			return errors.New("creation of the secret got aborted")
		}
		log.Info("\n")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(registrySecretOp.KubeConfig)
	if err != nil {
		return err
	}

	if _, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Adding registry secret '%s'...", registrySecretOp.SecretName), true); err != nil {
		return err
	}

	if err := pkgClient.AddRegistrySecret(registrySecretOp); err != nil {
		return err
	}

	log.Infof("\n Added registry secret '%s' into namespace '%s'", registrySecretOp.SecretName, registrySecretOp.Namespace)
	return nil
}

// extractPassword extracts the password from one of the corresponding flags
func extractPassword() (string, error) {
	var (
		isPasswordSet bool
		password      string
	)

	if registrySecretOp.PasswordInput != "" {
		password = registrySecretOp.PasswordInput
		isPasswordSet = true
	}
	if registrySecretOp.PasswordStdin {
		if isPasswordSet {
			return "", errors.New(errInvalidPasswordFlags)
		}
		isPasswordSet = true
		log.Info("Password:")
		b, err := term.ReadPassword(0)
		if err != nil {
			return "", errors.Wrap(err, "failed to read the password from standard input")
		}
		password = string(b)
	}
	if registrySecretOp.PasswordFile != "" {
		if isPasswordSet {
			return "", errors.New(errInvalidPasswordFlags)
		}
		isPasswordSet = true
		b, err := ioutil.ReadFile(registrySecretOp.PasswordFile)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to read from the password file '%s'", registrySecretOp.PasswordFile))
		}
		password = string(b)
	}
	if registrySecretOp.PasswordEnvVar != "" {
		if isPasswordSet {
			return "", errors.New(errInvalidPasswordFlags)
		}
		password = os.Getenv(registrySecretOp.PasswordEnvVar)
	}

	return password, nil
}
