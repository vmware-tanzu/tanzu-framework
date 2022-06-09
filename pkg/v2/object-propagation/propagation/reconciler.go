// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package propagation provides object-propagation controller reconciler.
package propagation

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
)

type Reconciler struct {
	Ctx context.Context
	Log logr.Logger

	Client client.Client
	Config Config
}

type Config struct {
	ObjectType       client.Object
	ObjectListType   client.ObjectList
	SourceNamespace  string
	SourceSelector   labels.Selector
	TargetNSSelector labels.Selector
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			r.Config.ObjectType,
			builder.WithPredicates(
				predicate.NewPredicateFuncs(r.matchesSourceSelectorWithinSourceNamespace),
				predicate.ResourceVersionChangedPredicate{})).
		Watches(
			&source.Kind{Type: &corev1.Namespace{}},
			handler.EnqueueRequestsFromMapFunc(r.toAllSourceObjectsForNonExcludedNamespace),
			builder.WithPredicates(predicate.LabelChangedPredicate{})).
		Watches(
			&source.Kind{Type: r.Config.ObjectType},
			handler.EnqueueRequestsFromMapFunc(r.toSourceObject),
			builder.WithPredicates(
				predicate.NewPredicateFuncs(r.matchesSourceSelectorWithinSourceNamespace),
				predicate.ResourceVersionChangedPredicate{})).
		Named(fmt.Sprintf("object_propagator_%s", r.Config.ObjectType.GetObjectKind().GroupVersionKind().Kind)).
		Complete(r)
}

func (r *Reconciler) matchesSourceSelectorWithinSourceNamespace(sourceObj client.Object) bool {
	return sourceObj.GetNamespace() == r.Config.SourceNamespace &&
		r.Config.SourceSelector.Matches(labels.Set(sourceObj.GetLabels()))
}

func (r *Reconciler) toAllSourceObjectsForNonExcludedNamespace(ns client.Object) []ctrl.Request {
	if !ns.GetDeletionTimestamp().IsZero() {
		return nil
	}
	if !r.Config.TargetNSSelector.Matches(labels.Set(ns.GetLabels())) {
		return nil
	}

	list := r.Config.ObjectListType.DeepCopyObject().(client.ObjectList)
	if err := r.Client.List(r.Ctx, list, &client.ListOptions{
		Namespace:     r.Config.SourceNamespace,
		LabelSelector: r.Config.SourceSelector,
	}); err != nil {
		r.Log.Error(err, "error listing source objects")
		return nil
	}

	items := reflect.ValueOf(list).Elem().FieldByName("Items")
	result := make([]ctrl.Request, items.Len())

	for i := 0; i < items.Len(); i++ {
		objValue := items.Index(i)
		object := objValue.Addr().Interface().(client.Object)

		result[i].Namespace = object.GetNamespace()
		result[i].Name = object.GetName()
	}

	return result
}

func (r *Reconciler) toSourceObject(targetObj client.Object) []ctrl.Request {
	if targetObj.GetNamespace() == r.Config.SourceNamespace {
		return nil // target object cannot be in the source namespace
	}

	ns := &corev1.Namespace{}
	if err := r.Client.Get(r.Ctx, types.NamespacedName{Name: targetObj.GetNamespace()}, ns); err != nil {
		if !apierrors.IsNotFound(err) {
			r.Log.Error(err, "error getting namespace resource for target object",
				"type", targetObj.GetObjectKind().GroupVersionKind(),
				"object", types.NamespacedName{Namespace: targetObj.GetNamespace(), Name: targetObj.GetName()})
		}
		return nil
	}
	if !ns.DeletionTimestamp.IsZero() || !r.Config.TargetNSSelector.Matches(labels.Set(ns.GetLabels())) {
		return nil
	}

	return []ctrl.Request{{NamespacedName: types.NamespacedName{
		Namespace: r.Config.SourceNamespace,
		Name:      targetObj.GetName(),
	}}}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sourceObj := r.Config.ObjectType.DeepCopyObject().(client.Object)

	if err := r.Client.Get(ctx, req.NamespacedName, sourceObj); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		// sourceObj not found
		sourceObj.SetNamespace(req.Namespace)
		sourceObj.SetName(req.Name)
		sourceObj.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
	}

	nsList := &corev1.NamespaceList{}
	if err := r.Client.List(ctx, nsList, client.MatchingLabelsSelector{Selector: r.Config.TargetNSSelector}); err != nil {
		return ctrl.Result{}, err
	}

	var errs []error
	for i := range nsList.Items {
		nsObj := &nsList.Items[i]
		// skip if nsObj is the source namespace or if it is being deleted
		if nsObj.Name == r.Config.SourceNamespace || !nsObj.DeletionTimestamp.IsZero() {
			continue
		}
		errs = append(errs, r.propagate(ctx, nsObj.Name, sourceObj))
	}

	err := kerrors.NewAggregate(errs)
	errSansConflict := kerrors.FilterOut(err, apierrors.IsConflict)

	return ctrl.Result{Requeue: err != nil}, errSansConflict
}

func (r *Reconciler) propagate(ctx context.Context, targetNS string, sourceObj client.Object) error {
	targetObj := r.Config.ObjectType.DeepCopyObject().(client.Object)

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: targetNS, Name: sourceObj.GetName()}, targetObj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		// targetObj not found
		if !sourceObj.GetDeletionTimestamp().IsZero() { // sourceObj is being deleted
			return nil // nothing else to do
		}

		// targetObj not found, create
		r.Log.Info("Creating object", "type", sourceObj.GetObjectKind().GroupVersionKind(),
			"namespace", targetNS, "name", sourceObj.GetName())
		targetObj.SetNamespace(targetNS)
		if err := overwrite(targetObj, sourceObj); err != nil {
			return err
		}
		return r.Client.Create(ctx, targetObj)
	}

	// targetObj exists
	if !sourceObj.GetDeletionTimestamp().IsZero() { // sourceObj is being deleted
		r.Log.Info("Deleting object", "type", sourceObj.GetObjectKind().GroupVersionKind(),
			"namespace", targetNS, "name", sourceObj.GetName())
		err := r.Client.Delete(ctx, targetObj)
		return errors.Wrap(err, "deleting target object")
	}

	// targetObj exists, patch
	r.Log.Info("Patching object", "type", sourceObj.GetObjectKind().GroupVersionKind(),
		"namespace", targetNS, "name", sourceObj.GetName())
	ps := patchset.New(r.Client)
	ps.Add(targetObj)

	if err := overwrite(targetObj, sourceObj); err != nil {
		return err
	}
	return ps.Apply(ctx)
}

func overwrite(targetObj, sourceObj client.Object) error {
	orig := targetObj.DeepCopyObject().(client.Object)

	if err := mergo.Merge(targetObj, sourceObj, mergo.WithOverwriteWithEmptyValue); err != nil {
		return err
	}

	restoreMeta(targetObj, orig)

	return nil
}

func restoreMeta(targetObj, orig client.Object) {
	targetObj.SetNamespace(orig.GetNamespace())
	targetObj.SetUID(orig.GetUID())
	targetObj.SetResourceVersion(orig.GetResourceVersion())
	targetObj.SetGeneration(orig.GetGeneration())
	targetObj.SetSelfLink(orig.GetSelfLink())
	targetObj.SetCreationTimestamp(orig.GetCreationTimestamp())
}
