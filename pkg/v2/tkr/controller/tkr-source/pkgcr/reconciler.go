// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pkgcr provides the TKR Package reconciler: it installs yet uninstalled TKR packages.
package pkgcr

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/kind/pkg/errors"

	kapppkgiv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	versionsv1 "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

type Reconciler struct {
	Log    logr.Logger
	Client client.Client

	Config Config
}

type Config struct {
	ServiceAccountName string
}

const (
	LabelTKRPackage = "run.tanzu.vmware.com/tkr-package"
)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kapppkgv1.Package{}, builder.WithPredicates(hasTKRPackageLabelPredicate, predicate.GenerationChangedPredicate{})).
		Owns(&kapppkgiv1.PackageInstall{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("tkr_source").
		Complete(r)
}

var hasTKRPackageLabelPredicate = predicate.NewPredicateFuncs(hasTKRPackageLabel)

func hasTKRPackageLabel(o client.Object) bool {
	return labels.Set(o.GetLabels()).Has(LabelTKRPackage)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	pkg := &kapppkgv1.Package{}

	if err := r.Client.Get(ctx, req.NamespacedName, pkg); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !pkg.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	pkgi := r.packageInstall(pkg)

	r.Log.Info("installing TKR package", "name", pkg.Spec.RefName, "version", pkg.Spec.Version)

	if err := r.Client.Create(ctx, pkgi); err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, errors.Wrap(err, "failed to create PackageInstall")
	}

	return ctrl.Result{}, nil
}

var pkgAPIVersion, pkgKind = kapppkgv1.SchemeGroupVersion.WithKind(reflect.TypeOf(kapppkgv1.Package{}).Name()).ToAPIVersionAndKind()

func (r *Reconciler) packageInstall(pkg *kapppkgv1.Package) *kapppkgiv1.PackageInstall {
	return &kapppkgiv1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tkr-" + version.Label(pkg.Spec.Version),
			Namespace: pkg.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: pkgAPIVersion,
				Kind:       pkgKind,
				Name:       pkg.Name,
				UID:        pkg.UID,
				Controller: pointer.BoolPtr(true),
			}},
		},
		Spec: kapppkgiv1.PackageInstallSpec{
			ServiceAccountName: r.Config.ServiceAccountName,
			PackageRef: &kapppkgiv1.PackageRef{
				RefName: pkg.Spec.RefName,
				VersionSelection: &versionsv1.VersionSelectionSemver{
					Constraints: pkg.Spec.Version,
				},
			},
		},
	}
}
