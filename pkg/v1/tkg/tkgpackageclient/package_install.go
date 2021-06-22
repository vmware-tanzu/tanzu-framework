// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

// InstallPackage installs the PackageInstall and its associated resources in the cluster
func (p *pkgClient) InstallPackage(o *tkgpackagedatamodel.PackageOptions) error {
	if _, _, err := p.GetPackage(o.PackageName, o.Version, o.Namespace); err != nil {
		return err
	}

	if o.CreateNamespace {
		if err := p.createNamespace(o.Namespace); err != nil {
			return err
		}
	}

	if o.ServiceAccountName == "" {
		if err := p.createServiceAccount(o); err != nil {
			return err
		}
		if err := p.createClusterAdminRole(o); err != nil {
			return err
		}
		if err := p.createClusterRoleBinding(o); err != nil {
			return err
		}
	} else {
		objKey := crtclient.ObjectKey{Name: o.ServiceAccountName, Namespace: o.Namespace}
		svcAccount := &corev1.ServiceAccount{}
		if err := p.kappClient.GetClient().Get(context.Background(), objKey, svcAccount); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to find service account '%s' in namespace '%s'", o.ServiceAccountName, o.Namespace))
		}
		if _, ok := svcAccount.GetAnnotations()[tkgpackagedatamodel.TanzuPkgPluginAnnotation]; ok {
			return errors.New(fmt.Sprintf("provided service account '%s' is already used by another package in namespace '%s'", o.ServiceAccountName, o.Namespace))
		}
	}

	if o.ValuesFile != "" {
		if err := p.createDataValuesSecret(o); err != nil {
			return err
		}
	}

	if err := p.createPackageInstall(o); err != nil {
		return err
	}

	log.Infof("Installing package '%s' in namespace '%s'", o.PkgInstallName, o.Namespace)
	if o.Wait {
		if err := p.waitForPackageInstallation(o); err != nil {
			return err
		}
	}

	return nil
}

// createClusterAdminRole creates a ClusterRole resource
func (p *pkgClient) createClusterAdminRole(o *tkgpackagedatamodel.PackageOptions) error {
	o.ClusterRoleName = fmt.Sprintf(tkgpackagedatamodel.ClusterRoleName, o.PkgInstallName, o.Namespace)
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: o.ClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: []string{"*"}, Resources: []string{"*"}},
		},
	}

	if err := p.kappClient.GetClient().Create(context.Background(), clusterRole); err != nil {
		return errors.Wrap(err, "failed to create ClusterRole resource")
	}

	return nil
}

// createClusterRoleBinding creates a ClusterRoleBinding resource
func (p *pkgClient) createClusterRoleBinding(o *tkgpackagedatamodel.PackageOptions) error {
	clusterRoleBindingName := fmt.Sprintf(tkgpackagedatamodel.ClusterRoleBindingName, o.PkgInstallName, o.Namespace)
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: clusterRoleBindingName},
		Subjects:   []rbacv1.Subject{{Kind: tkgpackagedatamodel.KindServiceAccount, Name: o.ServiceAccountName, Namespace: o.Namespace}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     tkgpackagedatamodel.KindClusterRole,
			Name:     o.ClusterRoleName,
		},
	}

	if err := p.kappClient.GetClient().Create(context.Background(), clusterRoleBinding); err != nil {
		return errors.Wrap(err, "failed to create ClusterRoleBinding resource")
	}

	return nil
}

// createDataValuesSecret create a secret object containing the user-provided configuration.
func (p *pkgClient) createDataValuesSecret(o *tkgpackagedatamodel.PackageOptions) error {
	var err error
	dataValues := make(map[string][]byte)

	if dataValues[filepath.Base(o.ValuesFile)], err = ioutil.ReadFile(o.ValuesFile); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to read from data values file '%s'", o.ValuesFile))
	}
	secretName := fmt.Sprintf(tkgpackagedatamodel.SecretName, o.PkgInstallName, o.Namespace)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: o.Namespace}, Data: dataValues,
	}

	if err := p.kappClient.GetClient().Create(context.Background(), secret); err != nil {
		return errors.Wrap(err, "failed to create Secret resource")
	}

	o.CreateSecret = true

	return nil
}

// createNamespace creates a namespace resource if it doesn't already exist
func (p *pkgClient) createNamespace(namespace string) error {
	err := p.kappClient.GetClient().Get(
		context.Background(),
		crtclient.ObjectKey{Name: namespace},
		&corev1.Namespace{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		ns := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindNamespace},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}
		if err := p.kappClient.GetClient().Create(context.Background(), ns); err != nil {
			return errors.Wrap(err, "failed to create namespace")
		}
	}

	return nil
}

// createPackageInstall creates the PackageInstall CR
func (p *pkgClient) createPackageInstall(o *tkgpackagedatamodel.PackageOptions) error {
	// construct the PackageInstall CR
	packageInstall := &kappipkg.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{Name: o.PkgInstallName, Namespace: o.Namespace},
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
	if o.CreateSecret {
		packageInstall.Spec.Values = []kappipkg.PackageInstallValues{
			{
				SecretRef: &kappipkg.PackageInstallValuesSecretRef{
					Name: fmt.Sprintf(tkgpackagedatamodel.SecretName, o.PkgInstallName, o.Namespace),
				},
			},
		}
	}

	if err := p.kappClient.CreatePackageInstall(packageInstall, o.CreateServiceAccount, o.CreateSecret); err != nil {
		return errors.Wrap(err, "failed to create PackageInstall resource")
	}

	return nil
}

// createServiceAccount creates a ServiceAccount resource
func (p *pkgClient) createServiceAccount(o *tkgpackagedatamodel.PackageOptions) error {
	o.ServiceAccountName = fmt.Sprintf(tkgpackagedatamodel.ServiceAccountName, o.PkgInstallName, o.Namespace)
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.ServiceAccountName,
			Namespace: o.Namespace,
			Annotations: map[string]string{
				tkgpackagedatamodel.TanzuPkgPluginAnnotation: o.ServiceAccountName},
		},
	}

	if err := p.kappClient.GetClient().Create(context.Background(), serviceAccount); err != nil {
		return errors.Wrap(err, "failed to create ServiceAccount resource")
	}

	o.CreateServiceAccount = true

	return nil
}

// waitForPackageInstallation waits until the package get installed successfully or a failure happen
func (p *pkgClient) waitForPackageInstallation(o *tkgpackagedatamodel.PackageOptions) error {
	if err := wait.Poll(o.PollInterval, o.PollTimeout, func() (done bool, err error) {
		pkg, err := p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace)
		if err != nil {
			return false, err
		}
		for _, cond := range pkg.Status.Conditions {
			switch cond.Type {
			case kappctrl.ReconcileSucceeded:
				return true, nil
			case kappctrl.ReconcileFailed:
				return false, fmt.Errorf("app reconciliation failed: %s", pkg.Status.UsefulErrorMessage)
			}
		}
		return false, nil
	}); err != nil {
		return err
	}

	return nil
}
