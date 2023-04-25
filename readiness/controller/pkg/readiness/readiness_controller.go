// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readiness

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		return ctrl.Result{}, err
	}

	updateRequired := false
	currentStatus := readiness.Status

	if len(readiness.Spec.Checks) == 0 {
		if !currentStatus.Ready {
			readiness.Status.Ready = true
			readiness.Status.CheckStatus = []corev1alpha2.CheckStatus{}
			updateRequired = true
		}
	} else {
		allChecks := make(map[string][]int)
		readiness.Status.CheckStatus = []corev1alpha2.CheckStatus{}
		providersList := &corev1alpha2.ReadinessProviderList{}

		// TODO: Find a better way to index and fetch the proviers in a single list call
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

		for i, provider := range providersList.Items {
			if _, ok := allChecks[provider.Spec.CheckRef]; !ok {
				allChecks[provider.Spec.CheckRef] = make([]int, 0)
			}

			allChecks[provider.Spec.CheckRef] = append(allChecks[provider.Spec.CheckRef], i)
		}

		for _, check := range readiness.Spec.Checks {
			checkStatus := corev1alpha2.CheckStatus{
				Name:      check.Name,
				Providers: make([]corev1alpha2.Provider, 0),
				Ready:     false,
			}

			if indices, ok := allChecks[check.Name]; ok {
				for _, index := range indices {
					provider := providersList.Items[index]

					if provider.Status.State == corev1alpha2.ProviderSuccessState {
						checkStatus.Ready = true
					}
					checkStatus.Providers = append(checkStatus.Providers, corev1alpha2.Provider{
						Name:     provider.Name,
						IsActive: provider.Status.State == corev1alpha2.ProviderSuccessState,
					})
				}

			}

			readiness.Status.CheckStatus = append(readiness.Status.CheckStatus, checkStatus)
		}

		readiness.Status.Ready = true

		for _, checkStatus := range readiness.Status.CheckStatus {
			readiness.Status.Ready = readiness.Status.Ready && checkStatus.Ready
		}

		updateRequired = isUpdateRequired(currentStatus, readiness.Status)

	}

	if updateRequired {
		time := metav1.Now()
		readiness.Status.LastUpdatedTime = &time

		err = r.Client.Status().Update(ctxCancel, readiness)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func isUpdateRequired(oldStatus, newStatus corev1alpha2.ReadinessStatus) bool {
	oldCheckStatusByName := make(map[string]corev1alpha2.CheckStatus)
	for _, checkStatus := range oldStatus.CheckStatus {
		oldCheckStatusByName[checkStatus.Name] = checkStatus
	}

	newCheckStatusByName := make(map[string]corev1alpha2.CheckStatus)
	for _, checkStatus := range newStatus.CheckStatus {
		newCheckStatusByName[checkStatus.Name] = checkStatus
	}

	if len(oldCheckStatusByName) != len(newCheckStatusByName) {
		return true
	}

	for name, oldCheckStatus := range oldCheckStatusByName {
		newCheckStatus, ok := newCheckStatusByName[name]
		if !ok {
			return true
		}

		if oldCheckStatus.Ready != newCheckStatus.Ready {
			return true
		}

		if len(oldCheckStatus.Providers) != len(newCheckStatus.Providers) {
			return true
		}

		for _, oldProvider := range oldCheckStatus.Providers {
			found := false
			for _, newProvider := range newCheckStatus.Providers {
				if oldProvider.Name == newProvider.Name {
					found = true

					if oldProvider.IsActive != newProvider.IsActive {
						return true
					}

					break
				}
			}

			if !found {
				return true
			}
		}
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReadinessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mgr.GetFieldIndexer().IndexField(context.Background(), &corev1alpha2.ReadinessProvider{}, "spec.checkRef", func(rawObj client.Object) []string {
		provider := rawObj.(*corev1alpha2.ReadinessProvider)
		return []string{provider.Spec.CheckRef}
	})

	mgr.GetFieldIndexer().IndexField(context.Background(), &corev1alpha2.Readiness{}, "spec.checks.name", func(rawObj client.Object) []string {
		provider := rawObj.(*corev1alpha2.Readiness)

		keys := []string{}
		for _, check := range provider.Spec.Checks {
			keys = append(keys, check.Name)
		}

		return keys
	})

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

	readinessList := &corev1alpha2.ReadinessList{}
	err := r.Client.List(ctxCancel, readinessList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.checks.name", provider.Spec.CheckRef),
	})
	if err != nil {
		r.Log.Error(err, "error while updating readiness status")
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}

	for _, readiness := range readinessList.Items {
		for _, check := range readiness.Spec.Checks {
			if check.Name == provider.Spec.CheckRef {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: readiness.Name,
					},
				})
			}
		}
	}

	return requests
}
