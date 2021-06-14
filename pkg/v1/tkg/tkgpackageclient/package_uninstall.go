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

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

// UnInstallPackage uninstalls the InstalledPackage and its associated resources from the cluster
func (p *pkgClient) UnInstallPackage(o *tkgpackagedatamodel.PackageUninstallOptions) error {
	installedPkg, err := p.kappClient.GetInstalledPackage(o.InstalledPkgName, o.Namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof("package %s is not installed in namespace %s", o.InstalledPkgName, o.Namespace)
			return nil
		}
		return errors.Wrap(err, fmt.Sprintf("failed to find installed package %s in namespace %s", o.InstalledPkgName, o.Namespace))
	}

	log.Infof("Uninstalling package '%s' from namespace '%s'", o.InstalledPkgName, o.Namespace)

	if err := p.deleteInstalledPackage(o); err != nil {
		return errors.Wrap(err, "Error deleting installed package.")
	}

	if err := p.waitForAppCRDeletion(o); err != nil {
		return err
	}

	if err := p.deletePkgPluginCreatedResources(installedPkg); err != nil {
		return err
	}

	return nil
}

// deletePkgPluginCreatedResources deletes the associated resources which were installed upon installation of the InstalledPackage CR
func (p *pkgClient) deletePkgPluginCreatedResources(installedPkg *kappipkg.InstalledPackage) error {
	for k, v := range installedPkg.GetAnnotations() {
		split := strings.Split(k, "/")
		if len(split) <= 1 {
			continue
		}
		resourceKind := strings.Split(split[1], tkgpackagedatamodel.TanzuPkgPluginPrefix)
		if len(resourceKind) <= 1 {
			continue
		}
		var obj runtime.Object
		objMeta := metav1.ObjectMeta{Name: v, Namespace: installedPkg.Namespace}

		switch resourceKind[1] {
		case tkgpackagedatamodel.KindSecret:
			obj = &corev1.Secret{ObjectMeta: objMeta, TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindSecret}}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete Secret resource")
			}
		case tkgpackagedatamodel.KindServiceAccount:
			obj = &corev1.ServiceAccount{ObjectMeta: objMeta, TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindServiceAccount}}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete ServiceAccount resource")
			}
		case tkgpackagedatamodel.KindClusterRole:
			obj = &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{Name: v},
				TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindClusterRole},
			}
			if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
				return errors.Wrap(err, "failed to delete ClusterRole resource")
			}
		case tkgpackagedatamodel.KindClusterRoleBinding:
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

// deleteInstalledPackage deletes the InstalledPackage CR
func (p *pkgClient) deleteInstalledPackage(o *tkgpackagedatamodel.PackageUninstallOptions) error {
	obj := &kappipkg.InstalledPackage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.InstalledPkgName,
			Namespace: o.Namespace,
		},
		TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindInstalledPackage},
	}

	if err := p.kappClient.GetClient().Delete(context.Background(), obj); err != nil {
		return errors.Wrap(err, "failed to delete InstalledPackage resource")
	}

	return nil
}

// waitForAppCRDeletion waits until the App CR get deleted successfully or a failure happen
func (p *pkgClient) waitForAppCRDeletion(o *tkgpackagedatamodel.PackageUninstallOptions) error {
	if err := wait.Poll(o.PollInterval, o.PollTimeout, func() (done bool, err error) {
		app, err := p.kappClient.GetAppCR(o.InstalledPkgName, o.Namespace)
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		for _, cond := range app.Status.Conditions {
			if cond.Type == kappctrl.DeleteFailed {
				return false, fmt.Errorf("app reconciliation failed: %s", app.Status.UsefulErrorMessage)
			}
		}

		return false, nil
	}); err != nil {
		return err
	}

	return nil
}
