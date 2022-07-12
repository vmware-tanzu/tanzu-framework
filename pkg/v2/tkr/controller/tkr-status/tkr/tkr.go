// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkr provides the TKR reconciler for the TKR status controller / TKR resolver cache refresher.
package tkr

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
)

type Reconciler struct {
	Ctx    context.Context
	Log    logr.Logger
	Client client.Client
	Cache  resolver.Cache
	Config Config
}

type Config struct {
	Namespace string
}

const indexTKROSImages = ".index.tkrOSImages"
const indexTKRBootstrapPackages = ".index.tkrBootstrapPackages"

func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		r.Ctx, &runv1.TanzuKubernetesRelease{}, indexTKROSImages, tkrOSImages,
	); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(
		r.Ctx, &runv1.TanzuKubernetesRelease{}, indexTKRBootstrapPackages, tkrBootstrapPackages,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1.TanzuKubernetesRelease{}, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &runv1.OSImage{}},
			handler.EnqueueRequestsFromMapFunc(r.osImageToTKRs),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &kapppkgv1.Package{}},
			handler.EnqueueRequestsFromMapFunc(r.pkgToTKRs),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &runv1.ClusterBootstrapTemplate{}},
			handler.EnqueueRequestsFromMapFunc(r.cbtToTKRs),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

func tkrOSImages(o client.Object) []string {
	tkr := o.(*runv1.TanzuKubernetesRelease)
	result := make([]string, len(tkr.Spec.OSImages))
	for i, osImageRef := range tkr.Spec.OSImages {
		result[i] = osImageRef.Name
	}
	return result
}

func tkrBootstrapPackages(o client.Object) []string {
	tkr := o.(*runv1.TanzuKubernetesRelease)
	result := make([]string, len(tkr.Spec.BootstrapPackages))
	for i, pkgRef := range tkr.Spec.BootstrapPackages {
		result[i] = pkgRef.Name
	}
	return result
}

func (r *Reconciler) osImageToTKRs(object client.Object) []reconcile.Request {
	osImage := object.(*runv1.OSImage)

	tkrList := &runv1.TanzuKubernetesReleaseList{}
	if err := r.Client.List(r.Ctx, tkrList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexTKROSImages, osImage.Name),
	}); err != nil {
		r.Log.Error(err, "error listing TKRs with OSImage", "name", osImage.Name)
		return nil
	}

	count := len(tkrList.Items)
	result := make([]ctrl.Request, count)
	for i := range tkrList.Items {
		result[i].Name = tkrList.Items[i].Name
	}

	r.Log.Info("Enque TKRs for OSImage", "name", osImage.Name, "count", count)
	return result
}

func (r *Reconciler) pkgToTKRs(object client.Object) []reconcile.Request {
	pkg := object.(*kapppkgv1.Package)
	if pkg.Namespace != r.Config.Namespace {
		return nil
	}

	tkrList := &runv1.TanzuKubernetesReleaseList{}
	if err := r.Client.List(r.Ctx, tkrList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexTKRBootstrapPackages, pkg.Name),
	}); err != nil {
		r.Log.Error(err, "error listing TKRs with bootstrap Package", "name", pkg.Name)
		return nil
	}

	count := len(tkrList.Items)
	result := make([]ctrl.Request, count)
	for i := range tkrList.Items {
		result[i].Name = tkrList.Items[i].Name
	}

	r.Log.Info("Enque TKRs for bootstrap Package", "name", pkg.Name, "count", count)
	return result
}

func (r *Reconciler) cbtToTKRs(object client.Object) []reconcile.Request {
	if object.GetNamespace() != r.Config.Namespace {
		return nil
	}

	r.Log.Info("Enque TKR for CBT", "name", object.GetName())
	return []ctrl.Request{{NamespacedName: types.NamespacedName{Name: object.GetName()}}}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	tkr := &runv1.TanzuKubernetesRelease{}

	if err := r.Client.Get(ctx, req.NamespacedName, tkr); err != nil {
		if apierrors.IsNotFound(err) {
			tkr.SetName(req.Name)
			r.Cache.Remove(tkr)
			r.Log.Info("removed", "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	ps := patchset.New(r.Client)
	defer func() {
		// apply patches unless an error is being returned
		if retErr != nil {
			return
		}
		if err := ps.Apply(ctx); err != nil {
			if err = kerrors.FilterOut(err, apierrors.IsConflict); err == nil {
				// retry if someone updated an object we wanted to patch
				result = ctrl.Result{Requeue: true}
			}
			retErr = errors.Wrap(err, "applying patches to TKRs")
		}
	}()

	ps.Add(tkr)

	if err := r.setValidCondition(ctx, tkr); err != nil {
		return ctrl.Result{}, err
	}

	r.Cache.Add(tkr)
	r.Log.Info("added", "name", req.Name)
	return ctrl.Result{}, nil
}
