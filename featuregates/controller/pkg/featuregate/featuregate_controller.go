// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregate

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/util"
)

const contextTimeout = 30 * time.Second

// FeatureGateReconciler reconciles a FeatureGate object.
type FeatureGateReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=config.tanzu.vmware.com,resources=featuregates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.tanzu.vmware.com,resources=featuregates/status,verbs=get;update;patch

// Reconcile reconciles the FeatureGate spec by computing activated, deactivated and unavailable features.
func (r *FeatureGateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	log := r.Log.WithValues("featuregate", req.NamespacedName)
	log.Info("Starting reconcile")

	featureGate := &configv1alpha1.FeatureGate{}
	if err := r.Client.Get(ctxCancel, req.NamespacedName, featureGate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List all currently available feature resources.
	features := &configv1alpha1.FeatureList{}
	if err := r.Client.List(ctxCancel, features); err != nil {
		return ctrl.Result{}, err
	}

	// Get namespaces from NamespaceSelector.
	namespaces, err := util.NamespacesMatchingSelector(ctxCancel, r.Client, &featureGate.Spec.NamespaceSelector)
	if err != nil {
		return ctrl.Result{}, err
	}
	featureGate.Status.Namespaces = namespaces

	// Compute feature states.
	activated, deactivated, unavailable := util.ComputeFeatureStates(featureGate.Spec, features.Items)
	featureGate.Status.ActivatedFeatures = activated
	featureGate.Status.DeactivatedFeatures = deactivated
	featureGate.Status.UnavailableFeatures = unavailable

	log.Info("Successfully reconciled")
	return ctrl.Result{}, r.Client.Status().Update(ctxCancel, featureGate)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FeatureGateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.FeatureGate{}).
		Watches(
			&source.Kind{Type: &configv1alpha1.Feature{}},
			handler.EnqueueRequestsFromMapFunc(r.toFeatureGateRequests)).
		Watches(
			&source.Kind{Type: &corev1.Namespace{}},
			handler.EnqueueRequestsFromMapFunc(r.toFeatureGateRequests)).
		Complete(r)
}

func (r *FeatureGateReconciler) toFeatureGateRequests(o client.Object) []reconcile.Request {
	var requests []reconcile.Request

	featuregates := &configv1alpha1.FeatureGateList{}
	if err := r.Client.List(context.Background(), featuregates); err != nil {
		r.Log.Error(err, "failed to list featuregates in event handler")
		return requests
	}

	for i := range featuregates.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: featuregates.Items[i].Namespace,
				Name:      featuregates.Items[i].Name,
			},
		})
	}

	return requests
}
