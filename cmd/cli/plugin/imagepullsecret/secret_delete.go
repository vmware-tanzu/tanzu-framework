// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
)

var imagePullSecretDeleteCmd = &cobra.Command{
	Use:   "delete SECRET_NAME",
	Short: "Deletes v1/Secret  resource of type kubernetes.io/dockerconfigjson and the associated SecretExport from the cluster",
	Example: `
    # Delete an image pull secret
    tanzu imagepullsecret delete test-secret`,
	PreRunE: secretGenAvailabilityCheck,
	RunE:    imagePullSecretDelete,
}

func init() {
	imagePullSecretDeleteCmd.Flags().StringVarP(&imagePullSecretOp.Namespace, "namespace", "n", "default", "Namespace for the image pull secret, optional")
	imagePullSecretDeleteCmd.Flags().StringVarP(&imagePullSecretOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	imagePullSecretDeleteCmd.Flags().BoolVarP(&imagePullSecretOp.SkipPrompt, "yes", "y", false, "Delete the image pull secret without asking for confirmation, optional")
	imagePullSecretDeleteCmd.Args = cobra.ExactArgs(1)
}

func imagePullSecretDelete(cmd *cobra.Command, args []string) error {
	imagePullSecretOp.SecretName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(imagePullSecretOp.KubeConfig)
	if err != nil {
		return err
	}

	if !imagePullSecretOp.SkipPrompt {
		if err := cli.AskForConfirmation(fmt.Sprintf("Deleting image pull secret '%s' from namespace '%s'. Are you sure?",
			imagePullSecretOp.SecretName, imagePullSecretOp.Namespace)); err != nil {
			return err
		}
	}

	if _, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Deleting image pull secret '%s'...", imagePullSecretOp.SecretName), true); err != nil {
		return err
	}

	found, err := pkgClient.DeleteImagePullSecret(imagePullSecretOp)
	if !found {
		log.Warningf("\n image pull secret '%s' does not exist in namespace '%s'", imagePullSecretOp.SecretName, imagePullSecretOp.Namespace)
		return nil
	}
	if err != nil {
		return err
	}

	log.Infof("\n Deleted image pull secret '%s' from namespace '%s'", imagePullSecretOp.SecretName, imagePullSecretOp.Namespace)
	return nil
}
