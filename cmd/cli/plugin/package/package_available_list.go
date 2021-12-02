// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/k14s/semver/v4"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"

	"go.uber.org/multierr"
)

var packageAvailableListCmd = &cobra.Command{
	Use:   "list or list PACKAGE_NAME",
	Short: "List available packages",
	Args:  cobra.MaximumNArgs(1),
	Example: `
    # List available packages across all namespaces 	
    tanzu package available list -A
	
    # List all versions for available package from specified namespace	
    tanzu package available list contour.tanzu.vmware.com --namespace test-ns`,
	RunE:         packageAvailableList,
	SilenceUsage: true,
}

func init() {
	packageAvailableListCmd.Flags().BoolVarP(&packageAvailableOp.AllNamespaces, "all-namespaces", "A", false, "If present, list packages across all namespaces, optional")
	packageAvailableCmd.AddCommand(packageAvailableListCmd)
}

func packageAvailableList(cmd *cobra.Command, args []string) error {
	kc, err := kappclient.NewKappClient(packageAvailableOp.KubeConfig)
	if err != nil {
		return err
	}
	if packageAvailableOp.AllNamespaces {
		packageAvailableOp.Namespace = ""
	}

	if len(args) == 0 {
		t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
			"Retrieving available packages...", true)
		if err != nil {
			return err
		}
		packageMetadataList, err := kc.ListPackageMetadata(packageAvailableOp.Namespace)
		if err != nil {
			return err
		}
		if packageAvailableOp.AllNamespaces {
			t.SetKeys("NAME", "DISPLAY-NAME", "SHORT-DESCRIPTION", "LATEST-VERSION", "NAMESPACE")
		} else {
			t.SetKeys("NAME", "DISPLAY-NAME", "SHORT-DESCRIPTION", "LATEST-VERSION")
		}
		var multiGetVerErr error
		for i := range packageMetadataList.Items {
			pkg := packageMetadataList.Items[i]
			latestVersion, getVerErr := getPackageLatestVersion(pkg.Name, pkg.Namespace, kc)
			// It is safe to use multierr.Append() even when getVerErr is nil. The API handles that case.
			multiGetVerErr = multierr.Append(multiGetVerErr, getVerErr)
			if packageAvailableOp.AllNamespaces {
				t.AddRow(pkg.Name, pkg.Spec.DisplayName, pkg.Spec.ShortDescription, latestVersion, pkg.Namespace)
			} else {
				t.AddRow(pkg.Name, pkg.Spec.DisplayName, pkg.Spec.ShortDescription, latestVersion)
			}
		}
		t.RenderWithSpinner()

		// If there are any errors when getting the latest package version, we do not want to error out the command and
		// block user from moving forward. When an error occurs, the latest version will be empty. What we want is to
		// show a warning message to explain what the empty latest version means.
		if multiGetVerErr != nil {
			log.Warning("\nUnable to get the latest version for all packages, some packages' latest version fields" +
				" are empty. Please try again or use `tanzu package available list <package-name>` to get the version.")
		}
		return nil
	}
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Retrieving package versions for %s...", args[0]), true)
	if err != nil {
		return err
	}
	packageAvailableOp.PackageName = args[0]
	pkgs, err := kc.ListPackages(packageAvailableOp.PackageName, packageAvailableOp.Namespace)
	if err != nil {
		return err
	}
	if packageAvailableOp.AllNamespaces {
		t.SetKeys("NAME", "VERSION", "RELEASED-AT", "NAMESPACE")
	} else {
		t.SetKeys("NAME", "VERSION", "RELEASED-AT")
	}
	for i := range pkgs.Items {
		pkg := pkgs.Items[i]
		if packageAvailableOp.AllNamespaces {
			t.AddRow(pkg.Spec.RefName, pkg.Spec.Version, pkg.Spec.ReleasedAt, pkg.Namespace)
		} else {
			t.AddRow(pkg.Spec.RefName, pkg.Spec.Version, pkg.Spec.ReleasedAt)
		}
	}
	t.RenderWithSpinner()
	return nil
}

// getPackageLatestVersion returns the latest version of a particular package under a namespace
func getPackageLatestVersion(packageName, namespace string, kc kappclient.Client) (string, error) {
	pkgs, listErr := kc.ListPackages(packageName, namespace)
	if listErr != nil {
		return "", listErr
	}
	var latest *semver.Version
	var parseErr error
	for i := range pkgs.Items {
		pkg := pkgs.Items[i]
		current, err := semver.Make(pkg.Spec.Version)
		if err != nil {
			// If we are not able to parse the version of current package, record the error and continue
			parseErr = multierr.Append(parseErr, err)
			continue
		}
		if latest == nil {
			latest = &current
		} else if latest.LT(current) {
			latest = &current
		}
	}

	// If we are not able to get and compare versions from all packages, latest is nil
	if latest == nil {
		return "", parseErr
	}
	return latest.String(), nil
}
