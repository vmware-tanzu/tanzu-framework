// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package feature

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/util"
)

const contextTimeout = 30 * time.Second

// FeatureReconciler reconciles a Feature object.
type FeatureReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=featuregates,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=featuregates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=features,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.tanzu.vmware.com,resources=features/status,verbs=get;update;patch

// Reconcile reconciles the FeatureGate spec by computing activated, deactivated and unavailable features.
func (r *FeatureReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	log := r.Log.WithValues("feature", req.NamespacedName)
	log.Info("Starting reconcile")

	feature := &corev1alpha2.Feature{}
	if err := r.Client.Get(ctxCancel, req.NamespacedName, feature); err != nil {
		if apierrors.IsNotFound(err) {
			// Reconcile deleted feature CR. If the Feature CR is deleted, check if that feature is being gated by
			// any FeatureGate.
			// 1. If the feature is being gated by a FeatureGate, update the status of FeatureGate by updating the
			// feature reference result status as Invalid.
			// 2. If the feature is not gated by any FeatureGate, finish the reconciliation
			if err := reconcileDeletedFeature(ctx, r.Client, req.NamespacedName.Name); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the feature is part of any FeatureGate spec
	featureGate, found, err := util.GetFeatureGateForFeature(ctx, r.Client, feature.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If the feature is not found in any FeatureGate spec. Check if it exists in the FeatureGate status.
	// If found in the FeatureGate status, remove its entry from Results and update the feature status to default
	// activation
	if !found {
		if err := reconcileFeatureNotInFeatureGateSpec(ctx, r.Client, feature); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// If the feature is found in any FeatureGate spec, update the Results in FeatureGate status and the feature status
	// to the intent specified in the FeatureGate spec
	if err := reconcileFeatureInFeatureGateSpec(ctx, r.Client, featureGate, feature); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// reconcileFeatureInFeatureGateSpec reconciles Feature resource that is present in FeatureGate spec
func reconcileFeatureInFeatureGateSpec(ctx context.Context, c client.Client, featureGate *corev1alpha2.FeatureGate, feature *corev1alpha2.Feature) error {
	policy := corev1alpha2.GetPolicyForStabilityLevel(feature.Spec.Stability)
	featureReference, _ := util.GetFeatureReferenceFromFeatureGate(featureGate, feature.Name)
	featureResult, activate := applyPolicyToComputeFeatureResultAndActivation(policy, featureReference)

	// Update FeatureGate status
	featureGate.Status.FeatureReferenceResults = computeFeatureGateStatusResults(featureGate.Status, featureResult, true)
	if err := c.Status().Update(ctx, featureGate); err != nil {
		return fmt.Errorf("could not update %s FeatureGate status :%w", featureGate.Name, err)
	}

	// Update Feature status to the intent specified in the FeatureGate spec
	feature.Status.Activated = activate
	if err := c.Update(ctx, feature); err != nil {
		return fmt.Errorf("could not update %s Feature status :%w", feature.Name, err)
	}
	return nil
}

// reconcileDeletedFeature reconciles Feature resource that has been deleted
func reconcileDeletedFeature(ctx context.Context, c client.Client, featureName string) error {
	// Check if the feature is part of any FeatureGate spec and update the feature Result in FeatureGate status to
	// Invalid
	featureGate, found, err := util.GetFeatureGateForFeature(ctx, c, featureName)
	if err != nil {
		return err
	}
	if found {
		featureGate.Status.FeatureReferenceResults = computeFeatureGateStatusResults(featureGate.Status, corev1alpha2.FeatureReferenceResult{
			Name:    featureName,
			Status:  corev1alpha2.InvalidReferenceStatus,
			Message: "Feature does not exist in cluster",
		}, true)
		if err := c.Status().Update(ctx, featureGate); err != nil {
			return fmt.Errorf("could not update %s FeatureGate status :%w", featureGate.Name, err)
		}
	}
	return nil
}

// reconcileFeatureNotInFeatureGateSpec reconciles Feature resource that is not found in any FeatureGate spec
func reconcileFeatureNotInFeatureGateSpec(ctx context.Context, c client.Client, feature *corev1alpha2.Feature) error {
	// If feature is not part of any FeatureGate spec, check if it exists in any FeatureGate status.
	// If found in the FeatureGate status, remove its entry from Results in FeatureGate status and update the
	// feature status to default activation
	featureGate, found, err := util.GetFeatureGateWithFeatureInStatus(ctx, c, feature.Name)
	policy := corev1alpha2.GetPolicyForStabilityLevel(feature.Spec.Stability)
	if err != nil {
		return err
	}
	if found {
		// Remove feature from FeatureGate status
		featureGate.Status.FeatureReferenceResults = computeFeatureGateStatusResults(featureGate.Status, corev1alpha2.FeatureReferenceResult{
			Name: feature.Name,
		}, false)
		if err := c.Status().Update(ctx, featureGate); err != nil {
			return fmt.Errorf("could not update %s FeatureGate status :%w", featureGate.Name, err)
		}
	}
	// Update Feature status to set feature as deactivated
	feature.Status.Activated = policy.DefaultActivation
	if err := c.Update(ctx, feature); err != nil {
		return fmt.Errorf("could not update %s Feature status :%w", feature.Name, err)
	}
	return nil
}

// applyPolicyToComputeFeatureResultAndActivation applies stability level policy and returns feature result for
// FeatureGate status and feature activate status
func applyPolicyToComputeFeatureResultAndActivation(policy corev1alpha2.Policy, featureRef corev1alpha2.FeatureReference) (corev1alpha2.FeatureReferenceResult, bool) {
	activated := policy.DefaultActivation
	result := corev1alpha2.FeatureReferenceResult{Name: featureRef.Name}
	// Check for immutability and change in intent of feature status
	if policy.Immutable && policy.DefaultActivation != featureRef.Activate {
		result.Status = corev1alpha2.InvalidReferenceStatus
		result.Message = "Feature could not be toggled because it is immutable"
	} else if policy.VoidsWarranty && !featureRef.PermanentlyVoidAllSupportGuarantees &&
		policy.DefaultActivation != featureRef.Activate {
		result.Status = corev1alpha2.InvalidReferenceStatus
		result.Message = "The stability level of this feature indicates that it should not be activated in " +
			"production environments. To activate the feature, you must agree to permanently void all support " +
			"guarantees for this environment by setting featureRef.permanentlyVoidAllSupportGuarantees to true."
	} else {
		result.Status = corev1alpha2.AppliedReferenceStatus
		activated = featureRef.Activate
	}
	return result, activated
}

// computeFeatureGateStatusResults takes the featuregate status and feature result, and returns the computed featuregate
// results slice
func computeFeatureGateStatusResults(featureGateStatus corev1alpha2.FeatureGateStatus, featureResult corev1alpha2.FeatureReferenceResult, upsert bool) []corev1alpha2.FeatureReferenceResult {
	if featureResult.Message == "" {
		if featureResult.Status == corev1alpha2.InvalidReferenceStatus {
			featureResult.Message = "Invalid operation, feature cannot be toggled"
		} else {
			featureResult.Message = "Feature has been successfully toggled"
		}
	}

	results := featureGateStatus.FeatureReferenceResults
	if upsert {
		for i, result := range featureGateStatus.FeatureReferenceResults {
			if result.Name == featureResult.Name {
				results[i].Status = featureResult.Status
				results[i].Message = featureResult.Message
				return results
			}
		}

		results = append(results, corev1alpha2.FeatureReferenceResult{
			Name:    featureResult.Name,
			Status:  featureResult.Status,
			Message: featureResult.Message,
		})
		return results
	}
	// remove the feature result from the status if upsert is false
	for i, result := range featureGateStatus.FeatureReferenceResults {
		if result.Name == featureResult.Name {
			results = append(results[:i], results[i+1:]...)
			return results
		}
	}
	return results
}

// SetupWithManager sets up the controller with the Manager.
func (r *FeatureReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.Feature{}).
		Watches(
			&source.Kind{Type: &corev1alpha2.FeatureGate{}},
			handler.EnqueueRequestsFromMapFunc(r.toFeatureRequests)).
		Complete(r)
}

func (r *FeatureReconciler) toFeatureRequests(o client.Object) []reconcile.Request {
	var requests []reconcile.Request

	featureGate := &corev1alpha2.FeatureGate{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{
		Name: o.GetName(),
	}, featureGate); err != nil {
		r.Log.Error(err, "failed to get featuregate in event handler",
			"Featuregate", o.GetName())
		return requests
	}

	// Enqueue all the features that are part of the updated FeatureGate spec
	features := sets.String{}
	for _, feature := range featureGate.Spec.Features {
		features.Insert(feature.Name)
	}

	// To handle the case where feature reference has removed from the FeatureGate spec
	for _, result := range featureGate.Status.FeatureReferenceResults {
		features.Insert(result.Name)
	}

	for feature := range features {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: feature,
			},
		})
	}
	return requests
}
