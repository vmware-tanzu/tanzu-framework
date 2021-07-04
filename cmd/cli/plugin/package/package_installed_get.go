// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/kappclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

var packageInstalledGetCmd = &cobra.Command{
	Use:   "get INSTALLED_PACKAGE_NAME",
	Short: "Get details for an installed package",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Get package details for installed package with name 'contour-pkg' in specified namespace 	
    tanzu package installed get contour-pkg --namespace test-ns`,
	RunE: packageInstalledGet,
}

func init() {
	packageInstalledGetCmd.Flags().StringVarP(&packageInstalledOp.Namespace, "namespace", "n", "default", "Namespace for installed package CR")
	packageInstalledCmd.AddCommand(packageInstalledGetCmd)
}

func packageInstalledGet(cmd *cobra.Command, args []string) error {
	kc, err := kappclient.NewKappClient(packageAvailableOp.KubeConfig)
	if err != nil {
		return err
	}

	pkgName = args[0]
	packageInstalledOp.PkgInstallName = pkgName
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Retrieving installation details for %s...", pkgName), true)
	if err != nil {
		return err
	}

	pkg, err := kc.GetPackageInstall(packageInstalledOp.PkgInstallName, packageInstalledOp.Namespace)
	if err != nil {
		t.StopSpinner()
		if apierrors.IsNotFound(err) {
			log.Warningf("installed package '%s' does not exist in namespace '%s'", pkgName, packageInstalledOp.Namespace)
			return nil
		}
		return err
	}

	t.AddRow("NAME:", pkg.Name)
	t.AddRow("PACKAGE-NAME:", pkg.Spec.PackageRef.RefName)
	t.AddRow("PACKAGE-VERSION:", pkg.Spec.PackageRef.VersionSelection.Constraints)
	t.AddRow("STATUS:", pkg.Status.FriendlyDescription)
	t.AddRow("CONDITIONS:", pkg.Status.Conditions)
	t.AddRow("USEFUL-ERROR-MESSAGE:", pkg.Status.UsefulErrorMessage)

	t.RenderWithSpinner()

	return nil
}
