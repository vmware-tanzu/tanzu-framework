// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packageclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

const (
	addonsManagerName = "addons-manager"
	addonFinalizer    = "tkg.tanzu.vmware.com/addon"
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

func generateAddonSecretName(clusterName, addonName string) string {
	return fmt.Sprintf("%s-tanzu-%s-addon", clusterName, addonName)
}

func pauseAddonSecretReconciliation(clusterClient clusterclient.Client, addonSecreteName, namespace string) error {
	log.Infof("Pausing reconciliation for %s/%s secret", namespace, addonSecreteName)
	secret := &corev1.Secret{}
	jsonPatch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  fmt.Sprintf("/metadata/annotations/tkg.tanzu.vmware.com~1addon-paused"),
			"value": "",
		},
	}
	payloadBytes, err := json.Marshal(jsonPatch)
	if err != nil {
		return errors.Wrap(err, "unable to generate json patch")
	}

	err = clusterClient.PatchResource(secret, addonSecreteName, namespace, string(payloadBytes), types.JSONPatchType, nil)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to pause %s secret reconciliation", addonSecreteName)
	}
	return nil
}

func pausePackageInstallReconciliation(clusterClient clusterclient.Client, pkgiName, namespace string) error {
	log.Infof("Pausing reconciliation for %s/%s packageinstall", namespace, pkgiName)
	pkgi := &kappipkg.PackageInstall{}
	jsonPatch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/paused",
			"value": true,
		},
	}
	payloadBytes, err := json.Marshal(jsonPatch)
	if err != nil {
		return errors.Wrap(err, "unable to generate json patch")
	}
	err = clusterClient.PatchResource(pkgi, pkgiName, namespace, string(payloadBytes), types.JSONPatchType, nil)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to pause %s packageinstall reconciliation", pkgiName)
	}
	return nil
}

// PausedAddonLifecycleManagement pauses/unpauses the lifecycle management of addon package with given name and namespace
func PauseAddonLifecycleManagement(clusterClient clusterclient.Client, clusterName, addonName, namespace string) error {
	log.Infof("Pausing lifecycle management for %s", addonName)
	pkgiName := fmt.Sprintf("tanzu-%s", addonsManagerName)
	addonSecretName := generateAddonSecretName(clusterName, addonsManagerName)

	err := pauseAddonSecretReconciliation(clusterClient, addonSecretName, namespace)
	if err != nil {
		return err
	}

	err = pausePackageInstallReconciliation(clusterClient, pkgiName, namespace)
	if err != nil {
		return err
	}

	return nil
}

// NoopDeletePackageInstall sets spec.noopdelete = true before deleting the package install
func NoopDeletePackageInstall(clusterClient clusterclient.Client, addonName, namespace string) error {
	log.Infof("Deleting %s/%s packageinstall with noopdelete", namespace, addonName)
	pkgiName := fmt.Sprintf("tanzu-%s", addonName)

	jsonPatch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/noopDelete",
			"value": true,
		},
	}
	payloadBytes, err := json.Marshal(jsonPatch)
	if err != nil {
		return errors.Wrap(err, "unable to generate json patch")
	}
	pkgi := &kappipkg.PackageInstall{}
	err = clusterClient.PatchResource(pkgi, pkgiName, namespace, string(payloadBytes), types.JSONPatchType, nil)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to patch %s packageinstall", pkgiName)
	}

	pkgi.Name = pkgiName
	pkgi.Namespace = namespace
	err = clusterClient.DeleteResource(pkgi)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete PackageInstall resource %s", pkgiName)
	}
	return nil
}

// DeleteAddonSecret deletes the secrete associated with the addon if present. Return no error if secret not found.
func DeleteAddonSecret(clusterClient clusterclient.Client, clusterName, addonName, namespace string) error {
	addonSecret := &corev1.Secret{}
	addonSecret.Name = generateAddonSecretName(clusterName, addonName)
	addonSecret.Namespace = constants.TkgNamespace
	log.Infof("Deleting %s/%s secret", addonSecret.Namespace, addonSecret.Name)
	err := clusterClient.GetResource(addonSecret, addonSecret.Name, addonSecret.Namespace, nil, nil)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if controllerutil.ContainsFinalizer(addonSecret, addonFinalizer) {
		controllerutil.RemoveFinalizer(addonSecret, addonFinalizer)
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return clusterClient.UpdateResource(addonSecret, addonSecret.Name, addonSecret.Namespace)
	})

	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	err = clusterClient.DeleteResource(addonSecret)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to delete addon addonSecret %s", addonSecret)
	}
	return nil
}

// AddonSecretExists returns true if given addon is present and was installed from core repository.
func AddonSecretExists(clusterClient clusterclient.Client, clusterName, addonName, namespace string) (bool, error) {
	addonSecret := &corev1.Secret{}
	addonSecret.Name = generateAddonSecretName(clusterName, addonName)
	addonSecret.Namespace = constants.TkgNamespace

	err := clusterClient.GetResource(addonSecret, addonSecret.Name, addonSecret.Namespace, nil, nil)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// InstallManagementComponents installs the management component to cluster
func InstallManagementComponents(mcip *ManagementComponentsInstallOptions) error {
	clusterClient, err := clusterclient.NewClient(mcip.ClusterOptions.Kubeconfig, mcip.ClusterOptions.Kubecontext, clusterclient.Options{})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}
	clusterName, err := clusterClient.GetCurrentClusterName(mcip.ClusterOptions.Kubecontext)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster name")
	}

	// If the addons-manager is moving from core repository to management repository, its lifecycle management
	// needs to be paused.
	previousAddonsManagerIsFromCoreRepo, err := AddonSecretExists(clusterClient, clusterName, addonsManagerName, constants.TkgNamespace)
	if err != nil {
		return err
	}
	if previousAddonsManagerIsFromCoreRepo {
		err = PauseAddonLifecycleManagement(clusterClient, clusterName, addonsManagerName, constants.TkgNamespace)
		if err != nil {
			return err
		}

		err = NoopDeletePackageInstall(clusterClient, addonsManagerName, constants.TkgNamespace)
		if err != nil {
			return err
		}
	}

	// create package client
	pkgClient, err := packageclient.NewPackageClientForContext(mcip.ClusterOptions.Kubeconfig, mcip.ClusterOptions.Kubecontext)
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

	if previousAddonsManagerIsFromCoreRepo {
		err = DeleteAddonSecret(clusterClient, clusterName, addonsManagerName, constants.TkgNamespace)
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
func InstallManagementPackages(pkgClient packageclient.PackageClient, mpro ManagementPackageRepositoryOptions) error {
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

func installManagementPackageRepository(pkgClient packageclient.PackageClient, mpro ManagementPackageRepositoryOptions) error {
	repositoryOptions := packagedatamodel.NewRepositoryOptions()
	repositoryOptions.RepositoryName = constants.TKGManagementPackageRepositoryName
	repositoryOptions.RepositoryURL = mpro.ManagementPackageRepoImage
	repositoryOptions.Namespace = constants.TkgNamespace
	repositoryOptions.CreateRepository = true
	repositoryOptions.Wait = true
	repositoryOptions.PollInterval = packagePollInterval
	repositoryOptions.PollTimeout = packagePollTimeout

	return pkgClient.UpdateRepositorySync(repositoryOptions, packagedatamodel.OperationTypeUpdate)
}

func installTKGManagementPackage(pkgClient packageclient.PackageClient, mpro ManagementPackageRepositoryOptions) error {
	packageOptions := packagedatamodel.NewPackageOptions()
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
	return pkgClient.InstallPackageSync(packageOptions, packagedatamodel.OperationTypeInstall)
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
		log.V(3).Warningf("waiting for package: %s", pn)
		group.Go(
			func() error {
				err := clusterClient.WaitForPackageInstall(pn, constants.TkgNamespace, packageInstallTimeout)
				if err != nil {
					log.V(3).Warningf("error while waiting for package '%s'", pn)
				} else {
					log.V(3).Infof("successfully reconciled package: %s", pn)
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
