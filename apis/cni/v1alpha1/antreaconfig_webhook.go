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
var antreaconfiglog = logf.Log.WithName("antreaconfig-resource")

func (r *AntreaConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-cni-tanzu-vmware-com-v1alpha1-antreaconfig,mutating=true,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=antreaconfigs,verbs=create;update,versions=v1alpha1,name=mantreaconfig.kb.io

var _ webhook.Defaulter = &AntreaConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AntreaConfig) Default() {
	antreaconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:verbs=create;update,path=/validate-cni-tanzu-vmware-com-v1alpha1-antreaconfig,mutating=false,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=antreaconfigs,versions=v1alpha1,name=vantreaconfig.kb.io

var _ webhook.Validator = &AntreaConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AntreaConfig) ValidateCreate() error {
	antreaconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AntreaConfig) ValidateUpdate(old runtime.Object) error {
	antreaconfiglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AntreaConfig) ValidateDelete() error {
	antreaconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
