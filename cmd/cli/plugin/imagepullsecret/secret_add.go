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

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

var imagePullSecretOp = tkgpackagedatamodel.NewImagePullSecretOptions()

var imagePullSecretAddCmd = &cobra.Command{
	Use:   "add SECRET_NAME --registry REGISTRY_URL --username USERNAME --password PASSWORD",
	Short: "Creates a v1/Secret resource of type kubernetes.io/dockerconfigjson. In case of specifying the --export-to-all-namespaces flag, a SecretExport resource will also get created",
	Example: `
    # Add an image pull secret
    tanzu imagepullsecret add test-secret --registry projects-stg.registry.vmware.com --username test-user --password-file test-file,

	# Add an image pull secret with 'export-to-all-namespaces' flag being set
	tanzu imagepullsecret add test-secret --registry projects-stg.registry.vmware.com --username test-user --password test-pass --export-to-all-namespaces`,
	PreRunE: secretGenAvailabilityCheck,
	RunE:    imagePullSecretAdd,
}

func init() {
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.Registry, "registry", "", "", "URL of the private registry")
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.Username, "username", "", "", "Username for authenticating to the private registry")
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.PasswordInput, "password", "", "", "Password for authenticating to the private registry")
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.PasswordFile, "password-file", "", "", "File containing the password for authenticating to the private registry")
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.PasswordEnvVar, "password-env-var", "", "", "Environment variable containing the password for authenticating to the private registry")
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.Namespace, "namespace", "n", "default", "Target namespace to add the image pull secret, optional")
	imagePullSecretAddCmd.Flags().StringVarP(&imagePullSecretOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	imagePullSecretAddCmd.Flags().BoolVarP(&imagePullSecretOp.PasswordStdin, "password-stdin", "", false, "When provided, password for authenticating to the private registry would be taken from the standard input, optional")
	imagePullSecretAddCmd.Flags().BoolVarP(&imagePullSecretOp.ExportToAllNamespaces, "export-to-all-namespaces", "", false, "Make the image pull secret available across all namespaces (i.e. create SecretExport with toNamespace=*), optional")
	imagePullSecretAddCmd.Args = cobra.ExactArgs(1)
	imagePullSecretAddCmd.MarkFlagRequired("registry") //nolint
	imagePullSecretAddCmd.MarkFlagRequired("username") //nolint
}

func secretGenAvailabilityCheck(cmd *cobra.Command, _ []string) error {
	const secretGenGVR = "secretgen.carvel.dev/v1alpha1"
	found, err := isSecretGenAPIAvailable()
	if err != nil {
		cmd.SilenceUsage = true
		return errors.Wrap(err, fmt.Sprintf("failed to check for the availability of '%s' API", secretGenGVR))
	}
	if !found {
		cmd.SilenceUsage = true
		return errors.New(fmt.Sprintf("imagepullsecret plugin can not be used as '%s' API is not available in the cluster", secretGenGVR))
	}
	return nil
}

func imagePullSecretAdd(cmd *cobra.Command, args []string) error {
	imagePullSecretOp.SecretName = args[0]

	password, err := extractPassword()
	if err != nil {
		return err
	}
	imagePullSecretOp.Password = password

	if imagePullSecretOp.ExportToAllNamespaces {
		log.Warning("Warning: By choosing --export-to-all-namespaces, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials are read only and are safe to share.\n\n")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(imagePullSecretOp.KubeConfig)
	if err != nil {
		return err
	}

	if _, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Adding image pull secret '%s'...", imagePullSecretOp.SecretName), true); err != nil {
		return err
	}

	if err := pkgClient.AddImagePullSecret(imagePullSecretOp); err != nil {
		return err
	}

	log.Infof("\n Added image pull secret '%s' into namespace '%s'", imagePullSecretOp.SecretName, imagePullSecretOp.Namespace)
	return nil
}

// extractPassword extracts the password from one of the corresponding flags
func extractPassword() (string, error) {
	var (
		isPasswordSet bool
		password      string
	)

	errInvalidPasswordFlags := "exactly one of --password, --password-file, --password-env-var flags should be provided"

	if imagePullSecretOp.PasswordInput != "" {
		password = imagePullSecretOp.PasswordInput
		isPasswordSet = true
	}
	if imagePullSecretOp.PasswordStdin {
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
	if imagePullSecretOp.PasswordFile != "" {
		if isPasswordSet {
			return "", errors.New(errInvalidPasswordFlags)
		}
		isPasswordSet = true
		b, err := ioutil.ReadFile(imagePullSecretOp.PasswordFile)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to read from the password file '%s'", imagePullSecretOp.PasswordFile))
		}
		password = string(b)
	}
	if imagePullSecretOp.PasswordEnvVar != "" {
		if isPasswordSet {
			return "", errors.New(errInvalidPasswordFlags)
		}
		password = os.Getenv(imagePullSecretOp.PasswordEnvVar)
	}

	if password == "" {
		return "", errors.New(errInvalidPasswordFlags)
	}

	return password, nil
}
