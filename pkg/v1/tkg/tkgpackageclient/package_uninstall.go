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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
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
			log.Warningf("\npackage '%s' is not installed in namespace '%s'", o.PkgInstallName, o.Namespace)
			err = nil
			return
		}
		err = errors.Wrap(err, fmt.Sprintf("\nfailed to find installed package '%s' in namespace '%s'", o.PkgInstallName, o.Namespace))
		return
	}

	progress.ProgressMsg <- fmt.Sprintf("Deleting package install '%s' from namespace '%s'", o.PkgInstallName, o.Namespace)
	if err = p.deletePackageInstall(o); err != nil {
		return
	}

	if err = p.waitForAppCRDeletion(o, progress.ProgressMsg); err != nil {
		return
	}

	if err = p.deletePkgPluginCreatedResources(pkgInstall, progress.ProgressMsg); err != nil {
		return
	}

	progress.Success <- true
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
func (p *pkgClient) deletePkgPluginCreatedResources(pkgInstall *kappipkg.PackageInstall, progress chan string) error { //nolint:gocyclo
	for k, v := range pkgInstall.GetAnnotations() {
		split := strings.Split(k, "/")
		if len(split) <= 1 {
			continue
		}
		resourceKind := strings.Split(split[1], tkgpackagedatamodel.TanzuPkgPluginPrefix)
		if len(resourceKind) <= 1 {
			continue
		}
		var obj runtime.Object
		objMeta := metav1.ObjectMeta{Name: v, Namespace: pkgInstall.Namespace}

		switch resourceKind[1] {
		case tkgpackagedatamodel.KindSecret:
			if progress != nil {
				progress <- fmt.Sprintf("Deleting secret '%s'", objMeta.Name)
			}
			obj = &corev1.Secret{ObjectMeta: objMeta, TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindSecret}}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete Secret resource")
			}
		case tkgpackagedatamodel.KindServiceAccount:
			if progress != nil {
				progress <- fmt.Sprintf("Deleting service account '%s'", objMeta.Name)
			}
			obj = &corev1.ServiceAccount{ObjectMeta: objMeta, TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindServiceAccount}}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete ServiceAccount resource")
			}
		case tkgpackagedatamodel.KindClusterRole:
			if progress != nil {
				progress <- fmt.Sprintf("Deleting admin role '%s'", objMeta.Name)
			}
			obj = &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{Name: v},
				TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindClusterRole},
			}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete ClusterRole resource")
			}
		case tkgpackagedatamodel.KindClusterRoleBinding:
			if progress != nil {
				progress <- fmt.Sprintf("Deleting role binding '%s'", objMeta.Name)
			}
			obj = &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: v},
				TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindClusterRoleBinding},
			}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete ClusterRoleBinding resource")
			}
		}
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
		return errors.Wrap(err, "failed to delete PackageInstall resource")
	}

	return nil
}

// waitForAppCRDeletion waits until the App CR get deleted successfully or a failure happen
func (p *pkgClient) waitForAppCRDeletion(o *tkgpackagedatamodel.PackageOptions, progress chan string) error {
	if err := wait.Poll(o.PollInterval, o.PollTimeout, func() (done bool, err error) {
		app, err := p.kappClient.GetAppCR(o.PkgInstallName, o.Namespace)
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
