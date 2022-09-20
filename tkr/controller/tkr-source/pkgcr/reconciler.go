// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pkgcr provides the TKR Package reconciler: it installs yet uninstalled TKR packages.
package pkgcr

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/registry"
	"github.com/vmware-tanzu/tanzu-framework/util/patchset"
)

type Reconciler struct {
	Log    logr.Logger
	Client client.Client

	Config Config

	Registry registry.Registry
}

type Config struct {
	ServiceAccountName string
}

type InstallData struct {
	ObservedGeneration int64
	Success            bool
}

const (
	LabelTKRPackage  = "run.tanzu.vmware.com/tkr-package"
	FieldInstallData = "installData"
)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kapppkgv1.Package{}, builder.WithPredicates(hasTKRPackageLabelPredicate, predicate.GenerationChangedPredicate{})).
		Owns(&corev1.ConfigMap{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
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
		return ctrl.Result{}, errors.Wrapf(err, "getting Package '%s'", req)
	}
	if !pkg.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	done, err := r.install(ctx, pkg)
	return ctrl.Result{Requeue: !done}, errors.Wrapf(err, "failed to install TKR package '%s'", pkg.Name)
}

var pkgAPIVersion, pkgKind = kapppkgv1.SchemeGroupVersion.WithKind(reflect.TypeOf(kapppkgv1.Package{}).Name()).ToAPIVersionAndKind()
var cmAPIVersion, cmKind = corev1.SchemeGroupVersion.WithKind(reflect.TypeOf(corev1.ConfigMap{}).Name()).ToAPIVersionAndKind()

func (r *Reconciler) install(ctx context.Context, pkg *kapppkgv1.Package) (done bool, err error) {
	r.Log.Info("Processing TKR package", "name", pkg.Name)

	cm, err := r.createCM(ctx, pkg)
	if cm == nil {
		return false, err // err == nil is possible
	}

	installData := parseInstallData(cm.Data[FieldInstallData])
	if installData.Success {
		r.Log.Info("Already installed TKR package", "name", pkg.Name)
		return true, nil
	}

	done, err = r.doInstall(ctx, pkg, cm)
	if done {
		r.Log.Info("Installed TKR package", "name", pkg.Name)
	}
	return done, err
}

func (r *Reconciler) createCM(ctx context.Context, pkg *kapppkgv1.Package) (*corev1.ConfigMap, error) {
	cm0 := cmForPkg(pkg)
	if err := r.Client.Create(ctx, cm0); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, errors.Wrapf(err, "creating ConfigMap '%s'", objKey(cm0))
		}

		cm := &corev1.ConfigMap{}
		if err := r.Client.Get(ctx, objKey(cm0), cm); err != nil {
			return nil, errors.Wrapf(err, "getting ConfigMap '%s'", objKey(cm0))
		}

		installData := parseInstallData(cm.Data[FieldInstallData])
		if installData == nil || installData.ObservedGeneration != pkg.GetGeneration() {
			err := r.Client.Delete(ctx, cm)
			err = kerrors.FilterOut(err, apierrors.IsNotFound)
			return nil, errors.Wrapf(err, "deleting ConfigMap '%s'", objKey(cm)) // err == nil is possible
		}

		return cm, nil
	}

	return cm0, nil
}

func objKey(o client.Object) client.ObjectKey {
	return client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}
}

func cmForPkg(pkg *kapppkgv1.Package) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pkg.Namespace,
			Name:      fmt.Sprintf("tkr-%s", version.Label(pkg.Spec.Version)),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: pkgAPIVersion,
				Kind:       pkgKind,
				Name:       pkg.Name,
				UID:        pkg.UID,
				Controller: pointer.BoolPtr(true),
			}},
			Labels: map[string]string{
				LabelTKRPackage: pkg.Labels[LabelTKRPackage],
			},
		},
		Data: map[string]string{
			FieldInstallData: marshalInstallData(&InstallData{
				ObservedGeneration: pkg.Generation,
			}),
		},
	}
}

func parseInstallData(s string) *InstallData {
	installData := &InstallData{}
	_ = yaml.Unmarshal([]byte(s), installData)
	return installData
}

func marshalInstallData(installData *InstallData) string {
	result, _ := yaml.Marshal(installData)
	return string(result)
}

func (r *Reconciler) doInstall(ctx context.Context, pkg *kapppkgv1.Package, cm *corev1.ConfigMap) (done bool, retErr error) {
	ps := patchset.New(r.Client)
	defer func() {
		if err := ps.Apply(ctx); err != nil {
			err = kerrors.FilterOut(err, apierrors.IsConflict)
			done, retErr = false, errors.Wrap(err, "applying patchset") // err == nil is possible
		}
	}()

	installData := parseInstallData(cm.Data[FieldInstallData])
	defer func() {
		cm.Data[FieldInstallData] = marshalInstallData(installData)
	}()

	packageContent, err := r.fetchPackageContent(pkg)
	if err != nil {
		return false, err
	}

	for path, bytes := range packageContent {
		if strings.HasPrefix(path, ".") {
			continue
		}
		u, err := parseObject(cm, bytes)
		if err != nil {
			r.Log.Error(err, "Failed to parse an object from package", "pkg", pkg.Name, "path", path)
			continue
		}
		if err = r.create(ctx, u); err != nil {
			r.Log.Error(err, "Failed to create an object from package", "pkg", pkg.Name, "path", path)
			return false, err
		}
	}

	ps.Add(cm)
	installData.Success = true

	return true, nil
}

func parseObject(cm *corev1.ConfigMap, bytes []byte) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(bytes, u); err != nil {
		return nil, err
	}
	u.SetNamespace(cm.Namespace)
	addOwnerRefs(u, []metav1.OwnerReference{{
		APIVersion: cmAPIVersion,
		Kind:       cmKind,
		Name:       cm.Name,
		UID:        cm.UID,
	}})
	return u, nil
}

func (r *Reconciler) fetchPackageContent(pkg *kapppkgv1.Package) (map[string][]byte, error) {
	if pkg.Spec.Template.Spec == nil {
		return nil, nil
	}
	for _, fetch := range pkg.Spec.Template.Spec.Fetch {
		if fetch.ImgpkgBundle == nil {
			return nil, nil
		}
		files, err := r.Registry.GetFiles(fetch.ImgpkgBundle.Image)
		if err != nil {
			return nil, err
		}
		return files, nil // nolint:staticcheck // loop is unconditionally terminated: there's only one Fetch
	}
	return nil, nil
}

func (r *Reconciler) create(ctx context.Context, u *unstructured.Unstructured) error {
	r.Log.Info("Creating object", "GVK", u.GetObjectKind().GroupVersionKind(), "objectKey", objKey(u))
	for {
		if err := r.Client.Create(ctx, u); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return errors.Wrapf(err, "creating object '%s', named '%s'", u.GetObjectKind().GroupVersionKind(), objKey(u))
			}
			if err := r.patchExisting(ctx, u); err != nil {
				if err := kerrors.FilterOut(err, apierrors.IsConflict, apierrors.IsNotFound); err != nil {
					return err // not IsConflict, not IsNotFound
				}
				r.Log.Info("Re-trying to create object", "GVK", u.GetObjectKind().GroupVersionKind(), "objectKey", objKey(u))
				continue // try to create again
			}
		}
		return nil
	}
}

func (r *Reconciler) patchExisting(ctx context.Context, u *unstructured.Unstructured) error {
	existing := &unstructured.Unstructured{}
	existing.SetAPIVersion(u.GetAPIVersion())
	existing.SetKind(u.GetKind())
	if err := r.Client.Get(ctx, objKey(u), existing); err != nil {
		return errors.Wrapf(err, "getting object '%s', named '%s'", u.GetObjectKind().GroupVersionKind(), objKey(u))
	}

	ps := patchset.New(r.Client)
	ps.Add(existing)

	addOwnerRefs(existing, u.GetOwnerReferences())

	if err := ps.Apply(ctx); err != nil {
		return err
	}
	return nil
}

func addOwnerRefs(object client.Object, ownerRefs []metav1.OwnerReference) {
	switch object.GetObjectKind().GroupVersionKind().Kind {
	case "TanzuKubernetesRelease", "OSImage":
		return // not adding ownerRef to cluster scoped resources
	}
	for _, ownerRef := range ownerRefs {
		for _, r := range object.GetOwnerReferences() {
			if r == ownerRef {
				return
			}
		}
		object.SetOwnerReferences(append(object.GetOwnerReferences(), ownerRef))
	}
}
