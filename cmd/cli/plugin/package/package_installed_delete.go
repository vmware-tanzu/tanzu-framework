// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var packageInstalledDeleteCmd = &cobra.Command{
	Use:   "delete INSTALLED_PACKAGE_NAME",
	Short: "Delete an installed package",
	Long:  "Remove the installed package and almost all resources installed as part of installation of the package from the cluster. Namespaces created during installation of the package, do not automatically get deleted at the time of package uninstallation.",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Delete installed package with name 'contour-pkg' from specified namespace 	
    tanzu package installed delete contour-pkg -n test-ns`,
	RunE:         packageUninstall,
	SilenceUsage: true,
}

func init() {
	packageInstalledDeleteCmd.Flags().StringVarP(&packageInstalledOp.Namespace, "namespace", "n", "default", "Target namespace from which the package should be deleted, optional")
	packageInstalledDeleteCmd.Flags().DurationVarP(&packageInstalledOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package deletion status, optional")
	packageInstalledDeleteCmd.Flags().DurationVarP(&packageInstalledOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package deletion status, optional")
	packageInstalledDeleteCmd.Flags().BoolVarP(&packageInstalledOp.SkipPrompt, "yes", "y", false, "Delete installed package without asking for confirmation, optional")
	packageInstalledCmd.AddCommand(packageInstalledDeleteCmd)
}

func packageUninstall(cmd *cobra.Command, args []string) error {
	packageInstalledOp.PkgInstallName = args[0]

	if !packageInstalledOp.SkipPrompt {
		if err := cli.AskForConfirmation(fmt.Sprintf("Deleting installed package '%s' in namespace '%s'. Are you sure?",
			packageInstalledOp.PkgInstallName, packageInstalledOp.Namespace)); err != nil {
			return err
		}
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	return pkgClient.UninstallPackageSync(packageInstalledOp)
}
