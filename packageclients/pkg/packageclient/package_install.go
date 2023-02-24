// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

const (
	msgRunPackageInstalledUpdate = "\n\nPlease consider using 'tanzu package installed update' to update the installed package with correct settings\n"
)

// InstallPackage installs the PackageInstall and its associated resources in the cluster
func (p *pkgClient) InstallPackage(o *packagedatamodel.PackageOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType) {
	p.installPackage(o, progress, operationType)
}

// InstallPackageSync installs the PackageInstall and its associated resources in the cluster and returns an error if any
func (p *pkgClient) InstallPackageSync(o *packagedatamodel.PackageOptions, operationType packagedatamodel.OperationType) error {
	pp := newPackageProgress()

	go p.installPackage(o, pp, operationType)

	initialMsg := fmt.Sprintf("Installing package '%s'", o.PackageName)
	if err := DisplayProgress(initialMsg, pp); err != nil {
		if err.Error() == packagedatamodel.ErrPackageAlreadyExists {
			log.Infof("Updated installed package '%s'", o.PkgInstallName)
			return nil
		}
		return err
	}

	log.Infof("\n %s", fmt.Sprintf("Added installed package '%s'",
		o.PkgInstallName))
	return nil
}

func (p *pkgClient) installPackage(o *packagedatamodel.PackageOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType) {
	var (
		pkgInstall                      *kappipkg.PackageInstall
		pkgPluginResourceCreationStatus *packagedatamodel.PkgPluginResourceCreationStatus
		err                             error
	)

	defer func() {
		if err != nil {
			progress.Err <- err
		}
		if operationType == packagedatamodel.OperationTypeInstall {
			close(progress.ProgressMsg)
			close(progress.Done)
		}
	}()

	if pkgInstall, err = p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace); err != nil {
		if !k8serror.IsNotFound(err) {
			return
		}
		err = nil
	}

	if pkgInstall != nil {
		progress.ProgressMsg <- fmt.Sprintf("Updating package '%s'", o.PkgInstallName)
		p.UpdatePackage(o, progress, packagedatamodel.OperationTypeInstall)
		err = &packagedatamodel.PackagePluginNonCriticalError{Reason: packagedatamodel.ErrPackageAlreadyExists}
		return
	}

	if err = p.validateValuesFile(o); err != nil {
		return
	}

	if o.CreateNamespace {
		progress.ProgressMsg <- fmt.Sprintf("Creating namespace '%s'", o.Namespace)
		if err = p.createNamespace(o.Namespace); err != nil {
			return
		}
	} else if err = p.kappClient.GetClient().Get(context.Background(), crtclient.ObjectKey{Name: o.Namespace}, &corev1.Namespace{}); err != nil {
		return
	}

	progress.ProgressMsg <- fmt.Sprintf("Getting package metadata for '%s'", o.PackageName)
	if _, _, err = p.GetPackage(o); err != nil {
		return
	}

	if pkgPluginResourceCreationStatus, err = p.createRelatedResources(o, progress.ProgressMsg); err != nil {
		return
	}

	progress.ProgressMsg <- "Creating package resource"
	if err = p.createPackageInstall(o, pkgPluginResourceCreationStatus); err != nil {
		return
	}

	if o.Wait {
		if err = p.waitForResourceInstallation(o.PkgInstallName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, packagedatamodel.ResourceTypePackageInstall); err != nil {
			log.Warning(msgRunPackageInstalledUpdate)
			return
		}
	}
}

func (p *pkgClient) createRelatedResources(o *packagedatamodel.PackageOptions, progress chan string) (*packagedatamodel.PkgPluginResourceCreationStatus, error) {
	var (
		pkgPluginResourceCreationStatus packagedatamodel.PkgPluginResourceCreationStatus
		err                             error
	)

	if o.ServiceAccountName == "" {
		o.ServiceAccountName = fmt.Sprintf(packagedatamodel.ServiceAccountName, o.PkgInstallName, o.Namespace)
		progress <- fmt.Sprintf("Creating service account '%s'", o.ServiceAccountName)
		if pkgPluginResourceCreationStatus.IsServiceAccountCreated, err = p.createOrUpdateServiceAccount(o); err != nil {
			return &pkgPluginResourceCreationStatus, err
		}

		o.ClusterRoleName = fmt.Sprintf(packagedatamodel.ClusterRoleName, o.PkgInstallName, o.Namespace)
		progress <- fmt.Sprintf("Creating cluster admin role '%s'", o.ClusterRoleName)
		if err := p.createOrUpdateClusterAdminRole(o); err != nil {
			return &pkgPluginResourceCreationStatus, err
		}

		o.ClusterRoleBindingName = fmt.Sprintf(packagedatamodel.ClusterRoleBindingName, o.PkgInstallName, o.Namespace)
		progress <- fmt.Sprintf("Creating cluster role binding '%s'", o.ClusterRoleBindingName)
		if err := p.createOrUpdateClusterRoleBinding(o); err != nil {
			return &pkgPluginResourceCreationStatus, err
		}
	} else {
		objKey := crtclient.ObjectKey{Name: o.ServiceAccountName, Namespace: o.Namespace}
		svcAccount := &corev1.ServiceAccount{}
		if err = p.kappClient.GetClient().Get(context.Background(), objKey, svcAccount); err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to find service account '%s' in namespace '%s'", o.ServiceAccountName, o.Namespace))
			return &pkgPluginResourceCreationStatus, err
		}
		if svcAccountAnnotation, ok := svcAccount.GetAnnotations()[packagedatamodel.TanzuPkgPluginAnnotation]; ok {
			if svcAccountAnnotation != fmt.Sprintf(packagedatamodel.TanzuPkgPluginResource, o.PkgInstallName, o.Namespace) {
				err = fmt.Errorf("provided service account '%s' is already used by another package in namespace '%s'", o.ServiceAccountName, o.Namespace)
				return &pkgPluginResourceCreationStatus, err
			}
		}
	}

	if o.ValuesFile != "" {
		o.SecretName = fmt.Sprintf(packagedatamodel.SecretName, o.PkgInstallName, o.Namespace)
		progress <- fmt.Sprintf("Creating secret '%s'", o.SecretName)
		if pkgPluginResourceCreationStatus.IsSecretCreated, err = p.createOrUpdateDataValuesSecret(o); err != nil {
			return &pkgPluginResourceCreationStatus, err
		}
	}

	return &pkgPluginResourceCreationStatus, nil
}

// createOrUpdateClusterAdminRole creates or updates a ClusterRole resource
func (p *pkgClient) createOrUpdateClusterAdminRole(o *packagedatamodel.PackageOptions) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.ClusterRoleName,
			Annotations: map[string]string{packagedatamodel.TanzuPkgPluginAnnotation: fmt.Sprintf(packagedatamodel.TanzuPkgPluginResource, o.PkgInstallName, o.Namespace)},
		},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: []string{"*"}, Resources: []string{"*"}},
		},
	}

	if err := p.kappClient.GetClient().Create(context.Background(), clusterRole); err != nil {
		if k8serror.IsAlreadyExists(err) {
			if err := p.kappClient.GetClient().Update(context.Background(), clusterRole); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// createOrUpdateClusterRoleBinding creates or updates a ClusterRoleBinding resource
func (p *pkgClient) createOrUpdateClusterRoleBinding(o *packagedatamodel.PackageOptions) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.ClusterRoleBindingName,
			Annotations: map[string]string{packagedatamodel.TanzuPkgPluginAnnotation: fmt.Sprintf(packagedatamodel.TanzuPkgPluginResource, o.PkgInstallName, o.Namespace)},
		},
		Subjects: []rbacv1.Subject{{Kind: packagedatamodel.KindServiceAccount, Name: o.ServiceAccountName, Namespace: o.Namespace}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     packagedatamodel.KindClusterRole,
			Name:     o.ClusterRoleName,
		},
	}

	if err := p.kappClient.GetClient().Create(context.Background(), clusterRoleBinding); err != nil {
		if k8serror.IsAlreadyExists(err) {
			if err := p.kappClient.GetClient().Update(context.Background(), clusterRoleBinding); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// createOrUpdateDataValuesSecret create or updates a secret object containing the user-provided configuration.
func (p *pkgClient) createOrUpdateDataValuesSecret(o *packagedatamodel.PackageOptions) (bool, error) {
	var err error

	dataValues := make(map[string][]byte)

	if dataValues[filepath.Base(o.ValuesFile)], err = os.ReadFile(o.ValuesFile); err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed to read from data values file '%s'", o.ValuesFile))
	}
	Secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.SecretName,
			Namespace:   o.Namespace,
			Annotations: map[string]string{packagedatamodel.TanzuPkgPluginAnnotation: fmt.Sprintf(packagedatamodel.TanzuPkgPluginResource, o.PkgInstallName, o.Namespace)},
		},
		Data: dataValues,
	}

	if err := p.kappClient.GetClient().Create(context.Background(), Secret); err != nil {
		if k8serror.IsAlreadyExists(err) {
			if err := p.kappClient.GetClient().Update(context.Background(), Secret); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	return true, nil
}

// createNamespace creates a namespace resource if it doesn't already exist
func (p *pkgClient) createNamespace(namespace string) error {
	err := p.kappClient.GetClient().Get(
		context.Background(),
		crtclient.ObjectKey{Name: namespace},
		&corev1.Namespace{})
	if err != nil {
		if !k8serror.IsNotFound(err) {
			return err
		}
		ns := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: packagedatamodel.KindNamespace},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}
		if err := p.kappClient.GetClient().Create(context.Background(), ns); err != nil {
			return errors.Wrap(err, "failed to create namespace")
		}
	}

	return nil
}

// createPackageInstall creates the PackageInstall CR
func (p *pkgClient) createPackageInstall(o *packagedatamodel.PackageOptions, pkgPluginResourceCreationStatus *packagedatamodel.PkgPluginResourceCreationStatus) error {
	// construct the PackageInstall CR
	packageInstall := &kappipkg.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{Name: o.PkgInstallName,
			Namespace: o.Namespace,
			Labels:    o.Labels,
		},
		Spec: kappipkg.PackageInstallSpec{
			ServiceAccountName: o.ServiceAccountName,
			PackageRef: &kappipkg.PackageRef{
				RefName: o.PackageName,
				VersionSelection: &versions.VersionSelectionSemver{
					Constraints: o.Version,
					Prereleases: &versions.VersionSelectionSemverPrereleases{},
				},
			},
		},
	}

	// if configuration data file was provided, reference the secret name in the PackageInstall
	if pkgPluginResourceCreationStatus.IsSecretCreated {
		packageInstall.Spec.Values = []kappipkg.PackageInstallValues{
			{
				SecretRef: &kappipkg.PackageInstallValuesSecretRef{
					Name: fmt.Sprintf(packagedatamodel.SecretName, o.PkgInstallName, o.Namespace),
				},
			},
		}
	}

	if err := p.kappClient.CreatePackageInstall(packageInstall, pkgPluginResourceCreationStatus); err != nil {
		return errors.Wrap(err, "failed to create PackageInstall resource")
	}

	return nil
}

// createOrUpdateServiceAccount creates or updates a ServiceAccount resource
func (p *pkgClient) createOrUpdateServiceAccount(o *packagedatamodel.PackageOptions) (bool, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.ServiceAccountName,
			Namespace:   o.Namespace,
			Annotations: map[string]string{packagedatamodel.TanzuPkgPluginAnnotation: fmt.Sprintf(packagedatamodel.TanzuPkgPluginResource, o.PkgInstallName, o.Namespace)}},
	}

	if err := p.kappClient.GetClient().Create(context.Background(), serviceAccount); err != nil {
		if k8serror.IsAlreadyExists(err) {
			if err := p.kappClient.GetClient().Update(context.Background(), serviceAccount); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	return true, nil
}

func (p *pkgClient) validateValuesFile(o *packagedatamodel.PackageOptions) error {
	if o.ValuesFile == "" {
		return nil
	}

	if _, err := os.ReadFile(o.ValuesFile); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to read from data values file '%s'", o.ValuesFile))
		return err
	}

	return nil
}

// waitForResourceInstallation waits until the package get installed successfully or a failure happen
func (p *pkgClient) waitForResourceInstallation(name, namespace string, pollInterval, pollTimeout time.Duration, progress chan string, rscType packagedatamodel.ResourceType) error { //nolint:gocyclo
	var (
		status             kappctrl.GenericStatus
		reconcileSucceeded bool
	)
	progress <- fmt.Sprintf("Waiting for '%s' reconciliation for '%s'", rscType.String(), name)
	if err := wait.Poll(pollInterval, pollTimeout, func() (done bool, err error) {
		switch rscType {
		case packagedatamodel.ResourceTypePackageRepository:
			resource, err := p.kappClient.GetPackageRepository(name, namespace)
			if err != nil {
				return false, err
			}
			if resource.Generation != resource.Status.ObservedGeneration {
				// Should wait for generation to be observed before checking the reconciliation status so that we know we are checking the new spec
				return false, nil
			}
			status = resource.Status.GenericStatus
		case packagedatamodel.ResourceTypePackageInstall:
			resource, err := p.kappClient.GetPackageInstall(name, namespace)
			if err != nil {
				return false, err
			}
			if resource.Generation != resource.Status.ObservedGeneration {
				// Should wait for generation to be observed before checking the reconciliation status so that we know we are checking the new spec
				return false, nil
			}
			status = resource.Status.GenericStatus
		}

		for _, cond := range status.Conditions {
			if progress != nil {
				progress <- fmt.Sprintf("'%s' resource install status: %s", rscType.String(), cond.Type)
			}
			switch {
			case cond.Type == kappctrl.ReconcileSucceeded && cond.Status == corev1.ConditionTrue:
				if progress != nil {
					progress <- fmt.Sprintf("'%s' resource successfully reconciled", rscType.String())
				}
				reconcileSucceeded = true
				return true, nil
			case cond.Type == kappctrl.ReconcileFailed && cond.Status == corev1.ConditionTrue:
				return false, fmt.Errorf("resource reconciliation failed: %s. %s", status.UsefulErrorMessage, status.FriendlyDescription)
			}
		}
		return false, nil
	}); err != nil {
		return err
	}

	if !reconcileSucceeded {
		return fmt.Errorf("'%s' resource reconciliation timeout", rscType.String())
	}

	return nil
}
