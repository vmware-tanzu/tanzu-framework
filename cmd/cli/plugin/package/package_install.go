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

var packageInstallOp = tkgpackagedatamodel.NewPackageInstalledOptions()

var packageInstallCmd = &cobra.Command{
	Use:   "install INSTALL_NAME",
	Short: "Install a package",
	Args:  cobra.ExactArgs(1),
	RunE:  packageInstall,
}

func init() {
	packageInstallCmd.Flags().StringVarP(&packageInstallOp.PackageName, "package-name", "p", "", "Name of the package to be installed")
	packageInstallCmd.Flags().StringVarP(&packageInstallOp.Version, "version", "v", "", "Version of the package to be installed")
	packageInstallCmd.Flags().BoolVarP(&packageInstallOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	packageInstallCmd.Flags().StringVarP(&packageInstallOp.Namespace, "namespace", "n", "default", "Target namespace to install the package, optional")
	packageInstallCmd.Flags().StringVarP(&packageInstallOp.ServiceAccountName, "service-account-name", "", "", "Name of an existing service account used to install underlying package contents, optional")
	packageInstallCmd.Flags().StringVarP(&packageInstallOp.ValuesFile, "values-file", "f", "", "The path to the configuration values file, optional")
	packageInstallCmd.Flags().StringVarP(&packageInstallOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	packageInstallCmd.Flags().BoolVarP(&packageInstallOp.Wait, "wait", "", true, "Wait for the package reconciliation to complete, optional")
	packageInstallCmd.Flags().DurationVarP(&packageInstallOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package reconciliation status, optional")
	packageInstallCmd.Flags().DurationVarP(&packageInstallOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package reconciliation status, optional")
	packageInstallCmd.MarkFlagRequired("package-name") //nolint
	packageInstallCmd.MarkFlagRequired("version")      //nolint
}

func packageInstall(_ *cobra.Command, args []string) error {
	packageInstallOp.PkgInstallName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(packageInstallOp.KubeConfig)
	if err != nil {
		return err
	}

	pp := &tkgpackagedatamodel.PackageProgress{
		ProgressMsg: make(chan string, 10),
		Err:         make(chan error),
		Done:        make(chan struct{}),
	}
	go pkgClient.InstallPackageWithProgress(packageInstallOp, pp)

	if err := displayInstallProgress(fmt.Sprintf("Installing package %s", packageInstallOp.PackageName), pp); err != nil {
		return err
	}
	log.Infof("Added installed package '%s' in namespace '%s'\n", packageInstallOp.PkgInstallName, packageInstallOp.Namespace)
	return nil
}

func displayInstallProgress(initialMsg string, pp *tkgpackagedatamodel.PackageProgress) error {
	var currMsg string

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	if err := s.Color("bgBlack", "bold", "fgWhite"); err != nil {
		return err
	}
	s.Suffix = fmt.Sprintf(" %s", initialMsg)
	s.Start()

	defer func() {
		if s.Active() {
			s.Stop()
		}
	}()
	for {
		select {
		case err := <-pp.Err:
			s.FinalMSG = fmt.Sprintf("%s\n", err.Error())
			return err
		case msg := <-pp.ProgressMsg:
			if msg != currMsg {
				log.Infof("\n")
				s.Suffix = fmt.Sprintf(" %s", msg)
				currMsg = msg
			}
		case <-pp.Done:
			for msg := range pp.ProgressMsg {
				if msg != currMsg {
					log.Infof("\n")
					s.Suffix = fmt.Sprintf(" %s", msg)
					currMsg = msg
				}
			}
			return nil
		}
	}
}
