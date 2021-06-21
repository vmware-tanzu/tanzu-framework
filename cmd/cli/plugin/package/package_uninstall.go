// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var packageUninstallOp = tkgpackagedatamodel.NewPackageUninstallOptions()

var packageUninstallCmd = &cobra.Command{
	Use:   "uninstall INSTALL_NAME",
	Short: "Uninstall a package",
	Long:  "Remove the installed package and almost all resources installed as part of installation of the package from the cluster. Namespaces created during installation of the package, do not automatically get deleted at the time of package uninstallation.",
	Args:  cobra.ExactArgs(1),
	RunE:  packageUninstall,
}

func init() {
	packageUninstallCmd.Flags().StringVarP(&packageUninstallOp.Namespace, "namespace", "n", "default", "Target namespace from which the package should be deleted, optional")
	packageUninstallCmd.Flags().StringVarP(&packageUninstallOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	packageUninstallCmd.Flags().DurationVarP(&packageUninstallOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package deletion status, optional")
	packageUninstallCmd.Flags().DurationVarP(&packageUninstallOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package deletion status, optional")
}

func packageUninstall(_ *cobra.Command, args []string) error {
	packageUninstallOp.InstalledPkgName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageUninstallOp.KubeConfig)
	if err != nil {
		return err
	}

	if err := pkgClient.UninstallPackage(packageUninstallOp); err != nil {
		return err
	}

	log.Infof("Uninstalled package '%s' from namespace '%s'\n", packageUninstallOp.InstalledPkgName, packageUninstallOp.Namespace)

	return nil
}
