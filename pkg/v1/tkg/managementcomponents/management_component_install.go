// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
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
	PackageVersion             string
	PackageInstallTimeout      time.Duration
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
}

// InstallManagementComponents installs the management component to cluster
func InstallManagementComponents(mcip *ManagementComponentsInstallOptions) error {
	clusterClient, err := clusterclient.NewClient(mcip.ClusterOptions.Kubeconfig, mcip.ClusterOptions.Kubecontext, clusterclient.Options{})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}

	// create package client
	pkgClient, err := tkgpackageclient.NewTKGPackageClientForContext(mcip.ClusterOptions.Kubeconfig, mcip.ClusterOptions.Kubecontext)
	if err != nil {
		return err
	}
	if err = InstallManagementPackages(pkgClient, mcip.ManagementPackageRepositoryOptions); err != nil {
		// instead of throwing error here, wait for some additional time for packages to get reconciled successfully
		// error will be thrown at the next step if packages are not reconciled after timeout value
		log.Warning(err.Error())
	}

	err = WaitForManagementPackages(clusterClient, mcip.ManagementPackageRepositoryOptions.PackageInstallTimeout)
	if err != nil {
		return errors.Wrap(err, "timed out waiting for management packages to get reconciled successfully")
	}

	// Hack: This is temporary implementation to deploy missing components after installing management packages
	// This is currently used to deploy TKR related resources. This can be removed once tkr-source-controller is in place
	// and can deploy the necessary tkr components
	resouceFile := os.Getenv("_ADDITIONAL_MANAGEMENT_COMPONENT_CONFIGURATION_FILE")
	if resouceFile != "" {
		log.Infof("Appling additional management component configuration from %q", resouceFile)
		err := clusterClient.ApplyFile(resouceFile)
		if err != nil {
			return err
		}
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
		return errors.Wrap(err, "unable to install management package repository")
	}

	// install tkg composite management package
	err = installTKGManagementPackage(pkgClient, mpro)
	if err != nil {
		return errors.Wrap(err, "failure while installing TKG management package")
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

	return pkgClient.UpdateRepositorySync(repositoryOptions, tkgpackagedatamodel.OperationTypeUpdate)
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
	packageOptions.Version = mpro.PackageVersion
	packageOptions.Labels = map[string]string{constants.PackageTypeLabel: constants.PackageTypeManagement}
	return pkgClient.InstallPackageSync(packageOptions, tkgpackagedatamodel.OperationTypeInstall)
}

func WaitForManagementPackages(clusterClient clusterclient.Client, packageInstallTimeout time.Duration) error {
	var packageInstalls kappipkg.PackageInstallList
	labelMatch, _ := labels.NewRequirement(constants.PackageTypeLabel, selection.Equals, []string{constants.PackageTypeManagement})
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*labelMatch)

	err := clusterClient.ListResources(&packageInstalls, &crtclient.ListOptions{
		Namespace:     constants.TkgNamespace,
		LabelSelector: labelSelector,
	})
	if err != nil {
		return errors.Wrap(err, "unable to list PackageInstalls")
	}

	// Filter management packages all the available packages from tkg-system namespace
	// Management package will have a specific label which is being used to filter the packages
	packageInstallNames := []string{constants.TKGManagementPackageInstallName}
	for i := range packageInstalls.Items {
		if packageInstalls.Items[i].Name != constants.TKGManagementPackageInstallName {
			packageInstallNames = append(packageInstallNames, packageInstalls.Items[i].Name)
		}
	}

	// Start waiting for all packages in parallel using group.Wait
	// Note: As PackageInstall resources are created in the cluster itself
	// we are using currentClusterClient which will point to correct cluster
	group, _ := errgroup.WithContext(context.Background())

	for _, packageName := range packageInstallNames {
		pn := packageName
		log.V(3).Warningf("Waiting for package: %s", pn)
		group.Go(
			func() error {
				err := clusterClient.WaitForPackageInstall(pn, constants.TkgNamespace, packageInstallTimeout)
				if err != nil {
					log.V(3).Warningf("Error while waiting for package '%s'", pn)
				} else {
					log.V(3).Infof("Successfully reconciled package: %s", pn)
				}
				return err
			})
	}

	err = group.Wait()
	if err != nil {
		return errors.Wrap(err, "error while waiting for management packages to be installed")
	}
	return nil
}
