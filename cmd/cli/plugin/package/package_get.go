// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageGetOp = tkgpackagedatamodel.NewPackageGetOptions()

var packageGetCmd = &cobra.Command{
	Use:   "get PACKAGE_NAME",
	Short: "Get details for a package or installed package",
	RunE:  packageGet,
}

func init() {
	packageGetCmd.Flags().StringVarP(&packageGetOp.Available, "available", "", "", "The name for the package")
	packageGetCmd.Flags().StringVarP(&packageGetOp.Version, "version", "v", "", "The version of the package")
	packageGetCmd.Flags().StringVarP(&packageGetOp.ValuesFile, "values-file", "f", "", "Show configuration values")
	packageGetCmd.Flags().StringVarP(&packageGetOp.Namespace, "namespace", "n", "default", "Namespace for installed package CR")
	packageGetCmd.Flags().StringVarP(&packageGetOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	packageGetCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

func packageGet(cmd *cobra.Command, args []string) error {
	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageGetOp.KubeConfig)
	if err != nil {
		return err
	}

	if packageGetOp.Available != "" {
		packageGetOp.PackageName = packageGetOp.Available
		if packageGetOp.Version == "" {
			return errors.New("version is required if available is set. Usage: tanzu package get --help")
		}
		// Get metadata of package
		pkg, pkgVersion, err := pkgClient.GetPackage(packageGetOp.PackageName, packageGetOp.Version, packageGetOp.Namespace)
		if err != nil {
			return err
		}
		t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat)
		t.SetKeys("NAME", "VERSION", "PACKAGEPROVIDER", "MINIMUMCAPACITYREQUIREMENTS", "SHORTDESCRIPTION")
		t.AddRow(pkg.Name, pkgVersion.Spec.Version, pkg.Spec.ProviderName, pkgVersion.Spec.CapactiyRequirementsDescription, pkg.Spec.ShortDescription)
		t.Render()
	} else {
		if len(args) != 0 {
			packageGetOp.PackageName = args[0]
		} else {
			return errors.New("incorrect number of input parameters. Usage: tanzu package get PACKAGE_NAME [FLAGS]")
		}
		if packageGetOp.Version != "" {
			return errors.New("version flag can't be use if available flag is not set. Usage: tanzu package get --help")
		}
		// Get metadata of installed package
		pkg, err := pkgClient.GetPackageInstall(packageGetOp)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to find installed package '%s' in namespace '%s'", packageGetOp.PackageName, packageGetOp.Namespace))
		}
		t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat)
		t.SetKeys("NAME", "VERSION", "NAMESPACE", "STATUS", "REASON")
		t.AddRow(pkg.Name, pkg.Status.Version, pkg.Namespace, pkg.Status.FriendlyDescription, pkg.Status.UsefulErrorMessage)
		t.Render()
	}
	return nil
}
