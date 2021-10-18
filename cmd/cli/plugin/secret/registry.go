// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

var registrySecretOp = tkgpackagedatamodel.NewRegistrySecretOptions()

var registrySecretCmd = &cobra.Command{
	Use:               "registry",
	Short:             "Registry secret operations",
	ValidArgs:         []string{"add", "list", "delete", "update"},
	Args:              cobra.RangeArgs(1, 3),
	Long:              `Add, list, delete or update a registry secret. Registry secrets enable the package and package repository consumers to authenticate to and pull images from private registries.`,
	PersistentPreRunE: secretGenAvailabilityCheck,
}

func init() {
	registrySecretCmd.PersistentFlags().StringVarP(&registrySecretOp.Namespace, "namespace", "n", "default", "Target namespace for the registry secret, optional")
	registrySecretCmd.PersistentFlags().StringVarP(&registrySecretOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
}

func secretGenAvailabilityCheck(_ *cobra.Command, _ []string) error {
	const secretGenGVR = "secretgen.carvel.dev/v1alpha1"
	found, err := isSecretGenAPIAvailable()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to check for the availability of '%s' API", secretGenGVR))
	}
	if !found {
		return errors.New(fmt.Sprintf("secret registry operations can not be used as '%s' API is not available in the cluster", secretGenGVR))
	}
	return nil
}
