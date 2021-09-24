// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var imagePullSecretListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all v1/Secret of type kubernetes.io/dockerconfigjson and checks for the associated SecretExport by the same name",
	Args:  cobra.NoArgs,
	Example: `
    # List all image pull secrets
    tanzu imagepullsecret list`,
	PreRunE: secretGenAvailabilityCheck,
	RunE:    imagePullSecretList,
}

func init() {
	imagePullSecretListCmd.Flags().BoolVarP(&imagePullSecretOp.AllNamespaces, "all-namespaces", "A", false, "If present, list image pull secrets across all namespaces, optional")
	imagePullSecretListCmd.Flags().StringVarP(&imagePullSecretOp.Namespace, "namespace", "n", "default", "Namespace for the image pull secret, optional")
	imagePullSecretListCmd.Flags().StringVarP(&imagePullSecretOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	imagePullSecretListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table), optional")
}

func imagePullSecretList(cmd *cobra.Command, args []string) error {
	return nil
}
