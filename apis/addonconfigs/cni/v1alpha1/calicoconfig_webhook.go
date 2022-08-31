// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *CalicoConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Validator = &CalicoConfig{}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cni-tanzu-vmware-com-v1alpha1-calicoconfig,mutating=false,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=calicoconfigs,versions=v1alpha1,name=vcalicoconfig.kb.io,admissionReviewVersions=v1,sideEffects=None

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateCreate() error {
	// No validation required for CalicoConfig creation
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateUpdate(old runtime.Object) error {
	// No validation required for CalicoConfig update
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateDelete() error {
	// No validation required for CalicoConfig deletion
	return nil
}
