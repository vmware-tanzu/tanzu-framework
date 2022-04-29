// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkr provides the reconciler for the TKR labeling controller / TKR resolver cache refresher.
package tkr

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
)

type Reconciler struct {
	Log    logr.Logger
	Client client.Client
	Cache  resolver.Cache
	Object client.Object
}

func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.Object.DeepCopyObject().(client.Object)).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	object := r.Object.DeepCopyObject().(client.Object)

	if err := r.Client.Get(ctx, req.NamespacedName, object); err != nil {
		if apierrors.IsNotFound(err) {
			object.SetName(req.Name)
			r.Cache.Remove(object)
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

	ps.Add(object)

	r.Cache.Add(object)
	r.Log.Info("added", "name", req.Name)
	return ctrl.Result{}, nil
}
