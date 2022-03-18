// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var calicoconfiglog = logf.Log.WithName("calicoconfig-resource")

func (r *CalicoConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Validator = &CalicoConfig{}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cni-tanzu-vmware-com-v1alpha1-calicoconfig,mutating=false,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=calicoconfigs,versions=v1alpha1,name=vcalicoconfig.kb.io,admissionReviewVersions=v1,sideEffects=None

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateCreate() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateUpdate(old runtime.Object) error {
	calicoconfiglog.Info("validate update", "name", r.Name)

	oldObj, ok := old.(*CalicoConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("Expected an CalicoConfig but got a %T", oldObj))
	}

	var allErrs field.ErrorList

	// Check for changes to immutable fields and return errors
	if !reflect.DeepEqual(r.Spec.Namespace, oldObj.Spec.Namespace) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "namespace"),
				r.Spec.Namespace, "field is immutable"),
		)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "cni.tanzu.vmware.com", Kind: "CalicoConfig"}, r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateDelete() error {
	// No validation required for CalicoConfig deletion
	return nil
}
