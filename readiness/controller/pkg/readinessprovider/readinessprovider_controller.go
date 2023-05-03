// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readinessprovider

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/readiness/controller/pkg/constants"
)

// ReadinessProviderReconciler reconciles a ReadinessProvider object
type ReadinessProviderReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	ResourceExistenceCondition func(context.Context, *corev1alpha2.ResourceExistenceCondition) (corev1alpha2.ReadinessConditionState, string)
	ShellScriptCondition       func(context.Context, *corev1alpha2.ShellScriptCondition) (corev1alpha2.ReadinessConditionState, string)
}

//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinessproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinessproviders/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReadinessProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, constants.ContextTimeout)
	defer cancel()

	log := r.Log.WithValues("readinessprovider", req.NamespacedName)
	log.Info("starting reconcile")

	readinessProvider := corev1alpha2.ReadinessProvider{}
	result := ctrl.Result{}

	if err := r.Client.Get(ctxCancel, req.NamespacedName, &readinessProvider); err != nil {
		log.Error(err, "unable to fetch ReadinessProvider")
		return result, client.IgnoreNotFound(err)
	}

	// Evaluate provider conditions
	readinessProvider.Status.Conditions = make([]corev1alpha2.ReadinessConditionStatus, len(readinessProvider.Spec.Conditions))

	for i, condition := range readinessProvider.Spec.Conditions {
		readinessProvider.Status.Conditions[i].Name = condition.Name
		var state corev1alpha2.ReadinessConditionState
		var message string
		if condition.ResourceExistenceCondition != nil {
			state, message = r.ResourceExistenceCondition(ctxCancel, condition.ResourceExistenceCondition)
		} else if condition.ShellScriptCondition != nil {
			state, message = r.ShellScriptCondition(ctxCancel, condition.ShellScriptCondition)
		}
		readinessProvider.Status.Conditions[i].State = state
		readinessProvider.Status.Conditions[i].Message = message
	}

	readinessProvider.Status.State = determineProviderStatus(log, readinessProvider.Status.Conditions)

	log.Info("Successfully reconciled")

	// ReadinessProvider can be auto-evaluated periodically if
	// RepeatInterval is configured in the spec
	if readinessProvider.Spec.RepeatInterval != nil {
		repeatAfterDuration := readinessProvider.Spec.RepeatInterval.Duration
		log.V(2).Info("requeing for evaluation", "after", repeatAfterDuration)
		result.RequeueAfter = repeatAfterDuration
	}

	return result, r.Status().Update(ctxCancel, &readinessProvider)
}

// Evaluate and return cumulative state of ReadinessProvider based on ReadinessConditionStatus values
func determineProviderStatus(log logr.Logger, conditionStatusList []corev1alpha2.ReadinessConditionStatus) (state corev1alpha2.ReadinessProviderState) {
	inProgress := false
	for _, status := range conditionStatusList {
		if status.State == corev1alpha2.ConditionFailureState {
			return corev1alpha2.ProviderFailureState
		} else if status.State == corev1alpha2.ConditionInProgressState {
			inProgress = true
		}
	}
	log.V(2).Info("evaluating ReadinessProviderState")
	if inProgress {
		return corev1alpha2.ProviderInProgressState
	}

	return corev1alpha2.ProviderSuccessState
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReadinessProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.ReadinessProvider{}).
		Complete(r)
}
