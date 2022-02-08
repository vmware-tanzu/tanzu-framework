// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-cni-tanzu-vmware-com-v1alpha1-calicoconfig,mutating=true,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=calicoconfigs,verbs=create;update,versions=v1alpha1,name=mcalicoconfig.kb.io

var _ webhook.Defaulter = &CalicoConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CalicoConfig) Default() {
	calicoconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:verbs=create;update,path=/validate-cni-tanzu-vmware-com-v1alpha1-calicoconfig,mutating=false,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=calicoconfigs,versions=v1alpha1,name=vcalicoconfig.kb.io

var _ webhook.Validator = &CalicoConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateCreate() error {
	calicoconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateUpdate(old runtime.Object) error {
	calicoconfiglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CalicoConfig) ValidateDelete() error {
	calicoconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
