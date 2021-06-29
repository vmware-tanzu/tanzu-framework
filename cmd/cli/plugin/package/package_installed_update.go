// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageInstalledUpdateOp = tkgpackagedatamodel.NewPackageInstalledOptions()

var packageInstalledUpdateCmd = &cobra.Command{
	Use:   "update INSTALLED_PACKAGE_NAME",
	Short: "Update installed package",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Update installed package with name 'mypkg' with some version to version '3.0.0-rc.1' in specified namespace 	
    tanzu package installed update mypkg --version 3.0.0-rc.1 --namespace test-ns`,
	RunE: packageUpdate,
}

func init() {
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledUpdateOp.Version, "version", "v", "", "The version which installed package needs to be updated to")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledUpdateOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file")
	packageInstalledUpdateCmd.Flags().BoolVarP(&packageInstalledUpdateOp.Install, "install", "", false, "Install package if the installed package does not exist, optional")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledUpdateOp.PackageName, "package-name", "p", "", "The public name for the package")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledUpdateOp.Namespace, "namespace", "n", "default", "The namespace to locate the installed package which needs to be updated")
	packageInstalledUpdateCmd.MarkFlagRequired("version") //nolint
	packageInstalledCmd.AddCommand(packageInstalledUpdateCmd)
}

func packageUpdate(_ *cobra.Command, args []string) error {
	packageInstalledUpdateOp.PkgInstallName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageInstalledUpdateOp.KubeConfig)
	if err != nil {
		return err
	}

	if err := pkgClient.UpdatePackageInstall(packageInstalledUpdateOp); err != nil {
		return err
	}

	log.Infof("Updated package '%s' in namespace '%s'", packageInstalledUpdateOp.PkgInstallName, packageInstalledUpdateOp.Namespace)

	return nil
}
