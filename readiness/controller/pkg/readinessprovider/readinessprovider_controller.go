// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readinessprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	capabilitiesdiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-framework/util/kubeclient"
)

const (
	requeueInterval = 60 * time.Second
	contextTimeout  = 60 * time.Second
)

// ReadinessProviderReconciler reconciles a ReadinessProvider object
type ReadinessProviderReconciler struct {
	client.Client
	Clientset                  *kubernetes.Clientset
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	ResourceExistenceCondition func(context.Context, *capabilitiesdiscovery.ClusterQueryClient, *corev1alpha2.ResourceExistenceCondition, string) (corev1alpha2.ReadinessConditionState, string)
	RestConfig                 *rest.Config
	DefaultQueryClient         *capabilitiesdiscovery.ClusterQueryClient
}

//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinessproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=readinessproviders/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReadinessProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	log := r.Log.WithValues("readinessprovider", req.NamespacedName)
	log.Info("starting reconcile")

	readinessProvider := corev1alpha2.ReadinessProvider{}
	result := ctrl.Result{
		RequeueAfter: requeueInterval,
	}

	if err := r.Client.Get(ctxCancel, req.NamespacedName, &readinessProvider); err != nil {
		log.Error(err, "unable to fetch ReadinessProvider")
		return result, client.IgnoreNotFound(err)
	}

	var clusterQueryClient *capabilitiesdiscovery.ClusterQueryClient

	// If provided in the spec, use the serviceAccount for evaluating conditions
	if readinessProvider.Spec.ServiceAccountRef != nil {
		cfg, err := kubeclient.GetConfigForServiceAccount(ctx, r.Clientset, r.RestConfig, readinessProvider.Spec.ServiceAccountRef.Namespace, readinessProvider.Spec.ServiceAccountRef.Name)
		if err != nil {
			readinessProvider.Status.Message = err.Error()
			readinessProvider.Status.State = corev1alpha2.ProviderFailureState
			readinessProvider.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
			return result, r.Status().Update(ctxCancel, &readinessProvider)
		}
		clusterQueryClient, err = capabilitiesdiscovery.NewClusterQueryClientForConfig(cfg)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to create ClusterQueryClient: %w", err)
		}
	} else {
		clusterQueryClient = r.DefaultQueryClient
	}

	// Evaluate provider conditions
	readinessProvider.Status.Conditions = make([]corev1alpha2.ReadinessConditionStatus, len(readinessProvider.Spec.Conditions))

	for i, condition := range readinessProvider.Spec.Conditions {
		readinessProvider.Status.Conditions[i].Name = condition.Name
		var state corev1alpha2.ReadinessConditionState
		var message string
		state, message = r.ResourceExistenceCondition(ctxCancel, clusterQueryClient, condition.ResourceExistenceCondition, condition.Name)
		readinessProvider.Status.Conditions[i].State = state
		readinessProvider.Status.Conditions[i].Message = message
	}

	readinessProvider.Status.State, readinessProvider.Status.Message = determineProviderStatus(log, readinessProvider.Status.Conditions)

	log.Info("Successfully reconciled")

	return result, r.Status().Update(ctxCancel, &readinessProvider)
}

// Evaluate and return cumulative state of ReadinessProvider based on ReadinessConditionStatus values
func determineProviderStatus(log logr.Logger, conditionStatusList []corev1alpha2.ReadinessConditionStatus) (state corev1alpha2.ReadinessProviderState, message string) {
	inProgress := false
	for _, status := range conditionStatusList {
		if status.State == corev1alpha2.ConditionFailureState {
			return corev1alpha2.ProviderFailureState, "one or more condition(s) failed"
		} else if status.State == corev1alpha2.ConditionInProgressState {
			inProgress = true
		}
	}
	log.V(2).Info("evaluating ReadinessProviderState")
	if inProgress {
		return corev1alpha2.ProviderInProgressState, "one or more condition(s) are being evaluated"
	}

	return corev1alpha2.ProviderSuccessState, "all condition(s) passed"
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReadinessProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.ReadinessProvider{}).
		Complete(r)
}
