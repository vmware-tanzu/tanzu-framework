// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var featuregatelog = logf.Log.WithName("featuregate-resource").WithValues("apigroup", "core")

var cl client.Client

func getScheme() (*runtime.Scheme, error) {
	s, err := SchemeBuilder.Build()
	if err != nil {
		return nil, err
	}
	if err := k8sscheme.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}

// Get a cached client.
func (r *FeatureGate) getClient() (client.Client, error) {
	if cl != nil && !reflect.ValueOf(cl).IsNil() {
		return cl, nil
	}

	s, err := getScheme()
	if err != nil {
		return nil, err
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: s})
}

// SetupWebhookWithManager adds the webhook to the manager.
func (r *FeatureGate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:verbs=create;update,path=/validate-core-tanzu-vmware-com-v1alpha2-featuregate,mutating=false,failurePolicy=fail,groups=core.tanzu.vmware.com,resources=featuregates,versions=v1alpha2,name=vfeaturegate.kb.io

var _ webhook.Validator = &FeatureGate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureGate) ValidateCreate() error {
	featuregatelog.Info("validate create", "name", r.Name)

	ctx := context.Background()
	var allErrors field.ErrorList

	c, err := r.getClient()
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	allErrors = append(allErrors, r.validateFeatureExists(ctx, c)...)
	allErrors = append(allErrors, r.validateConflictingFeaturesInFeatureGate(ctx, c)...)
	allErrors = append(allErrors, r.validateFeatureForStabilityPolicyViolation(ctx, c)...)
	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind("FeatureGate").GroupKind(), r.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureGate) ValidateUpdate(old runtime.Object) error {
	featuregatelog.Info("validate update", "name", r.Name)

	oldObj, ok := old.(*FeatureGate)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected FeatureGate object, but got object of type %T", old))
	}
	if oldObj == nil {
		return nil
	}

	c, err := r.getClient()
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	ctx := context.Background()
	var allErrors field.ErrorList

	allErrors = append(allErrors, r.validateFeatureExists(ctx, c)...)
	allErrors = append(allErrors, r.validateConflictingFeaturesInFeatureGate(ctx, c)...)
	allErrors = append(allErrors, r.validateWarrantyVoidOverride(oldObj)...)
	allErrors = append(allErrors, r.validateFeatureForStabilityPolicyViolation(ctx, c)...)

	if len(allErrors) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("FeatureGate").GroupKind(), r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureGate) ValidateDelete() error {
	featuregatelog.Info("validate delete", "name", r.Name)
	return nil
}

// validateFeatureForStabilityPolicyViolation validates features for any stability policy violation in a FeatureGate
// resource
func (r *FeatureGate) validateFeatureForStabilityPolicyViolation(ctx context.Context, c client.Client) field.ErrorList {
	var allErrors field.ErrorList

	features := &FeatureList{}
	if err := c.List(ctx, features); err != nil {
		allErrors = append(allErrors, field.InternalError(field.NewPath("spec").Child("features"), err))
		return allErrors
	}

	featuresThatVoidWarranty := computeFeaturesThatVoidSupportWarranty(r.Spec, features)
	immutableFeatures := computeImmutableFeatures(r.Spec, features)

	if len(featuresThatVoidWarranty) > 0 {
		allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("features"),
			r.Spec.Features, fmt.Sprintf("cannot toggle features %v as the the stability level of the "+
				"features indicate that it should not be activated in production environments. To activate the "+
				"feature, you must agree to permanently void all support guarantees for this environment by setting "+
				"featureRef.permanentlyVoidAllSupportGuarantees to true", featuresThatVoidWarranty)))
	}

	if len(immutableFeatures) > 0 {
		allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("features"),
			r.Spec.Features, fmt.Sprintf("cannot toggle immutable features: %v", immutableFeatures)))
	}

	return allErrors
}

// computeFeaturesThatVoidSupportWarranty computes and returns features that voids the support warranty in a FeatureGate
// resource spec
func computeFeaturesThatVoidSupportWarranty(spec FeatureGateSpec, features *FeatureList) []string {
	invalidFeatures := sets.String{}
	for _, featureRef := range spec.Features {
		stabilityLevel, found := getFeatureStabilityLevel(features, featureRef.Name)
		if !found {
			// Feature doesn't exist and is validated in validateFeatureExistence method
			continue
		}
		policy := GetPolicyForStabilityLevel(stabilityLevel)
		// checks for invalid features that voids warranty of the environment when activating it, ie if the intent for
		// the feature is different from default feature state and the stability policy for the feature says it voids
		// warranty, then that feature is considered as invalid.
		if policy.VoidsWarranty && !featureRef.PermanentlyVoidAllSupportGuarantees &&
			policy.DefaultActivation != featureRef.Activate {
			invalidFeatures.Insert(featureRef.Name)
		}
	}
	return invalidFeatures.List()
}

// computeImmutableFeatures computes and returns features that are immutable in a FeatureGate resource spec
func computeImmutableFeatures(spec FeatureGateSpec, features *FeatureList) []string {
	invalidFeatures := sets.String{}
	for _, featureRef := range spec.Features {
		stabilityLevel, found := getFeatureStabilityLevel(features, featureRef.Name)
		if !found {
			// Feature doesn't exist and is validated in validateFeatureExistence method
			continue
		}
		policy := GetPolicyForStabilityLevel(stabilityLevel)
		if policy.Immutable && policy.DefaultActivation != featureRef.Activate {
			invalidFeatures.Insert(featureRef.Name)
		}
	}
	return invalidFeatures.List()
}

// validateWarrantyVoidOverride determines if permanentlyVoidAllSupportGuarantees field for a feature in FeatureGate
// resource is set to false after setting it to true initially
func (r *FeatureGate) validateWarrantyVoidOverride(oldObject *FeatureGate) field.ErrorList {
	var allErrors field.ErrorList

	invalidFeatures := computeFeaturesThatOverridedWarranyVoidOverride(r.Spec, oldObject)
	if len(invalidFeatures) > 0 {
		allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("features"),
			r.Spec.Features, fmt.Sprintf("cannot toggle features due to policy violation: %v."+
				"Set feature.skipStabilityPolicyValidation to true to override policy."+
				"Once set to true, cannot be set back to false.", invalidFeatures)))
	}
	return allErrors
}

// computeFeaturesThatOverridedWarranyVoidOverride computes and returns features that have overrided the warranty void
// override in FeatureGate resource
func computeFeaturesThatOverridedWarranyVoidOverride(spec FeatureGateSpec, oldObject *FeatureGate) []string {
	invalidFeatures := sets.String{}
	for _, featureRef := range spec.Features {
		permanentlyVoidAllSupportGuaranteesFieldFromOldObj, found :=
			getPermanentlyVoidAllSupportGuaranteesFieldForFeature(oldObject, featureRef.Name)
		if !found {
			// Feature doesn't exist in old object and need not be validated
			continue
		}
		if !featureRef.PermanentlyVoidAllSupportGuarantees && permanentlyVoidAllSupportGuaranteesFieldFromOldObj {
			invalidFeatures.Insert(featureRef.Name)
		}
	}
	return invalidFeatures.List()
}

// validateFeatureExists checks if features that are part of FeatureGate resource exist in the cluster
func (r *FeatureGate) validateFeatureExists(ctx context.Context, c client.Client) field.ErrorList {
	var allErrors field.ErrorList

	features := &FeatureList{}
	if err := c.List(ctx, features); err != nil {
		allErrors = append(allErrors, field.InternalError(field.NewPath("spec").Child("features"), err))
		return allErrors
	}

	invalidFeatures := computeFeaturesThatDoNotExist(r.Spec, features)

	if len(invalidFeatures) > 0 {
		allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("features"),
			r.Spec.Features,
			fmt.Sprintf("some features in the FeatureGate spec do not exist in cluster: %v", invalidFeatures)))
	}
	return allErrors
}

// computeFeaturesThatDoNotExist computes and returns features that do not exist in cluster and are part of FeatureGate
// resource spec
func computeFeaturesThatDoNotExist(spec FeatureGateSpec, features *FeatureList) []string {
	allFeaturesInSpec := sets.String{}
	for _, feature := range spec.Features {
		allFeaturesInSpec.Insert(feature.Name)
	}

	allFeaturesInCluster := sets.String{}
	for _, feature := range features.Items {
		allFeaturesInCluster.Insert(feature.Name)
	}
	invalidFeatures := allFeaturesInSpec.Difference(allFeaturesInCluster)
	return invalidFeatures.List()
}

// validateConflictingFeaturesInFeatureGate validates that the features in FeatureGate resource does not conflict
// with features gated by other FeatureGate resources.
func (r *FeatureGate) validateConflictingFeaturesInFeatureGate(ctx context.Context, c client.Client) field.ErrorList {
	var allErrors field.ErrorList

	featureGates := &FeatureGateList{}
	if err := c.List(ctx, featureGates); err != nil {
		allErrors = append(allErrors, field.InternalError(field.NewPath("spec").Child("features"), err))
		return allErrors
	}

	conflicts := computeConflictingFeatures(r, featureGates)

	if len(conflicts) > 0 {
		allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("features"),
			r.Spec.Features, fmt.Sprintf("features %v cannot be gated by multiple featuregates", conflicts)))
	}
	return allErrors
}

// computeConflictingFeatures computes and returns features in FeatureGate resource that conflict with features in
// other FeatureGate resources
func computeConflictingFeatures(featureGate *FeatureGate, featureGates *FeatureGateList) []string {
	allFeaturesInSpec := sets.String{}

	for _, feature := range featureGate.Spec.Features {
		allFeaturesInSpec.Insert(feature.Name)
	}

	// Gather all the gated features in the cluster
	allGatedFeaturesInCluster := sets.String{}
	for _, fg := range featureGates.Items {
		// Skip comparing the object to itself during updates.
		if fg.Name == featureGate.Name {
			continue
		}
		for _, feat := range fg.Spec.Features {
			allGatedFeaturesInCluster.Insert(feat.Name)
		}
	}

	// Intersection gives us the features that are already being gated by other FeatureGate resources
	conflicts := allGatedFeaturesInCluster.Intersection(allFeaturesInSpec)
	return conflicts.List()
}

// getPermanentlyVoidAllSupportGuaranteesFieldForFeature returns permanentlyVoidAllSupportGuarantees field's value
// for a feature from the FeatureGate resource
func getPermanentlyVoidAllSupportGuaranteesFieldForFeature(featureGate *FeatureGate, featureName string) (bool, bool) {
	for _, featureRef := range featureGate.Spec.Features {
		if featureRef.Name == featureName {
			return featureRef.PermanentlyVoidAllSupportGuarantees, true
		}
	}
	return false, false
}

// getFeatureStabilityLevel returns feature stability level for a feature from a list of Features
func getFeatureStabilityLevel(list *FeatureList, featureName string) (StabilityLevel, bool) {
	for _, feature := range list.Items {
		if featureName == feature.Name {
			return feature.Spec.Stability, true
		}
	}
	return "", false
}
