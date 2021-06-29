// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageInstalledDeleteOp = tkgpackagedatamodel.NewPackageUninstallOptions()

var packageInstalledDeleteCmd = &cobra.Command{
	Use:   "delete INSTALLED_PACKAGE_NAME",
	Short: "Delete an installed package",
	Long:  "Remove the installed package and almost all resources installed as part of installation of the package from the cluster. Namespaces created during installation of the package, do not automatically get deleted at the time of package uninstallation.",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Delete installed package with name 'contour-pkg' from specified namespace 	
    tanzu package installed delete contour-pkg -n test-ns`,
	RunE: packageUninstall,
}

func init() {
	packageInstalledDeleteCmd.Flags().StringVarP(&packageInstalledDeleteOp.Namespace, "namespace", "n", "default", "Target namespace from which the package should be deleted, optional")
	packageInstalledDeleteCmd.Flags().DurationVarP(&packageInstalledDeleteOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package deletion status, optional")
	packageInstalledDeleteCmd.Flags().DurationVarP(&packageInstalledDeleteOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package deletion status, optional")
	packageInstalledCmd.AddCommand(packageInstalledDeleteCmd)
}

func packageUninstall(_ *cobra.Command, args []string) error {
	packageInstalledDeleteOp.PkgInstallName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageInstalledDeleteOp.KubeConfig)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	if err := s.Color("bgBlack", "bold", "fgWhite"); err != nil {
		return err
	}
	s.Suffix = fmt.Sprintf(" %s", fmt.Sprintf("Deleting installed package '%s' from namespace '%s'",
		packageInstalledDeleteOp.PkgInstallName, packageInstalledDeleteOp.Namespace))
	s.Start()

	found, err := pkgClient.UninstallPackage(packageInstalledDeleteOp)
	s.Stop()
	if !found {
		log.Infof("Installed package '%s' not found in namespace '%s'\n", packageInstalledDeleteOp.PkgInstallName, packageInstalledDeleteOp.Namespace)
		return nil
	}
	if err != nil {
		return err
	}
	log.Infof("Deleted installed package '%s' from namespace '%s'\n", packageInstalledDeleteOp.PkgInstallName, packageInstalledDeleteOp.Namespace)

	return nil
}
