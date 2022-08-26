// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/openapischema"
	"github.com/vmware-tanzu/tanzu-framework/tkg/kappclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
)

var packageAvailableGetCmd = &cobra.Command{
	Use:   "get PACKAGE_NAME or PACKAGE_NAME/VERSION",
	Short: "Get details for an available package",
	Long:  "Get details for an available package or the openAPI schema of a package with a specific version",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Get package details for a package without specifying the version
    tanzu package available get contour.tanzu.vmware.com --namespace test-ns

    # Get package details for a package with specified version 	
    tanzu package available get contour.tanzu.vmware.com/1.15.1-tkg.1-vmware1 --namespace test-ns

    # Get openAPI schema of a package with specified version
    tanzu package available get contour.tanzu.vmware.com/1.15.1-tkg.1-vmware1 --namespace test-ns --values-schema
    
    # Create default values.yaml for a package with specified version based on its openAPI schema
    tanzu package available get contour.tanzu.vmware.com/1.15.1-tkg.1-vmware1 --namespace test-ns --generate-default-values-file`,
	RunE:         packageAvailableGet,
	PreRunE:      validatePackage,
	SilenceUsage: true,
}

func init() {
	packageAvailableGetCmd.Flags().BoolVarP(&packageAvailableOp.ValuesSchema, "values-schema", "", false, "Values schema of the package, optional")
	packageAvailableGetCmd.Flags().BoolVarP(&packageAvailableOp.GenerateDefaultValuesFile, "generate-default-values-file", "", false, "Generate default values from schema of the package, optional")
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
	kc, kcErr := kappclient.NewKappClient(kubeConfig)
	if kcErr != nil {
		return kcErr
	}
	if packageAvailableOp.AllNamespaces {
		packageAvailableOp.Namespace = ""
	}

	if packageAvailableOp.GenerateDefaultValuesFile || packageAvailableOp.ValuesSchema {
		if pkgVersion == "" {
			return errors.New("version is required when --generate-default-values-file or --values-schema flag is true. Please specify <PACKAGE-NAME>/<VERSION>")
		}
	}
	if packageAvailableOp.ValuesSchema {
		if err := getValuesSchemaForPackage(packageAvailableOp.Namespace, pkgName, pkgVersion, kc, cmd.OutOrStdout()); err != nil {
			return err
		}
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
			log.Warningf("package '%s' does not exist in the '%s' namespace", pkgName, packageAvailableOp.Namespace)
			return nil
		}
		return err
	}
	var pkg *kapppkg.Package

	if pkgVersion != "" {
		pkg, err = kc.GetPackage(fmt.Sprintf("%s.%s", pkgName, pkgVersion), packageAvailableOp.Namespace)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.Errorf("package '%s/%s' does not exist in the '%s' namespace", pkgName, pkgVersion, packageAvailableOp.Namespace)
			}
			return err
		}
		t.SetKeys("name", "version", "released-at", "display-name", "short-description", "package-provider", "minimum-capacity-requirements",
			"long-description", "maintainers", "release-notes", "license", "support", "category")
		t.AddRow(pkg.Spec.RefName, pkg.Spec.Version, pkg.Spec.ReleasedAt, pkgMetadata.Spec.DisplayName, pkgMetadata.Spec.ShortDescription,
			pkgMetadata.Spec.ProviderName, pkg.Spec.CapactiyRequirementsDescription, pkgMetadata.Spec.LongDescription, pkgMetadata.Spec.Maintainers,
			pkg.Spec.ReleaseNotes, pkg.Spec.Licenses, pkgMetadata.Spec.SupportDescription, pkgMetadata.Spec.Categories)

		t.RenderWithSpinner()
	} else {
		t.SetKeys("name", "display-name", "short-description", "package-provider", "long-description", "maintainers", "support", "category")
		t.AddRow(pkgMetadata.Name, pkgMetadata.Spec.DisplayName, pkgMetadata.Spec.ShortDescription,
			pkgMetadata.Spec.ProviderName, pkgMetadata.Spec.LongDescription, pkgMetadata.Spec.Maintainers, pkgMetadata.Spec.SupportDescription, pkgMetadata.Spec.Categories)

		t.RenderWithSpinner()
	}

	if packageAvailableOp.GenerateDefaultValuesFile {
		return generateDefaultValuesForPackage(pkg)
	}
	return nil
}

func getValuesSchemaForPackage(namespace, name, version string, kc kappclient.Client, writer io.Writer) error {
	pkg, pkgGetErr := kc.GetPackage(fmt.Sprintf("%s.%s", name, version), namespace)
	if pkgGetErr != nil {
		if apierrors.IsNotFound(pkgGetErr) {
			return errors.Errorf("package '%s/%s' does not exist in the '%s' namespace", name, version, namespace)
		}
		return pkgGetErr
	}

	t, err := component.NewOutputWriterWithSpinner(writer, outputFormat,
		fmt.Sprintf("Retrieving package details for %s/%s...", name, version), true)
	if err != nil {
		return err
	}

	var parseErr error
	if len(pkg.Spec.ValuesSchema.OpenAPIv3.Raw) == 0 {
		t.StopSpinner()
		log.Warningf("package '%s/%s' does not have any user configurable values in the '%s' namespace", pkgName, pkgVersion, packageAvailableOp.Namespace)
		return nil
	}
	dataValuesSchemaParser, parseErr := tkgpackageclient.NewValuesSchemaParser(pkg.Spec.ValuesSchema)
	if parseErr != nil {
		return parseErr
	}
	parsedProperties, parseErr := dataValuesSchemaParser.ParseProperties()
	if parseErr != nil {
		return parseErr
	}

	t.SetKeys("KEY", "DEFAULT", "TYPE", "DESCRIPTION")
	for _, v := range parsedProperties {
		t.AddRow(v.Key, v.Default, v.Type, v.Description)
	}
	t.RenderWithSpinner()

	return nil
}

func generateDefaultValuesForPackage(pkg *kapppkg.Package, valuesFile ...*os.File) error {
	if pkg == nil {
		// should not happen
		return errors.New("pkg is nil")
	}

	if len(pkg.Spec.ValuesSchema.OpenAPIv3.Raw) == 0 {
		log.Warningf("package '%s/%s' does not have any user configurable values in the '%s' namespace", pkgName, pkgVersion, packageAvailableOp.Namespace)
		return nil
	}

	fileNamePrefix := ""
	pkgNameTokens := strings.Split(pkgName, ".")
	if len(pkgNameTokens) >= 1 {
		fileNamePrefix = fmt.Sprintf("%s-", pkgNameTokens[0])
	}
	valuesFileName := fmt.Sprintf("%sdefault-values.yaml", fileNamePrefix)
	var valuesFileToUse *os.File

	if len(valuesFile) > 0 && valuesFile[0] != nil {
		// caller is responsible for closing file
		valuesFileToUse = valuesFile[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		valuesFileToUse, err = os.Create(filepath.Join(cwd, valuesFileName))
		if err != nil {
			return err
		}
		defer valuesFileToUse.Close()
	}

	defaultValues, err := openapischema.SchemaDefault(pkg.Spec.ValuesSchema.OpenAPIv3.Raw)

	if err != nil {
		return err
	}
	_, err = valuesFileToUse.Write(defaultValues)
	if err != nil {
		return err
	}
	log.Infof("\nCreated default values file at %s", valuesFileToUse.Name())
	return nil
}
