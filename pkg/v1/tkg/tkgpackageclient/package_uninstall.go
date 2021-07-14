// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

// UninstallPackage uninstalls the PackageInstall and its associated resources from the cluster
func (p *pkgClient) UninstallPackage(o *tkgpackagedatamodel.PackageOptions, progress *tkgpackagedatamodel.PackageProgress) {
	var (
		pkgInstall *kappipkg.PackageInstall
		err        error
	)

	defer func() {
		packageProgressCleanup(err, progress)
	}()

	progress.ProgressMsg <- fmt.Sprintf("Getting package install for '%s'", o.PkgInstallName)
	pkgInstall, err = p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			progress.ProgressMsg <- fmt.Sprintf("package '%s' is not installed in namespace '%s'", o.PkgInstallName, o.Namespace)
			err = nil
		} else {
			err = errors.Wrap(err, fmt.Sprintf("\nfailed to find installed package '%s' in namespace '%s'", o.PkgInstallName, o.Namespace))
			return
		}
	}

	if pkgInstall != nil {
		progress.ProgressMsg <- fmt.Sprintf("Deleting package install '%s' from namespace '%s'", o.PkgInstallName, o.Namespace)
	}
	if err = p.deletePackageInstall(o); err != nil {
		return
	}

	if err = p.waitForAppCRDeletion(o, progress.ProgressMsg); err != nil {
		return
	}

	if err = p.deletePkgPluginCreatedResources(o, pkgInstall, progress.ProgressMsg, progress.Success); err != nil {
		return
	}
}

func packageProgressCleanup(err error, progress *tkgpackagedatamodel.PackageProgress) {
	if err != nil {
		progress.Err <- err
	}
	close(progress.ProgressMsg)
	close(progress.Done)
	close(progress.Success)
}

// deletePkgPluginCreatedResources deletes the associated resources which were installed upon installation of the PackageInstall CR
func (p *pkgClient) deletePkgPluginCreatedResources(o *tkgpackagedatamodel.PackageOptions, pkgInstall *kappipkg.PackageInstall, progressMsg chan string, success chan bool) error { //nolint:gocyclo
	var objMeta metav1.ObjectMeta

	if pkgInstall == nil {
		resourceAnnotation := fmt.Sprintf(tkgpackagedatamodel.TanzuPkgPluginResource, o.PkgInstallName, o.Namespace)

		// delete the related ClusterRolebinding resource if already installed
		objMeta = metav1.ObjectMeta{
			Name: fmt.Sprintf(tkgpackagedatamodel.ClusterRoleBindingName, o.PkgInstallName, o.Namespace),
		}
		if err := p.deleteResourceHelper(&rbacv1.ClusterRoleBinding{}, crtclient.ObjectKey{Name: objMeta.Name}, resourceAnnotation); err != nil {
			return err
		}

		// delete the related ClusterRole resource if already installed
		objMeta = metav1.ObjectMeta{
			Name: fmt.Sprintf(tkgpackagedatamodel.ClusterRoleName, o.PkgInstallName, o.Namespace),
		}
		if err := p.deleteResourceHelper(&rbacv1.ClusterRole{}, crtclient.ObjectKey{Name: objMeta.Name}, resourceAnnotation); err != nil {
			return err
		}

		// delete the related ServiceAccount resource if already installed
		objMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf(tkgpackagedatamodel.ServiceAccountName, o.PkgInstallName, o.Namespace),
			Namespace: o.Namespace,
		}
		if err := p.deleteResourceHelper(&corev1.ServiceAccount{}, crtclient.ObjectKey{Name: objMeta.Name, Namespace: o.Namespace}, resourceAnnotation); err != nil {
			return err
		}

		// delete the related Secret resource if already installed
		objMeta = metav1.ObjectMeta{
			Name:      fmt.Sprintf(tkgpackagedatamodel.SecretName, o.PkgInstallName, o.Namespace),
			Namespace: o.Namespace,
		}
		if err := p.deleteResourceHelper(&corev1.Secret{}, crtclient.ObjectKey{Name: objMeta.Name, Namespace: o.Namespace}, resourceAnnotation); err != nil {
			return err
		}

		return nil
	}

	for k, v := range pkgInstall.GetAnnotations() {
		split := strings.Split(k, "/")
		if len(split) <= 1 {
			continue
		}
		resourceKind := strings.Split(split[1], tkgpackagedatamodel.TanzuPkgPluginPrefix+"-")
		if len(resourceKind) <= 1 {
			continue
		}

		objMeta = metav1.ObjectMeta{Name: v}

		switch resourceKind[1] {
		case tkgpackagedatamodel.KindSecret:
			if progressMsg != nil {
				progressMsg <- fmt.Sprintf("Deleting secret '%s'", objMeta.Name)
			}
			objMeta.Namespace = pkgInstall.Namespace
			if err := p.deleteSecret(&objMeta); err != nil {
				return errors.Wrap(err, "failed to delete Secret resource")
			}
		case tkgpackagedatamodel.KindServiceAccount:
			if progressMsg != nil {
				progressMsg <- fmt.Sprintf("Deleting service account '%s'", objMeta.Name)
			}
			objMeta.Namespace = pkgInstall.Namespace
			if err := p.deleteServiceAccount(&objMeta); err != nil {
				return errors.Wrap(err, "failed to delete ServiceAccount resource")
			}
		case tkgpackagedatamodel.KindClusterRole:
			if progressMsg != nil {
				progressMsg <- fmt.Sprintf("Deleting admin role '%s'", objMeta.Name)
			}
			if err := p.deleteClusterRole(&objMeta); err != nil {
				return errors.Wrap(err, "failed to delete ClusterRole resource")
			}
		case tkgpackagedatamodel.KindClusterRoleBinding:
			if progressMsg != nil {
				progressMsg <- fmt.Sprintf("Deleting role binding '%s'", objMeta.Name)
			}
			if err := p.deleteClusterRolebinding(&objMeta); err != nil {
				return errors.Wrap(err, "failed to delete ClusterRoleBinding resource")
			}
		}
	}

	if success != nil {
		success <- true
	}

	return nil
}

// deletePackageInstall deletes the PackageInstall CR
func (p *pkgClient) deletePackageInstall(o *tkgpackagedatamodel.PackageOptions) error {
	obj := &kappipkg.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.PkgInstallName,
			Namespace: o.Namespace,
		},
		TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindPackageInstall},
	}

	if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to delete PackageInstall resource")
		}
	}

	return nil
}

// waitForAppCRDeletion waits until the App CR get deleted successfully or a failure happen
func (p *pkgClient) waitForAppCRDeletion(o *tkgpackagedatamodel.PackageOptions, progress chan string) error {
	if err := wait.Poll(o.PollInterval, o.PollTimeout, func() (done bool, err error) {
		app, err := p.kappClient.GetAppCR(o.PkgInstallName, o.Namespace) // TODO: wait on package CR deletion instead
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		for _, cond := range app.Status.Conditions {
			if progress != nil {
				progress <- fmt.Sprintf("Package uninstall status: %s", cond.Type)
			}
			if cond.Type == kappctrl.DeleteFailed {
				return false, fmt.Errorf("app deletion failed: %s", app.Status.UsefulErrorMessage)
			}
		}

		return false, nil
	}); err != nil {
		return err
	}

	return nil
}

// deleteSecret deletes the Secret object
func (p *pkgClient) deleteSecret(objMeta *metav1.ObjectMeta) error {
	obj := &corev1.Secret{ObjectMeta: *objMeta, TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindSecret}}
	if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
		return err
	}

	return nil
}

// deleteServiceAccount deletes the ServiceAccount object
func (p *pkgClient) deleteServiceAccount(objMeta *metav1.ObjectMeta) error {
	obj := &corev1.ServiceAccount{ObjectMeta: *objMeta, TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindServiceAccount}}
	if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
		return err
	}

	return nil
}

// deleteClusterRole deletes the ClusterRole object
func (p *pkgClient) deleteClusterRole(objMeta *metav1.ObjectMeta) error {
	obj := &rbacv1.ClusterRole{
		ObjectMeta: *objMeta,
		TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindClusterRole},
	}
	if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
		return err
	}

	return nil
}

// deleteClusterRolebinding deletes the ClusterRolebinding object
func (p *pkgClient) deleteClusterRolebinding(objMeta *metav1.ObjectMeta) error {
	obj := &rbacv1.ClusterRoleBinding{
		ObjectMeta: *objMeta,
		TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindClusterRoleBinding},
	}
	if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
		return err
	}

	return nil
}

func (p *pkgClient) deleteResourceHelper(obj runtime.Object, objKey crtclient.ObjectKey, resourceAnnotation string) error {
	if err := p.kappClient.GetClient().Get(context.Background(), objKey, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		o, err := meta.Accessor(obj)
		if err != nil {
			return err
		}
		for k, v := range o.GetAnnotations() {
			split := strings.Split(k, "/")
			if len(split) > 1 && split[1] == tkgpackagedatamodel.TanzuPkgPluginPrefix {
				if v == resourceAnnotation {
					if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}
