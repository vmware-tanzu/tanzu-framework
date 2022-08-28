// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var packageInstalledOp = tkgpackagedatamodel.NewPackageOptions()

var packageInstalledCmd = &cobra.Command{
	Use:               "installed",
	ValidArgs:         []string{"list", "create", "delete", "update", "get"},
	Short:             "Manage installed packages",
	Args:              cobra.RangeArgs(1, 2),
	PersistentPreRunE: packagingAvailabilityCheck,
}
