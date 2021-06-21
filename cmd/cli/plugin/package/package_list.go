// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageListOp = tkgpackagedatamodel.NewPackageListOptions()

var packageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages based on supplied flags",
	Args:  cobra.MaximumNArgs(1),
	RunE:  packageList,
}

func init() {
	packageListCmd.Flags().BoolVarP(&packageListOp.Available, "available", "a", false, "If present, show PackageMetadata CRs. Show all the available packages if value is blank else show all the PackageMetadata CRs for specified name with all package versions.")
	packageListCmd.Flags().BoolVarP(&packageListOp.AllNamespaces, "all-namespaces", "A", false, "If present, list the installed package(s) across all namespaces.")
	packageListCmd.Flags().StringVarP(&packageListOp.Namespace, "namespace", "n", "default", "The namespace from which to list installed packages")
	packageListCmd.Flags().StringVarP(&packageListOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	packageListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

func packageList(cmd *cobra.Command, args []string) error { //nolint:gocyclo
	if (cmd.Flags().NFlag() == 1 && packageListOp.Namespace == "--available") ||
		(cmd.Flags().NFlag() == 1 && packageListOp.Available && len(args) == 1 && args[0] == "--namespace") {
		return errors.New("available and namespace flags cannot be used together. Usage: tanzu package list --help")
	}

	if packageListOp.Available && len(args) == 1 && args[0] != "-A" {
		packageListOp.PackageName = args[0]
	} else if cmd.Flags().NFlag() == 0 || !packageListOp.Available || packageListOp.AllNamespaces || packageListOp.KubeConfig != "" {
		packageListOp.ListInstalled = true
		if packageListOp.AllNamespaces {
			packageListOp.Namespace = ""
		}
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageListOp.KubeConfig)
	if err != nil {
		return err
	}

	if packageListOp.ListInstalled {
		installedPackageList, err := pkgClient.ListPackageInstalls(packageListOp)
		if err != nil {
			return errors.Wrap(err, "failed to list installed packages")
		}
		t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat)
		if outputFormat == "" && len(installedPackageList.Items) == 0 {
			if packageListOp.AllNamespaces || packageListOp.Namespace == "" {
				log.Infof("No packages installed in any namespace")
			} else {
				log.Infof("No packages installed in namespace '%s'\n", packageListOp.Namespace)
			}
		}
		if packageListOp.Namespace != "" {
			// List Installed packages in given namespace
			t.SetKeys("NAME", "VERSION")
			for _, pkg := range installedPackageList.Items { //nolint:gocritic
				t.AddRow(pkg.Name, pkg.Status.Version)
			}
			t.Render()
		} else {
			// List Installed packages across all namespaces
			t.SetKeys("NAME", "VERSION", "NAMESPACE")
			for _, pkg := range installedPackageList.Items { //nolint:gocritic
				t.AddRow(pkg.Name, pkg.Status.Version, pkg.GetNamespace())
			}
			t.Render()
		}
		return nil
	}
	if packageListOp.Available {
		if packageListOp.PackageName == "" { // Show unique packages
			packageMetadataList, err := pkgClient.ListPackageMetadata(packageListOp)
			if err != nil {
				return err
			}
			// List all available Package CRs
			t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "DISPLAYNAME", "SHORTDESCRIPTION", "NAMESPACE")
			for _, pkg := range packageMetadataList.Items { //nolint:gocritic
				t.AddRow(pkg.Name, pkg.Spec.DisplayName, pkg.Spec.ShortDescription, pkg.Namespace)
			}
			t.Render()
		} else { // List all versions for specific package
			packages, err := pkgClient.ListPackages(packageListOp)
			if err != nil {
				return errors.Wrap(err, "failed to list package versions")
			}
			// List all available Package CRs
			t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "VERSION", "NAMESPACE")

			for _, pkg := range packages.Items { //nolint:gocritic
				t.AddRow(packageListOp.PackageName, pkg.Spec.Version, pkg.Namespace)
			}
			t.Render()
		}
		return nil
	}
	return nil
}
