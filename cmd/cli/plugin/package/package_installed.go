// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageinstalledOp = tkgpackagedatamodel.NewPackageInstalledOptions()

var packageInstalledCmd = &cobra.Command{
	Use:       "installed",
	ValidArgs: []string{"list", "create", "delete", "update", "get"},
	Short:     "Manage installed packages",
	Args:      cobra.RangeArgs(1, 2),
}

func init() {
	packageInstalledCmd.PersistentFlags().StringVarP(&packageinstalledOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	packageInstalledCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}
