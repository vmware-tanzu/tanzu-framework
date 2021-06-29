// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/kappclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
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

func packageAvailableGet(cmd *cobra.Command, args []string) error {
	kc, err := kappclient.NewKappClient(packageAvailableOp.KubeConfig)
	if err != nil {
		return err
	}
	if packageAvailableOp.AllNamespaces {
		packageAvailableOp.Namespace = ""
	}
	if packageAvailableOp.ValuesSchema {
		// TODO
		return fmt.Errorf("unimplemented")
	}

	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Retrieving package details for %s...", args[0]), true)
	if err != nil {
		return err
	}

	pkgMetadata, err := kc.GetPackageMetadataByName(pkgName, packageAvailableOp.Namespace)
	if err != nil {
		t.StopSpinner()
		if apierrors.IsNotFound(err) {
			log.Infof("failed to find package '%s'", pkgName)
		} else {
			return err
		}
	}

	if pkgVersion != "" {
		pkg, err := kc.GetPackage(fmt.Sprintf("%s.%s", pkgName, pkgVersion), packageAvailableOp.Namespace)
		if err != nil {
			return err
		}
		t.AddRow("NAME:", pkg.Spec.RefName)
		t.AddRow("VERSION:", pkg.Spec.Version)
		t.AddRow("RELEASED-AT:", pkg.Spec.ReleasedAt)
		t.AddRow("DISPLAY-NAME:", pkgMetadata.Spec.DisplayName)
		t.AddRow("SHORT-DESCRIPTION:", pkgMetadata.Spec.ShortDescription)
		t.AddRow("PACKAGE-PROVIDER:", pkgMetadata.Spec.ProviderName)
		t.AddRow("MINIMUM-CAPACITY-REQUIREMENTS:", pkg.Spec.CapactiyRequirementsDescription)
		t.AddRow("LONG-DESCRIPTION:", pkgMetadata.Spec.LongDescription)
		t.AddRow("MAINTAINERS:", pkgMetadata.Spec.Maintainers)
		t.AddRow("RELEASE-NOTES:", pkg.Spec.ReleaseNotes)
		t.AddRow("LICENSE:", pkg.Spec.Licenses)

		t.RenderWithSpinner()
	} else {
		t.AddRow("NAME:", pkgMetadata.Name)
		t.AddRow("DISPLAY-NAME:", pkgMetadata.Spec.DisplayName)
		t.AddRow("SHORT-DESCRIPTION:", pkgMetadata.Spec.ShortDescription)
		t.AddRow("PACKAGE-PROVIDER:", pkgMetadata.Spec.ProviderName)
		t.AddRow("LONG-DESCRIPTION:", pkgMetadata.Spec.LongDescription)
		t.AddRow("MAINTAINERS:", pkgMetadata.Spec.Maintainers)

		t.RenderWithSpinner()
	}
	return nil
}
