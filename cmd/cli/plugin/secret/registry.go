// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
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
}

func secretGenAvailabilityCheck(_ *cobra.Command, _ []string) error {
	found, err := isSecretGenAPIAvailable(kubeConfig)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to check for the availability of '%s' API", tkgpackagedatamodel.SecretGenAPIName))
	}
	if !found {
		return fmt.Errorf(tkgpackagedatamodel.SecretGenAPINotAvailable, tkgpackagedatamodel.SecretGenAPIName, tkgpackagedatamodel.SecretGenAPIVersion)
	}

	return nil
}
