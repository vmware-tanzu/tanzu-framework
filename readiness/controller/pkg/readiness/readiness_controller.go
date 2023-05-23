// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readiness

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

const contextTimeout = 30 * time.Second

// ReadinessReconciler reconciles a Readiness object
type ReadinessReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinesses/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReadinessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	readiness := &corev1alpha2.Readiness{}
	err := r.Client.Get(ctxCancel, req.NamespacedName, readiness)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(readiness.Spec.Checks) == 0 && !readiness.Status.Ready {
		readiness.Status.Ready = true
		readiness.Status.CheckStatus = []corev1alpha2.CheckStatus{}
		return ctrl.Result{}, r.Client.Status().Update(ctxCancel, readiness)
	}

	readiness.Status.CheckStatus = []corev1alpha2.CheckStatus{}
	providersList := &corev1alpha2.ReadinessProviderList{}

	// TODO: Find a better way to index and fetch the providers in a single list call
	for _, check := range readiness.Spec.Checks {
		providers := &corev1alpha2.ReadinessProviderList{}
		err = r.Client.List(ctxCancel, providers, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("spec.checkRef", check.Name),
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		providersList.Items = append(providersList.Items, providers.Items...)
	}

	// uniqueProviders contain all the providers that satisfy at least one of the checks
	uniqueProviders := make([]corev1alpha2.ReadinessProvider, 0)
	uniqueProviderNames := make(map[string]bool)

	for i := 0; i < len(providersList.Items); i++ {
		if _, ok := uniqueProviderNames[providersList.Items[i].Name]; !ok {
			uniqueProviders = append(uniqueProviders, providersList.Items[i])
			uniqueProviderNames[providersList.Items[i].Name] = true
		}
	}

	// allChecks contain the check names and the list of associated providers
	allChecks := make(map[string][]int)
	for i := 0; i < len(uniqueProviders); i++ {
		for _, checkRef := range uniqueProviders[i].Spec.CheckRefs {
			allChecks[checkRef] = append(allChecks[checkRef], i)
		}
	}

	// TODO: Handle Composite checks
	for _, check := range readiness.Spec.Checks {
		checkStatusUpdate := corev1alpha2.CheckStatus{
			Name:      check.Name,
			Providers: make([]corev1alpha2.Provider, 0),
			Ready:     false,
		}

		for _, index := range allChecks[check.Name] {
			provider := uniqueProviders[index]

			checkStatusUpdate.Ready = checkStatusUpdate.Ready || (provider.Status.State == corev1alpha2.ProviderSuccessState)

			checkStatusUpdate.Providers = append(checkStatusUpdate.Providers, corev1alpha2.Provider{
				Name:     provider.Name,
				IsActive: provider.Status.State == corev1alpha2.ProviderSuccessState,
			})
		}
		readiness.Status.CheckStatus = append(readiness.Status.CheckStatus, checkStatusUpdate)
	}

	readiness.Status.Ready = true

	for _, checkStatus := range readiness.Status.CheckStatus {
		readiness.Status.Ready = readiness.Status.Ready && checkStatus.Ready
	}

	return ctrl.Result{}, r.Client.Status().Update(ctxCancel, readiness)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReadinessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1alpha2.ReadinessProvider{}, "spec.checkRef", func(rawObj client.Object) []string {
		provider := rawObj.(*corev1alpha2.ReadinessProvider)
		var indices []string
		indices = append(indices, provider.Spec.CheckRefs...)
		return indices
	})

	if err != nil {
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(context.Background(), &corev1alpha2.Readiness{}, "spec.checks.name", func(rawObj client.Object) []string {
		provider := rawObj.(*corev1alpha2.Readiness)

		keys := []string{}
		for _, check := range provider.Spec.Checks {
			keys = append(keys, check.Name)
		}

		return keys
	})

	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.Readiness{}).
		Watches(
			&source.Kind{Type: &corev1alpha2.ReadinessProvider{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForReadinessProvider),
		).
		Complete(r)
}

func (r *ReadinessReconciler) findObjectsForReadinessProvider(readinessProviderObject client.Object) []reconcile.Request {
	ctxCancel, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	provider, _ := readinessProviderObject.(*corev1alpha2.ReadinessProvider)

	totalReadinessList := &corev1alpha2.ReadinessList{}
	for _, checkRef := range provider.Spec.CheckRefs {
		readinessList := &corev1alpha2.ReadinessList{}
		err := r.Client.List(ctxCancel, readinessList, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("spec.checks.name", checkRef),
		})
		if err != nil {
			r.Log.Error(err, "error while updating readiness status")
			return []reconcile.Request{}
		}
		totalReadinessList.Items = append(totalReadinessList.Items, readinessList.Items...)
	}

	requests := []reconcile.Request{}

	for i := 0; i < len(totalReadinessList.Items); i++ {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: totalReadinessList.Items[i].Name,
			},
		})
	}

	return requests
}
