// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

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
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.Version, "version", "v", "", "The version which installed package needs to be updated to")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file")
	packageInstalledUpdateCmd.Flags().BoolVarP(&packageInstalledOp.Install, "install", "", false, "Install package if the installed package does not exist, optional")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.PackageName, "package-name", "p", "", "The public name for the package")
	packageInstalledUpdateCmd.Flags().StringVarP(&packageInstalledOp.Namespace, "namespace", "n", "default", "The namespace to locate the installed package which needs to be updated")
	packageInstalledUpdateCmd.MarkFlagRequired("version") //nolint
	packageInstalledCmd.AddCommand(packageInstalledUpdateCmd)
}

func packageUpdate(_ *cobra.Command, args []string) error {
	packageInstalledOp.PkgInstallName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageInstalledOp.KubeConfig)
	if err != nil {
		return err
	}

	pp := &tkgpackagedatamodel.PackageProgress{
		ProgressMsg: make(chan string, 10),
		Err:         make(chan error),
		Done:        make(chan struct{}),
		Success:     make(chan bool),
	}
	go pkgClient.UpdatePackage(packageInstalledOp, pp)

	initialMsg := fmt.Sprintf("Updating package '%s'", packageInstalledOp.PkgInstallName)
	successMsg := fmt.Sprintf("Updated package install '%s' in namespace '%s'", packageInstalledOp.PkgInstallName, packageInstalledOp.Namespace)
	if err := displayProgress(initialMsg, successMsg, pp); err != nil {
		return err
	}

	return nil
}
