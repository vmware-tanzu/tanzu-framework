// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aunum/log"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

var packageInstalledGetCmd = &cobra.Command{
	Use:   "get INSTALLED_PACKAGE_NAME",
	Short: "Get details for an installed package",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Get package details for installed package with name 'contour-pkg' in specified namespace 	
    tanzu package installed get contour-pkg --namespace test-ns`,
	RunE:         packageInstalledGet,
	SilenceUsage: true,
}

func init() {
	packageInstalledGetCmd.Flags().StringVarP(&packageInstalledOp.Namespace, "namespace", "n", "default", "Namespace for installed package CR, optional")
	packageInstalledGetCmd.Flags().StringVarP(&packageInstalledOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file, optional")
	packageInstalledGetCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table), optional")
	packageInstalledCmd.AddCommand(packageInstalledGetCmd)
}

func packageInstalledGet(cmd *cobra.Command, args []string) error {
	kc, err := kappclient.NewKappClient(kubeConfig)
	if err != nil {
		return err
	}

	pkgName = args[0]
	packageInstalledOp.PkgInstallName = pkgName
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), getOutputFormat(),
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

	if packageInstalledOp.ValuesFile != "" {
		packageInstalledOp.SecretName = fmt.Sprintf(packagedatamodel.SecretName, packageInstalledOp.PkgInstallName, packageInstalledOp.Namespace)
		f, err := os.Create(packageInstalledOp.ValuesFile)
		if err != nil {
			return err
		}
		defer f.Close()
		w := bufio.NewWriter(f)

		dataValue := ""
		for _, value := range pkg.Spec.Values {
			if value.SecretRef == nil {
				continue
			}
			s, err := kc.GetSecretValue(value.SecretRef.Name, packageInstalledOp.Namespace)
			if err != nil {
				return err
			}

			stringValue := string(s)
			if len(stringValue) < 3 {
				dataValue += packagedatamodel.YamlSeparator
				dataValue += "\n"
			}
			if len(stringValue) >= 3 && stringValue[:3] != packagedatamodel.YamlSeparator {
				dataValue += packagedatamodel.YamlSeparator
				dataValue += "\n"
			}
			dataValue += string(s)
		}
		if _, err = fmt.Fprintf(w, "%s", dataValue); err != nil {
			return err
		}
		w.Flush()
		return nil
	}

	t.SetKeys("name", "package-name", "package-version", "status", "conditions", "useful-error-message")
	t.AddRow(pkg.Name, pkg.Spec.PackageRef.RefName, pkg.Status.Version,
		pkg.Status.FriendlyDescription, pkg.Status.Conditions, pkg.Status.UsefulErrorMessage)

	t.RenderWithSpinner()

	return nil
}
