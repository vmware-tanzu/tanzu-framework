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

const (
	resourceExistenceSuccess = "found all required resources in the cluster"
)

// ReadinessProviderReconciler reconciles a ReadinessProvider object
type ReadinessProviderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinessproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinessproviders/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReadinessProvider object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
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
		state, message := evaluateResourceExistenceCondition(log.WithValues("condition", condition.Name), condition.ResourceExistenceCondition)
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

// Evaluate resourceExistenceCondition and returns status with message
func evaluateResourceExistenceCondition(log logr.Logger, resourceCriteria corev1alpha2.ResourceExistenceCondition) (state corev1alpha2.ReadinessConditionState, message string) {
	// TODO: Implementation
	log.V(2).Info("evaluating resourceExistenceCondition")
	return corev1alpha2.ConditionSuccessState, resourceExistenceSuccess
}

// Evaluate and return cumulative state of ReadinessProvider based on ReadinessConditionStatus values
func determineProviderStatus(log logr.Logger, conditionStatusList []corev1alpha2.ReadinessConditionStatus) (state corev1alpha2.ReadinessProviderState) {
	// TODO: Implementation
	log.V(2).Info("evaluating ReadinessProviderState")
	return corev1alpha2.ProviderSuccessState
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReadinessProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.ReadinessProvider{}).
		Complete(r)
}
