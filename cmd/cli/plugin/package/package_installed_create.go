// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageInstalledCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Install a package",
	Args:  cobra.ExactArgs(1),
	RunE:  packageInstall,
}

func init() {
	packageInstalledCreateCmd.Flags().StringVarP(&packageInstalledOp.PackageName, "package-name", "p", "", "Name of the package to be installed")
	packageInstalledCreateCmd.Flags().StringVarP(&packageInstalledOp.Version, "version", "v", "", "Version of the package to be installed")
	packageInstalledCreateCmd.Flags().BoolVarP(&packageInstalledOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	packageInstalledCreateCmd.Flags().StringVarP(&packageInstalledOp.Namespace, "namespace", "n", "default", "Target namespace to install the package, optional")
	packageInstalledCreateCmd.Flags().StringVarP(&packageInstalledOp.ServiceAccountName, "service-account-name", "", "", "Name of an existing service account used to install underlying package contents, optional")
	packageInstalledCreateCmd.Flags().StringVarP(&packageInstalledOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file, optional")
	packageInstalledCreateCmd.Flags().BoolVarP(&packageInstalledOp.Wait, "wait", "", true, "Wait for the package reconciliation to complete, optional")
	packageInstalledCreateCmd.Flags().DurationVarP(&packageInstalledOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package reconciliation status, optional")
	packageInstalledCreateCmd.Flags().DurationVarP(&packageInstalledOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package reconciliation status, optional")
	packageInstalledCreateCmd.MarkFlagRequired("package-name") //nolint
	packageInstalledCreateCmd.MarkFlagRequired("version")      //nolint
	packageInstalledCmd.AddCommand(packageInstalledCreateCmd)
}
