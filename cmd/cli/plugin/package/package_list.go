// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
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
	packageListCmd.Flags().BoolVarP(&packageListOp.Available, "available", "a", false, "If present, show package CRs. Show all the available packages if value is blank else show all the package CRs for specified name with all versions.")
	packageListCmd.Flags().BoolVarP(&packageListOp.AllNamespaces, "all-namespaces", "A", false, "If present, list the installed package(s) across all namespaces.")
	packageListCmd.Flags().StringVarP(&packageListOp.Namespace, "namespace", "n", "default", "The namespace from which to list installed packages")
	packageListCmd.Flags().StringVarP(&packageListOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
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
		installedPackageList, err := pkgClient.ListInstalledPackages(packageListOp)
		if err != nil {
			return errors.Wrap(err, "failed to list installed packages")
		}
		if packageListOp.Namespace != "" {
			// List Installed packages in given namespace
			t := component.NewTableWriter("NAME", "VERSION")
			for _, pkg := range installedPackageList.Items { //nolint:gocritic
				t.Append([]string{pkg.Name, pkg.Status.Version})
			}
			t.Render()
		} else {
			// List Installed packages across all namespaces
			t := component.NewTableWriter("NAME", "VERSION", "NAMESPACE")
			for _, pkg := range installedPackageList.Items { //nolint:gocritic
				t.Append([]string{pkg.Name, pkg.Status.Version, pkg.GetNamespace()})
			}
			t.Render()
		}
	} else if packageListOp.Available && packageListOp.PackageName == "" {
		packageList, err := pkgClient.ListPackages()
		if err != nil {
			return err
		}
		// List all available Package CRs
		t := component.NewTableWriter("NAME", "DISPLAYNAME", "SHORTDESCRIPTION")
		for _, pkg := range packageList.Items { //nolint:gocritic
			t.Append([]string{pkg.Name, pkg.Spec.DisplayName, pkg.Spec.ShortDescription})
		}
		t.Render()
	} else if packageListOp.Available && packageListOp.PackageName != "" {
		packageVersionList, err := pkgClient.ListPackageVersions(packageListOp)
		if err != nil {
			return errors.Wrap(err, "failed to list package versions")
		}
		// List all available Package CRs
		t := component.NewTableWriter("NAME", "VERSION")
		for _, pkg := range packageVersionList.Items { //nolint:gocritic
			t.Append([]string{packageListOp.PackageName, pkg.Spec.Version})
		}
		t.Render()
	}

	return nil
}
