// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"

	"github.com/jeremywohl/flatten"
)

var packageAvailableGetCmd = &cobra.Command{
	Use:   "get PACKAGE_NAME or PACKAGE_NAME/VERSION",
	Short: "get package detail",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Get package details for a package with specified version 	
    tanzu package available get contour.tanzu.vmware.com/1.15.1-tkg.1-vmware1 --namespace test-ns`,
	RunE:    packageAvailableGet,
	PreRunE: validatePackage,
}

func init() {
	packageAvailableGetCmd.Flags().BoolVarP(&packageAvailableOp.ValuesSchema, "values-schema", "", false, "Values schema of the package")
	packageAvailableCmd.AddCommand(packageAvailableGetCmd)
}

var pkgName string
var pkgVersion string

func validatePackage(cmd *cobra.Command, args []string) error {
	pkgNameVersion := strings.Split(args[0], "/")
	if len(pkgNameVersion) == 2 {
		pkgName = pkgNameVersion[0]
		pkgVersion = pkgNameVersion[1]
	} else if len(pkgNameVersion) == 1 {
		pkgName = pkgNameVersion[0]
	} else {
		return fmt.Errorf("package should be of the format name or name/version")
	}
	return nil
}

func packageAvailableGet(cmd *cobra.Command, args []string) error { //nolint:gocyclo
	kc, err := kappclient.NewKappClient(packageAvailableOp.KubeConfig)
	if err != nil {
		return err
	}
	if packageAvailableOp.AllNamespaces {
		packageAvailableOp.Namespace = ""
	}
	if packageAvailableOp.ValuesSchema {
		pkg, err := kc.GetPackage(fmt.Sprintf("%s.%s", pkgName, pkgVersion), packageAvailableOp.Namespace)
		if err != nil {
			return err
		}

		if len(pkg.Spec.ValuesSchema.OpenAPIv3.Raw) == 0 {
			log.Infof("failed to find package '%s' values schema", pkgName)
			return nil
		}

		s := string(pkg.Spec.ValuesSchema.OpenAPIv3.Raw)
		sflat, err := flatten.FlattenString(s, "", flatten.DotStyle)
		if err != nil {
			return err
		}
		strim := strings.Trim(sflat, "{}")
		srep := strings.Replace(strim, "\"", "", -1)
		entries := strings.Split(srep, ",")

		t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
			fmt.Sprintf("Retrieving package details for %s...", args[0]), true)
		if err != nil {
			return err
		}

		t.SetKeys("KEY", "DEFAULT", "TYPE", "DESCRIPTION")
		for _, e := range entries {
			parts := strings.Split(e, ":")
			if strings.Contains(parts[0], "description") {
				t.AddRow(parts[0], "", "", parts[1])
			} else if strings.Contains(parts[0], "type") {
				t.AddRow(parts[0], "", parts[1], "")
			} else {
				t.AddRow(parts[0], parts[1], "", "")
			}
		}
		t.RenderWithSpinner()
		return nil
	}

	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), getOutputFormat(),
		fmt.Sprintf("Retrieving package details for %s...", args[0]), true)
	if err != nil {
		return err
	}

	pkgMetadata, err := kc.GetPackageMetadataByName(pkgName, packageAvailableOp.Namespace)
	if err != nil {
		t.StopSpinner()
		if apierrors.IsNotFound(err) {
			log.Warningf("package '%s' does not exist in namespace '%s'", pkgName, packageAvailableOp.Namespace)
			return nil
		}
		return err
	}

	if pkgVersion != "" {
		pkg, err := kc.GetPackage(fmt.Sprintf("%s.%s", pkgName, pkgVersion), packageAvailableOp.Namespace)
		if err != nil {
			return err
		}
		t.SetKeys("name", "version", "released-at", "display-name", "short-description", "package-provider", "minimum-capacity-requirements",
			"long-description", "maintainers", "release-notes", "license")
		t.AddRow(pkg.Spec.RefName, pkg.Spec.Version, pkg.Spec.ReleasedAt, pkgMetadata.Spec.DisplayName, pkgMetadata.Spec.ShortDescription,
			pkgMetadata.Spec.ProviderName, pkg.Spec.CapactiyRequirementsDescription, pkgMetadata.Spec.LongDescription, pkgMetadata.Spec.Maintainers,
			pkg.Spec.ReleaseNotes, pkg.Spec.Licenses)

		t.RenderWithSpinner()
	} else {
		t.SetKeys("name", "display-name", "short-description", "package-provider", "long-description", "maintainers")
		t.AddRow(pkgMetadata.Name, pkgMetadata.Spec.DisplayName, pkgMetadata.Spec.ShortDescription,
			pkgMetadata.Spec.ProviderName, pkgMetadata.Spec.LongDescription, pkgMetadata.Spec.Maintainers)

		t.RenderWithSpinner()
	}
	return nil
}
