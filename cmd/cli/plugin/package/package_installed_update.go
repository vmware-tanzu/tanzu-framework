// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var packageInstalledUpdateCmd = &cobra.Command{
	Use:   "update INSTALLED_PACKAGE_NAME",
	Short: "Update an installed package",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Update installed package with name 'mypkg' with some version to version '3.0.0-rc.1' in specified namespace 	
    tanzu package installed update mypkg --version 3.0.0-rc.1 --namespace test-ns`,
	RunE:         packageUpdate,
	SilenceUsage: true,
}

func init() {
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.Version, "version", "v", "", "The version which installed package needs to be updated to, optional")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file, optional")
	packageInstalledUpdateCmd.Flags().BoolVarP(&packageInstalledOp.Install, "install", "", false, "Install package if the installed package does not exist, optional")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.PackageName, "package-name", "p", "", "The public name for the package, optional")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.Namespace, "namespace", "n", "default", "The namespace to locate the installed package which needs to be updated")
	packageInstalledUpdateCmd.Flags().BoolVarP(&packageInstalledOp.Wait, "wait", "", true, "Wait for the package reconciliation to complete, optional. To disable wait, specify --wait=false")
	packageInstalledUpdateCmd.Flags().DurationVarP(&packageInstalledOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package reconciliation status, optional")
	packageInstalledUpdateCmd.Flags().DurationVarP(&packageInstalledOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package reconciliation status, optional")
	packageInstalledCmd.AddCommand(packageInstalledUpdateCmd)
}

func packageUpdate(cmd *cobra.Command, args []string) error {
	packageInstalledOp.PkgInstallName = args[0]

	if packageInstalledOp.Version == "" && packageInstalledOp.ValuesFile == "" {
		return errors.New("please provide --version and/or --values-file for updating the installed package")
	}

	if packageInstalledOp.Install {
		if packageInstalledOp.PackageName == "" {
			return errors.New("--package-name is required when --install flag is declared")
		}
		if packageInstalledOp.Version == "" {
			return errors.New("--version is required when --install flag is declared")
		}
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	return pkgClient.UpdatePackageSync(packageInstalledOp, tkgpackagedatamodel.OperationTypeUpdate)
}
