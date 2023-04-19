// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readiness

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Readiness object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *ReadinessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	readiness := &corev1alpha2.Readiness{}
	err := r.Client.Get(ctx, req.NamespacedName, readiness)
	if err != nil {
		r.Log.Error(err, "Error while fetching readiness")
		return ctrl.Result{}, err
	}

	updateRequired := false
	oldStatus := readiness.Status

	if len(readiness.Spec.Checks) == 0 {
		if !oldStatus.Ready {
			readiness.Status.Ready = true
			readiness.Status.CheckStatus = []corev1alpha2.CheckStatus{}
			updateRequired = true
		}
	} else {
		allChecks := make(map[string][]int)
		readiness.Status.CheckStatus = []corev1alpha2.CheckStatus{}
		providersList := &corev1alpha2.ReadinessProviderList{}
		err = r.Client.List(ctx, providersList)
		if err != nil {
			r.Log.Error(err, "Error while fetching readiness providers")
			return ctrl.Result{}, err
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

			indices, ok := allChecks[check.Name]
			if ok {
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

		updateRequired = isUpdateRequired(oldStatus, readiness.Status)

	}

	if updateRequired {
		time := metav1.Time{Time: time.Now()}
		readiness.Status.LastComputedTime = &time

		err = r.Client.Status().Update(ctx, readiness)
		if err != nil {
			r.Log.Error(err, "Error while updating readiness status")
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.Readiness{}).
		Complete(r)
}
