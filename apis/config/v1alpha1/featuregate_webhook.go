// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

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

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/util"
)

// log is for logging in this package.
var featuregatelog = logf.Log.WithName("featuregate-resource")

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

//+kubebuilder:webhook:verbs=create;update,path=/validate-config-tanzu-vmware-com-v1alpha1-featuregate,mutating=false,failurePolicy=fail,groups=config.tanzu.vmware.com,resources=featuregates,versions=v1alpha1,name=vfeaturegate.kb.io

var _ webhook.Validator = &FeatureGate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureGate) ValidateCreate() error {
	featuregatelog.Info("validate create", "name", r.Name)
	c, err := r.getClient()
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	ctx := context.Background()
	var allErrors field.ErrorList

	allErrors = append(allErrors, r.validateNamespaceConflicts(ctx, c, field.NewPath("spec"))...)

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

	allErrors = append(allErrors, r.validateNamespaceConflicts(ctx, c, field.NewPath("spec"))...)
	allErrors = append(allErrors, r.validateFeatureImmutability(ctx, c, oldObj, field.NewPath("spec").Child("features"))...)

	if len(allErrors) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("FeatureGate").GroupKind(), r.Name, allErrors)
}

// validateNamespaceConflicts validates that the namespaceSelector in spec does not conflict with namespaces gated by existing feature gates.
func (r *FeatureGate) validateNamespaceConflicts(ctx context.Context, c client.Client, fldPath *field.Path) field.ErrorList {
	var allErrors field.ErrorList

	conflicts, err := r.computeConflictingNamespaces(ctx, c)
	if err != nil {
		allErrors = append(allErrors, field.InternalError(fldPath.Child("namespaceSelector"), err))
		return allErrors
	}

	for fg, namespaces := range conflicts {
		if len(namespaces) == 0 {
			continue
		}
		allErrors = append(allErrors, field.Invalid(fldPath.Child("namespaceSelector"), r.Spec.NamespaceSelector,
			fmt.Sprintf("namespaces %v specified by namespaceSelector are already gated by FeatureGate %q", namespaces, fg)))
	}
	return allErrors
}

// computeConflictingNamespaces computes the namespaces that conflict with existing feature gates.
// This is a separate function for easier unit testing.
func (r *FeatureGate) computeConflictingNamespaces(ctx context.Context, c client.Client) (map[string][]string, error) {
	featureGates := &FeatureGateList{}
	if err := c.List(ctx, featureGates); err != nil {
		return nil, err
	}

	namespacesInSpec, err := util.NamespacesMatchingSelector(ctx, c, &r.Spec.NamespaceSelector)
	if err != nil {
		return nil, err
	}

	featureGateToNamespaces := make(map[string]sets.String)
	for i := range featureGates.Items {
		fg := featureGates.Items[i]
		// Skip comparing the object to itself during updates.
		if fg.Name == r.Name {
			continue
		}
		featureGateToNamespaces[fg.Name] = sets.NewString(fg.Status.Namespaces...)
	}

	// Output is a map of an existing feature gate resource name and the list of namespaces that conflict with the current spec.
	out := make(map[string][]string)
	set := sets.NewString(namespacesInSpec...)
	for fg, namespaces := range featureGateToNamespaces {
		intersection := namespaces.Intersection(set).List()
		if len(intersection) != 0 {
			out[fg] = intersection
		}
	}
	return out, nil
}

// validateFeatureImmutability validates that immutable features are not changed.
func (r *FeatureGate) validateFeatureImmutability(ctx context.Context, c client.Client, oldObject *FeatureGate, fldPath *field.Path) field.ErrorList {
	var allErrors field.ErrorList

	features := &FeatureList{}
	if err := c.List(ctx, features); err != nil {
		allErrors = append(allErrors, field.InternalError(fldPath, err))
		return allErrors
	}

	invalid := computeChangedImmutableFeatures(r.Spec, features.Items, oldObject)
	if len(invalid) != 0 {
		allErrors = append(allErrors, field.Invalid(fldPath, r.Spec.Features, fmt.Sprintf("cannot change immutable features: %v", invalid)))
	}
	return allErrors
}

// computeChangedImmutableFeatures returns immutable features which are changed in the current spec.
// This is a separate function for easier unit testing.
func computeChangedImmutableFeatures(spec FeatureGateSpec, currentFeatures []Feature, oldObject *FeatureGate) []string {
	immutable := sets.String{}
	for i := range currentFeatures {
		f := currentFeatures[i]
		if f.Spec.Discoverable && f.Spec.Immutable {
			immutable.Insert(f.Name)
		}
	}

	oldActivated := sets.NewString(oldObject.Status.ActivatedFeatures...)
	oldDeactivated := sets.NewString(oldObject.Status.DeactivatedFeatures...)

	changedFeatures := sets.String{}
	for _, featureRef := range spec.Features {
		name := featureRef.Name
		if !immutable.Has(name) {
			continue
		}
		// Features that changed states from activated to deactivated or vice versa in the current update.
		if (featureRef.Activate && oldDeactivated.Has(name)) || !featureRef.Activate && oldActivated.Has(name) {
			changedFeatures.Insert(name)
		}
	}
	return immutable.Intersection(changedFeatures).List()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureGate) ValidateDelete() error {
	featuregatelog.Info("validate delete", "name", r.Name)
	return nil
}
