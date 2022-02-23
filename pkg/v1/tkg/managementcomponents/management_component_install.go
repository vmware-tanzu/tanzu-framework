// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

// ClusterOptions specifies cluster configuration
type ClusterOptions struct {
	Kubeconfig  string
	Kubecontext string
}

// ManagementPackageRepositoryOptions specifies management package repository deployment options
type ManagementPackageRepositoryOptions struct {
	ManagementPackageRepoImage string
	TKGPackageValuesFile       string
}

// KappControllerOptions specifies kapp-controller deployment options
type KappControllerOptions struct {
	KappControllerConfigFile       string
	KappControllerInstallNamespace string
}

// ManagementComponentsInstallOptions specifies install options for management components
type ManagementComponentsInstallOptions struct {
	ClusterOptions                     ClusterOptions
	ManagementPackageRepositoryOptions ManagementPackageRepositoryOptions
	KappControllerOptions              KappControllerOptions
}

// InstallManagementComponents installs the management component to cluster
func InstallManagementComponents(mcip *ManagementComponentsInstallOptions) error {
	clusterClient, err := clusterclient.NewClient(mcip.ClusterOptions.Kubeconfig, mcip.ClusterOptions.Kubecontext, clusterclient.Options{})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}
	if err := InstallKappController(clusterClient, mcip.KappControllerOptions); err != nil {
		return errors.Wrap(err, "unable to install kapp-controller")
	}

	// create package client
	pkgClient, err := tkgpackageclient.NewTKGPackageClientForContext(mcip.ClusterOptions.Kubeconfig, mcip.ClusterOptions.Kubecontext)
	if err != nil {
		return err
	}
	if err = InstallManagementPackages(pkgClient, mcip.ManagementPackageRepositoryOptions); err != nil {
		return errors.Wrap(err, "unable to install management packages")
	}
	return nil
}

// InstallKappController installs kapp-controller to the cluster
func InstallKappController(clusterClient clusterclient.Client, kappControllerOptions KappControllerOptions) error {
	// Apply kapp-controller configuration
	err := clusterClient.ApplyFile(kappControllerOptions.KappControllerConfigFile)
	if err != nil {
		return errors.Wrapf(err, "error installing %s", constants.KappControllerDeploymentName)
	}
	// Wait for kapp-controller to be deployed and running
	err = clusterClient.WaitForDeployment(constants.KappControllerDeploymentName, kappControllerOptions.KappControllerInstallNamespace)
	if err != nil {
		return errors.Wrapf(err, "error while waiting for deployment %s", constants.KappControllerDeploymentName)
	}
	return nil
}

// InstallManagementPackages installs TKG management packages to the cluster
func InstallManagementPackages(pkgClient tkgpackageclient.TKGPackageClient, mpro ManagementPackageRepositoryOptions) error {
	// install management package repository
	err := installManagementPackageRepository(pkgClient, mpro)
	if err != nil {
		return err
	}

	// install tkg composite management package
	err = installTKGManagementPackage(pkgClient, mpro)
	if err != nil {
		return err
	}

	return nil
}

func installManagementPackageRepository(pkgClient tkgpackageclient.TKGPackageClient, mpro ManagementPackageRepositoryOptions) error {
	repositoryOptions := tkgpackagedatamodel.NewRepositoryOptions()
	repositoryOptions.RepositoryName = constants.TKGManagementPackageRepositoryName
	repositoryOptions.RepositoryURL = mpro.ManagementPackageRepoImage
	repositoryOptions.Namespace = constants.TkgNamespace
	repositoryOptions.CreateRepository = true
	repositoryOptions.Wait = true
	repositoryOptions.PollInterval = packagePollInterval
	repositoryOptions.PollTimeout = packagePollTimeout

	return pkgClient.UpdateRepositorySync(repositoryOptions, tkgpackagedatamodel.OperationTypeInstall)
}

func installTKGManagementPackage(pkgClient tkgpackageclient.TKGPackageClient, mpro ManagementPackageRepositoryOptions) error {
	packageOptions := tkgpackagedatamodel.NewPackageOptions()
	packageOptions.PackageName = constants.TKGManagementPackageName
	packageOptions.PkgInstallName = constants.TKGManagementPackageInstallName
	packageOptions.Namespace = constants.TkgNamespace
	packageOptions.Install = true
	packageOptions.Wait = true
	packageOptions.PollInterval = packagePollInterval
	packageOptions.PollTimeout = packagePollTimeout
	packageOptions.ValuesFile = mpro.TKGPackageValuesFile
	return pkgClient.InstallPackageSync(packageOptions, tkgpackagedatamodel.OperationTypeInstall)
}
