// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageUpdateOp = tkgpackagedatamodel.NewPackageOptions()

var packageUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "package update",
	Args:  cobra.ExactArgs(1),
	RunE:  packageUpdate,
}

func init() {
	packageUpdateCmd.Flags().StringVarP(&packageUpdateOp.Version, "version", "v", "", "The version which installed package needs to be updated to")
	packageUpdateCmd.Flags().StringVarP(&packageUpdateOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file")
	packageUpdateCmd.Flags().BoolVarP(&packageUpdateOp.Install, "install", "", false, "Install package if the installed package does not exist, optional")
	packageUpdateCmd.Flags().StringVarP(&packageUpdateOp.PackageName, "package-name", "p", "", "The public name for the package")
	packageUpdateCmd.Flags().StringVarP(&packageUpdateOp.Namespace, "namespace", "n", "default", "The namespace to locate the installed package which needs to be updated")
	packageUpdateCmd.Flags().StringVarP(&packageUpdateOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	packageUpdateCmd.MarkFlagRequired("version") //nolint
}

func packageUpdate(_ *cobra.Command, args []string) error {
	packageUpdateOp.PkgInstallName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageUpdateOp.KubeConfig)
	if err != nil {
		return err
	}

	if err := pkgClient.UpdatePackage(packageUpdateOp); err != nil {
		return err
	}

	log.Infof("Updated package '%s' in namespace '%s'", packageUpdateOp.PkgInstallName, packageUpdateOp.Namespace)

	return nil
}
