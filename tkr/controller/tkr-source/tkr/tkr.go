// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkr

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/util/patchset"
)

type Reconciler struct {
	Log    logr.Logger
	Client client.Client
}

type Config struct {
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1.TanzuKubernetesRelease{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("tkr_legacy_labeler").
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	tkr := &runv1.TanzuKubernetesRelease{}
	if err := r.Client.Get(ctx, req.NamespacedName, tkr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrapf(err, "getting TKR '%s'", req.Name)
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

	return ctrl.Result{}, r.doReconcile(tkr)
}

func (r *Reconciler) doReconcile(tkr *runv1.TanzuKubernetesRelease) error {
	if tkr.Spec.BootstrapPackages == nil {
		r.Log.Info("Labeling TKR with run.tanzu.vmware.com/legacy-tkr: '' as it does not appear to have bootstrap packages",
			"name", tkr.Name)
		getMap(&tkr.Labels)[runv1.LabelLegacyTKR] = ""
	}
	return nil
}

// getMap returns the map (creates it first if the map is nil). mp has to be a pointer to the variable holding the map,
// so that we could save the newly created map.
// Pre-reqs: mp != nil
func getMap(mp *map[string]string) map[string]string { // nolint:gocritic // suppress warning: ptrToRefParam: consider `mp' to be of non-pointer type (gocritic)
	if *mp == nil {
		*mp = map[string]string{}
	}
	return *mp
}
