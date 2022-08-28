// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var packageAvailableOp = tkgpackagedatamodel.NewPackageAvailableOptions()

var packageAvailableCmd = &cobra.Command{
	Use:               "available",
	ValidArgs:         []string{"list", "get"},
	Short:             "Manage available packages",
	Args:              cobra.RangeArgs(1, 2),
	PersistentPreRunE: packagingAvailabilityCheck,
}

func init() {
	packageAvailableCmd.PersistentFlags().StringVarP(&packageAvailableOp.Namespace, "namespace", "n", "default", "Namespace of packages, optional")
	packageAvailableCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table), optional")
}

func packagingAvailabilityCheck(_ *cobra.Command, _ []string) error {
	found, err := isPackagingAPIAvailable(kubeConfig)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to check for the availability of '%s' API", tkgpackagedatamodel.PackagingAPIName))
	}
	if !found {
		return fmt.Errorf(tkgpackagedatamodel.PackagingAPINotAvailable, tkgpackagedatamodel.PackagingAPIName, tkgpackagedatamodel.PackagingAPIVersion)
	}

	return nil
}
