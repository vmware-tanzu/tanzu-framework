// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/tkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

const errInvalidPasswordFlags = "exactly one of --password, --password-file, --password-env-var flags should be provided"
const errEmptyValue = "the value for %s flag should not be empty"

var registrySecretAddCmd = &cobra.Command{
	Use:   "add SECRET_NAME --server REGISTRY_SERVER --username USERNAME --password PASSWORD",
	Short: "Creates a v1/Secret resource",
	Long:  "Creates a v1/Secret resource of type kubernetes.io/dockerconfigjson. In case of specifying the --export-to-all-namespaces flag, a SecretExport resource will also get created.",
	Example: `
    # Add a registry secret
    tanzu secret registry add test-secret --server projects-stg.registry.vmware.com --username test-user --password-file test-file

    # Add a registry secret with 'export-to-all-namespaces' flag being set
    tanzu secret registry add test-secret --server projects-stg.registry.vmware.com --username test-user --password test-pass --export-to-all-namespaces`,
	RunE:         registrySecretAdd,
	SilenceUsage: true,
}

func init() {
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.Server, "server", "", "", "Private registry server FQDN")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.Username, "username", "", "", "Username for authenticating to the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.PasswordInput, "password", "", "", "Password for authenticating to the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.PasswordFile, "password-file", "", "", "File containing the password for authenticating to the private registry")
	registrySecretAddCmd.Flags().StringVarP(&registrySecretOp.PasswordEnvVar, "password-env-var", "", "", "Environment variable containing the password for authenticating to the private registry")
	registrySecretAddCmd.Flags().BoolVarP(&registrySecretOp.PasswordStdin, "password-stdin", "", false, "When provided, password for authenticating to the private registry would be taken from the standard input")
	registrySecretAddCmd.Flags().BoolVarP(&registrySecretOp.ExportToAllNamespaces, "export-to-all-namespaces", "", false, "Make the registry secret available across all namespaces (i.e. create SecretExport with toNamespace=*), optional")
	registrySecretAddCmd.Flags().BoolVarP(&registrySecretOp.SkipPrompt, "yes", "y", false, "In case the --export-to-all-namespaces flag was set, export the secret to all namespaces without asking for confirmation, optional")
	registrySecretCmd.AddCommand(registrySecretAddCmd)
	registrySecretAddCmd.Args = cobra.ExactArgs(1)
	registrySecretAddCmd.MarkFlagRequired("server")   //nolint
	registrySecretAddCmd.MarkFlagRequired("username") //nolint
}

func registrySecretAdd(cmd *cobra.Command, args []string) error {
	registrySecretOp.SecretName = args[0]

	if err := fetchAndValidateSecretCredentials(); err != nil {
		return err
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	kc, err := kappclient.NewKappClient(kubeConfig)
	if err != nil {
		return err
	}

	if registrySecretOp.ExportToAllNamespaces {
		// Creating a SecretExport resource X that conflicts with a previously defined SecretExport Y that was exported to all namespaces could result in privilege escalation if the user does not access to other namespaces. This check prevents it by trying to list the namespaces
		if err := checkNamespaceList(kc); err != nil {
			return err
		}
		msg := "Warning: By choosing --export-to-all-namespaces, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope.\n\n"
		err := issueWarningAndPromptUser(msg)
		if err != nil {
			return err
		}
	} else if err := checkSecretExportExists(pkgClient, kc); err != nil {
		return err
	}

	// as the secret might already exist, first check for its existence in order to have an idempotent "add" operation
	notFound, err := checkSecretExists(kc)
	if err != nil {
		return err
	}

	// If the secret doesn't exist, create it
	if notFound {
		if _, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat, fmt.Sprintf("Adding registry secret '%s'...", registrySecretOp.SecretName), true); err != nil {
			return err
		}

		if err := pkgClient.AddRegistrySecret(registrySecretOp); err != nil {
			return err
		}

		log.Infof("\n Added registry secret '%s' into namespace '%s'", registrySecretOp.SecretName, registrySecretOp.Namespace)
		if registrySecretOp.ExportToAllNamespaces {
			log.Infof(" Exported registry secret '%s' to all namespaces", registrySecretOp.SecretName)
		}
		return nil
	}

	// If the secret already exists, just update it
	if err := updateSecret(cmd, pkgClient); err != nil {
		return err
	}
	return nil
}

func updateSecret(cmd *cobra.Command, pkgClient tkgpackageclient.TKGPackageClient) error {
	if registrySecretOp.ExportToAllNamespaces {
		t := true
		registrySecretOp.Export = tkgpackagedatamodel.TypeBoolPtr{ExportToAllNamespaces: &t}
	}

	if _, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat, fmt.Sprintf("Updating registry secret '%s'...", registrySecretOp.SecretName), true); err != nil {
		return err
	}

	if err := pkgClient.UpdateRegistrySecret(registrySecretOp); err != nil {
		return err
	}

	log.Infof("\n Updated registry secret '%s' in namespace '%s'", registrySecretOp.SecretName, registrySecretOp.Namespace)
	if registrySecretOp.ExportToAllNamespaces {
		log.Infof(" Exported registry secret '%s' to all namespaces", registrySecretOp.SecretName)
	}

	return nil
}

func fetchAndValidateSecretCredentials() error {
	if registrySecretOp.Server == "" {
		return fmt.Errorf(errEmptyValue, "--server")
	}

	if registrySecretOp.Username == "" {
		return fmt.Errorf(errEmptyValue, "--username")
	}

	password, err := extractPassword()
	if err != nil {
		return err
	}
	if password == "" {
		return errors.New(errInvalidPasswordFlags)
	}
	registrySecretOp.Password = password

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
		b, err := os.ReadFile(registrySecretOp.PasswordFile)
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

func checkNamespaceList(kc kappclient.Client) error {
	ns := &corev1.NamespaceList{}

	err := kc.GetClient().List(context.Background(), ns)
	if err != nil {
		return err
	}
	return nil
}

func checkSecretExportExists(pkgClient tkgpackageclient.TKGPackageClient, kc kappclient.Client) error {
	export := ""
	secretExport, err := pkgClient.GetSecretExport(registrySecretOp)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// No error means we found a matching SecretExport

	// Creating a SecretExport resource X that conflicts with a previously defined SecretExport Y that was exported to all namespaces could result in privilege escalation if the user does not access to other namespaces. This check prevents it by trying to list the namespaces
	if err := checkNamespaceList(kc); err != nil {
		return err
	}

	if findInList(secretExport.Spec.ToNamespaces, "*") || secretExport.Spec.ToNamespace == "*" {
		export = "all namespaces"
	} else {
		export = "some namespaces"
	}

	// Ask user consent when SecretExport has been created by kubectl and user tries to add secret of the same name as SecretExport without using --export-to-all-namespaces flag
	msg := "Warning: SecretExport with the same name exists already, given secret contents will be available to %s. If you decide not to proceed, you can either delete the SecretExport or specify a different secret name.\n\n"
	msg = fmt.Sprintf(msg, export)
	err = issueWarningAndPromptUser(msg)
	if err != nil {
		return err
	}

	return nil
}

func checkSecretExists(kc kappclient.Client) (bool, error) {
	var notFound bool
	err := kc.GetClient().Get(context.Background(), crtclient.ObjectKey{Name: registrySecretOp.SecretName, Namespace: registrySecretOp.Namespace}, &corev1.Secret{})
	if err != nil {
		notFound = apierrors.IsNotFound(err)
		if !notFound {
			return notFound, err
		}
	}
	return notFound, nil
}

func issueWarningAndPromptUser(msg string) error {
	log.Warning(msg)
	if !registrySecretOp.SkipPrompt {
		if err := cli.AskForConfirmation("Are you sure you want to proceed?"); err != nil {
			return errors.New("creation of the secret got aborted")
		}
	}
	log.Info("\n")
	return nil
}
